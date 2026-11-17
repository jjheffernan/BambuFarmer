// Command studio-bridge is the optional Studio/Orca receive plugin (compose profile: studio-bridge).
// Receive: HTTP ingest + drop-folder + optional LAN VP (bind/MQTT/FTPS) → core artifact/job → fm_native.
// Proxy mode is deferred (LAN printers only).
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
	redisURL := os.Getenv("REDIS_URL")
	vpcfg := vp.LoadConfig()

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

	ingestPlate := func(c context.Context, filename string, body []byte) error {
		res, err := bridge.Ingest(c, coreURL, policy, bridge.Request{
			Filename: filename,
			Body:     body,
			Name:     strings.TrimSuffix(strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)), ".gcode"),
		}, nil)
		if err != nil {
			return err
		}
		emit(c, bus, siteID, res)
		return nil
	}

	if dropDir != "" {
		go bridge.WatchDropDir(ctx, bridge.WatchOptions{
			Dir:      dropDir,
			CoreURL:  coreURL,
			Policy:   policy,
			Interval: 5 * time.Second,
			Log:      slog.Default(),
		})
	}

	if vpcfg.Enabled && mode != "proxy" {
		tlsCfg, err := vp.LoadOrCreateTLS(vpcfg.CertDir)
		if err != nil {
			slog.Warn("vp tls cert", "err", err)
		} else {
			vp.StartBind(ctx, vpcfg, slog.Default())
			go vp.StartFTPS(ctx, vpcfg, tlsCfg, ingestPlate, slog.Default())
			go vp.StartMQTT(ctx, vpcfg, tlsCfg, slog.Default())
		}
	}

	mux := httpserver.NewMux(service)
	mux.HandleFunc("GET /api/v1/studio-bridge/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"service":     service,
			"status":      "ok",
			"mode":        mode,
			"policy":      string(policy),
			"core_url":    strings.TrimSpace(coreURL) != "",
			"events":      bus != nil,
			"drop_dir":    dropDir,
			"proxy_ready": false,
			"vp":          vpcfg.StatusMap(),
			"note":        statusNote(mode, vpcfg.Enabled),
		})
	})
	mux.HandleFunc("POST /api/v1/studio-bridge/ingest", func(w http.ResponseWriter, r *http.Request) {
		if mode == "proxy" {
			writeJSON(w, http.StatusNotImplemented, map[string]string{
				"error": "proxy mode deferred — use receive (STUDIO_BRIDGE_MODE=receive)",
			})
			return
		}
		req, err := parseIngest(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		res, err := bridge.Ingest(r.Context(), coreURL, policy, req, nil)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		emit(r.Context(), bus, siteID, res)
		writeJSON(w, http.StatusOK, res)
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

func statusNote(mode string, vpOn bool) string {
	if mode == "proxy" {
		return "Proxy mode is deferred (LAN/direct_printer only). Set STUDIO_BRIDGE_MODE=receive."
	}
	if vpOn {
		return "Receive + VP: bind :3000/:3002, FTPS :990 (bblp + access code), MQTT :8883 idle report. STOR → core artifact/job → fm_native. Add printer in Studio by IP/access code; slicer CA trust may be required."
	}
	return "Receive: POST /api/v1/studio-bridge/ingest or drop folder. Set STUDIO_BRIDGE_VP=1 for LAN virtual printer (bind/FTPS/MQTT)."
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
