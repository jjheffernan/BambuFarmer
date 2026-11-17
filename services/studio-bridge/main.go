// Command studio-bridge is the optional Studio/Orca receive plugin (compose profile: studio-bridge).
// Receive: HTTP/drop + optional LAN VP (SSDP/bind/MQTT/FTPS) → core artifact/job → fm_native.
// Proxy: transparent TCP relay to a LAN/direct printer (not FM-bound).
package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/services/studio-bridge/internal/bridge"
	"github.com/jjheffernan/BambuFarmer/services/studio-bridge/internal/proxy"
	"github.com/jjheffernan/BambuFarmer/services/studio-bridge/internal/vp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service := "studio-bridge"
	coreURL := os.Getenv("CORE_URL")
	siteID := getenv("SITE_ID", "default")
	mode := strings.ToLower(getenv("STUDIO_BRIDGE_MODE", "receive"))
	policy := bridge.ParsePolicy(getenv("STUDIO_BRIDGE_POLICY", "queue"))
	dropDir := strings.TrimSpace(os.Getenv("STUDIO_BRIDGE_DROP_DIR"))
	defaultModel := strings.TrimSpace(os.Getenv("STUDIO_BRIDGE_TARGET_MODEL"))
	redisURL := os.Getenv("REDIS_URL")
	vpcfg := vp.LoadConfig()
	pxcfg := proxy.Load()

	var bus *events.Bus
	if redisURL != "" {
		b, err := events.Dial(ctx, redisURL)
		if err != nil {
			slog.Warn("redis dial failed; studio-bridge events disabled", "err", err)
		} else {
			bus = b
			defer bus.Close()
		}
	}

	ingestPlate := func(c context.Context, filename string, body []byte, explicitModel string) error {
		hints := vp.HintsFrom3MF(body)
		model := vp.MergeHints(explicitModel, hints.TargetModel, defaultModel)
		res, err := bridge.Ingest(c, coreURL, policy, bridge.Request{
			Filename:    filename,
			Body:        body,
			Name:        strings.TrimSuffix(strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)), ".gcode"),
			TargetModel: model,
		}, nil)
		if err != nil {
			return err
		}
		if len(hints.Filaments) > 0 {
			slog.Info("plate hints", "filaments", hints.Filaments, "target_model", model)
		}
		emit(c, bus, siteID, res)
		return nil
	}

	if mode != "proxy" && dropDir != "" {
		go bridge.WatchDropDir(ctx, bridge.WatchOptions{
			Dir:      dropDir,
			CoreURL:  coreURL,
			Policy:   policy,
			Interval: 5 * time.Second,
			Log:      slog.Default(),
			OnFile: func(c context.Context, name string, body []byte) error {
				return ingestPlate(c, name, body, "")
			},
		})
	}

	var certFP string
	if mode == "proxy" {
		proxy.Start(ctx, pxcfg, slog.Default())
		if pxcfg.Ready() {
			pcfg := vp.Config{Serial: pxcfg.Serial, Name: pxcfg.Name, Model: pxcfg.Model}
			go vp.StartSSDP(ctx, pcfg, vp.AdvertiseIP(), slog.Default())
		}
	} else if vpcfg.Enabled {
		tlsCfg, err := vp.LoadOrCreateTLS(vpcfg.CertDir)
		if err != nil {
			slog.Warn("vp tls cert", "err", err)
		} else {
			if pem, err := vp.ReadCertPEM(vpcfg.CertDir); err == nil {
				certFP, _ = vp.CertFingerprintSHA256(pem)
			}
			adv := vp.AdvertiseIP()
			vp.StartBind(ctx, vpcfg, slog.Default())
			go vp.StartSSDP(ctx, vpcfg, adv, slog.Default())
			go vp.StartFTPS(ctx, vpcfg, tlsCfg, func(c context.Context, filename string, body []byte) error {
				return ingestPlate(c, filename, body, "")
			}, slog.Default())
			go vp.StartMQTT(ctx, vpcfg, tlsCfg, slog.Default())
		}
	}

	mux := httpserver.NewMux(service)
	mux.HandleFunc("GET /api/v1/studio-bridge/status", func(w http.ResponseWriter, _ *http.Request) {
		st := map[string]any{
			"service":       service,
			"status":        "ok",
			"mode":          mode,
			"policy":        string(policy),
			"core_url":      strings.TrimSpace(coreURL) != "",
			"events":        bus != nil,
			"drop_dir":      dropDir,
			"target_model":  defaultModel,
			"proxy_ready":   pxcfg.Ready(),
			"proxy":         pxcfg.StatusMap(),
			"vp":            vpcfg.StatusMap(),
			"ca_endpoint":   "/api/v1/studio-bridge/ca.pem",
			"ca_sha256":     certFP,
			"advertise_ip":  vp.AdvertiseIP(),
			"note":          statusNote(mode, vpcfg.Enabled, pxcfg.Ready()),
		}
		writeJSON(w, http.StatusOK, st)
	})
	mux.HandleFunc("GET /api/v1/studio-bridge/ca.pem", func(w http.ResponseWriter, _ *http.Request) {
		pem, err := vp.ReadCertPEM(vpcfg.CertDir)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "CA not generated yet — enable STUDIO_BRIDGE_VP=1 once"})
			return
		}
		w.Header().Set("Content-Type", "application/x-pem-file")
		w.Header().Set("Content-Disposition", `attachment; filename="bambufarmer-studio-bridge.pem"`)
		_, _ = w.Write(pem)
	})
	mux.HandleFunc("POST /api/v1/studio-bridge/ingest", func(w http.ResponseWriter, r *http.Request) {
		if mode == "proxy" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "ingest disabled in proxy mode — traffic goes to STUDIO_BRIDGE_PROXY_TARGET",
			})
			return
		}
		req, err := parseIngest(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		hints := vp.HintsFrom3MF(req.Body)
		req.TargetModel = vp.MergeHints(req.TargetModel, hints.TargetModel, defaultModel)
		res, err := bridge.Ingest(r.Context(), coreURL, policy, req, nil)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		out := map[string]any{
			"received":        res.Received,
			"artifact_id":     res.ArtifactID,
			"job_id":          res.JobID,
			"create_job_path": res.CreateJobPath,
			"reason":          res.Reason,
			"hints":           hints,
			"target_model":    req.TargetModel,
		}
		emit(r.Context(), bus, siteID, res)
		writeJSON(w, http.StatusOK, out)
	})

	if err := httpserver.ListenAndServe(httpserver.Options{
		Service: service,
		Addr:    getenv("HTTP_ADDR", ":8080"),
		Mux:     mux,
	}); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func statusNote(mode string, vpOn, proxyReady bool) string {
	if mode == "proxy" {
		if proxyReady {
			return "Proxy mode: transparent TCP to STUDIO_BRIDGE_PROXY_TARGET (LAN/direct_printer only). Not for FM-bound printers."
		}
		return "Proxy mode needs STUDIO_BRIDGE_PROXY_TARGET=printer-ip."
	}
	if vpOn {
		return "Receive + VP: SSDP :2021, bind, FTPS, MQTT. Download GET /api/v1/studio-bridge/ca.pem for slicer trust. STOR/HTTP → core → fm_native."
	}
	return "Receive: POST ingest or drop folder. Set STUDIO_BRIDGE_VP=1 for LAN VP; STUDIO_BRIDGE_MODE=proxy for LAN printer relay."
}

