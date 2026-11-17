package vp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

// StartMQTT runs a TLS MQTT broker that authenticates bblp/access-code and
// publishes a minimal idle push_status on device/{sn}/report.
func StartMQTT(ctx context.Context, cfg Config, tlsCfg *tls.Config, log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}
	server := mqtt.New(&mqtt.Options{InlineClient: true})
	_ = server.AddHook(new(auth.Hook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{{
				Username: auth.RString("bblp"),
				Password: auth.RString(cfg.AccessCode),
				Allow:    true,
			}},
			ACL: auth.ACLRules{{
				Username: auth.RString("bblp"),
				Filters:  auth.Filters{"#": auth.ReadWrite},
			}},
		},
	})
	tcp := listeners.NewTCP(listeners.Config{
		ID:        "vp-mqtt",
		Address:   cfg.MQTTAddr,
		TLSConfig: tlsCfg,
	})
	if err := server.AddListener(tcp); err != nil {
		log.Warn("vp mqtt listener", "err", err)
		return
	}
	go func() {
		if err := server.Serve(); err != nil {
			log.Warn("vp mqtt serve", "err", err)
		}
	}()
	log.Info("vp mqtt listening", "addr", cfg.MQTTAddr, "sn", cfg.Serial)

	topic := "device/" + cfg.Serial + "/report"
	var mu sync.Mutex
	publishIdle := func() {
		mu.Lock()
		defer mu.Unlock()
		payload, _ := json.Marshal(map[string]any{
			"print": map[string]any{
				"gcode_state":       "IDLE",
				"mc_percent":        0,
				"mc_remaining_time": 0,
				"wifi_signal":       "-40dBm",
				"nozzle_diameter":   "0.4",
				"subtask_name":      "",
			},
		})
		_ = server.Publish(topic, payload, false, 0)
	}
	publishIdle()

	// Nudge idle status so Studio sees a live printer after connect.
	t := time.NewTicker(10 * time.Second)
	go func() {
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				_ = server.Close()
				return
			case <-t.C:
				publishIdle()
			}
		}
	}()

	// Log print commands for operators (receive path still relies on FTPS STOR).
	_ = server.AddHook(new(cmdHook), &cmdHookOptions{Log: log, SN: cfg.Serial})
}

type cmdHookOptions struct {
	Log *slog.Logger
	SN  string
}

type cmdHook struct {
	mqtt.HookBase
	log *slog.Logger
	sn  string
}

func (h *cmdHook) ID() string { return "vp-cmd-log" }

func (h *cmdHook) Provides(b byte) bool {
	return b == mqtt.OnPublish
}

func (h *cmdHook) Init(config any) error {
	if o, ok := config.(*cmdHookOptions); ok && o != nil {
		h.log = o.Log
		h.sn = o.SN
	}
	if h.log == nil {
		h.log = slog.Default()
	}
	return nil
}

func (h *cmdHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	if h.sn != "" && (pk.TopicName == "device/"+h.sn+"/request" || pk.TopicName == "device/"+h.sn+"/print") {
		h.log.Info("vp mqtt command", "topic", pk.TopicName, "bytes", len(pk.Payload))
	}
	return pk, nil
}
