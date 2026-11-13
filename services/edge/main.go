package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/agent"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/bambu"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/config"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/mqtt"
)

func main() {
	cfg := config.Load()
	log := slog.Default()

	if !cfg.Configured() && cfg.FixtureDir == "" {
		for _, cand := range []string{"fixtures/fm", "/fixtures/fm", "../../fixtures/fm"} {
			if st, err := os.Stat(cand); err == nil && st.IsDir() {
				cfg.FixtureDir = cand
				log.Info("using fixture dir", "dir", cand)
				break
			}
		}
	}
	if cfg.FixtureDir != "" {
		if abs, err := filepath.Abs(cfg.FixtureDir); err == nil {
			cfg.FixtureDir = abs
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var busClient *events.Bus
	if cfg.RedisURL != "" {
		var err error
		busClient, err = events.Dial(ctx, cfg.RedisURL)
		if err != nil {
			log.Warn("redis dial failed; deltas will not publish", "err", err)
		} else {
			defer busClient.Close()
		}
	}

	plane, err := bambu.Open(cfg, log)
	if err != nil {
		slog.Error("farm control plane", "err", err)
		os.Exit(1)
	}
	defer plane.Close()

	var reports farm.EventSource
	if src, err := mqtt.Open(cfg.MQTTURL, cfg.CertDir, cfg.TLSInsecure); err == nil {
		reports = src
		log.Info("mqtt reports enabled", "url", cfg.MQTTURL)
	} else if !errors.Is(err, farm.ErrNotConfigured) {
		log.Warn("mqtt reports disabled", "err", err)
	}

	ag := agent.New(agent.Options{Config: cfg, Plane: plane, Bus: busClient, Reports: reports, Log: log})
	if err := ag.Connect(ctx); err != nil {
		slog.Warn("initial farm connect failed; will retry on poll", "kind", plane.Kind(), "err", err)
	}
	ag.Start(ctx)

	mux := httpserver.NewMux("edge")
	mux.HandleFunc("GET /api/v1/edge/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ag.Status())
	})
	mux.HandleFunc("POST /api/v1/edge/control/{sn}", func(w http.ResponseWriter, r *http.Request) {
		sn := r.PathValue("sn")
		var body struct {
			Opt api.ControlOpt `json:"opt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpserver.WriteError(w, api.BadRequest("invalid JSON body"))
			return
		}
		if err := ag.Control(r.Context(), sn, body.Opt); err != nil {
			if _, ok := api.AsError(err); ok {
				httpserver.WriteError(w, err)
				return
			}
			httpserver.WriteError(w, api.Wrap(http.StatusBadGateway, api.CodeBadGateway, "control failed", err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "sn": sn, "opt": string(body.Opt)})
	})

	if err := httpserver.ListenAndServe(httpserver.Options{
		Service: "edge",
		Addr:    cfg.HTTPAddr,
		Mux:     mux,
	}); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