func parseIngest(r *http.Request) (bridge.Request, error) {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/") {
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			return bridge.Request{}, err
		}
		file, hdr, err := r.FormFile("file")
		if err != nil {
			return bridge.Request{}, err
		}
		defer file.Close()
		body, err := io.ReadAll(io.LimitReader(file, 64<<20))
		if err != nil {
			return bridge.Request{}, err
		}
		req := bridge.Request{
			Name:        r.FormValue("name"),
			Filename:    filepath.Base(hdr.Filename),
			Body:        body,
			TargetModel: r.FormValue("target_model"),
		}
		if q := strings.TrimSpace(r.FormValue("queue")); q != "" {
			v := q == "1" || strings.EqualFold(q, "true") || strings.EqualFold(q, "yes")
			req.Queue = &v
		}
		return req, nil
	}
	return bridge.Request{}, errMultipart
}

type multipartErr string

func (e multipartErr) Error() string { return string(e) }

const errMultipart multipartErr = "multipart form with file field required"

func emit(ctx context.Context, bus *events.Bus, siteID string, res bridge.Result) {
	if bus == nil || !res.Received {
		return
	}
	if siteID == "" {
		siteID = "default"
	}
	payload := map[string]any{
		"artifact_id":     res.ArtifactID,
		"job_id":          res.JobID,
		"create_job_path": res.CreateJobPath,
		"reason":          res.Reason,
	}
	typ := events.TypeStudioBridgeReceived
	if res.JobID != "" {
		typ = events.TypeStudioBridgeQueued
	}
	_, _ = bus.Publish(ctx, events.StreamStudioBridge, events.NewEnvelope(typ, "studio-bridge", siteID, payload))
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
