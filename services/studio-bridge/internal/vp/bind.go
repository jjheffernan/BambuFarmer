package vp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"strings"
	"time"
)

// StartBind listens on bind/detect ports (typically :3000 and :3002).
// Studio versions differ; we accept TCP and reply with a JSON device identity when possible.
func StartBind(ctx context.Context, cfg Config, log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}
	for _, addr := range cfg.BindAddrs {
		addr := addr
		go serveBind(ctx, addr, cfg, log)
	}
}

func serveBind(ctx context.Context, addr string, cfg Config, log *slog.Logger) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Warn("vp bind listen failed", "addr", addr, "err", err)
		return
	}
	log.Info("vp bind listening", "addr", addr, "sn", cfg.Serial)
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Warn("vp bind accept", "err", err)
				return
			}
		}
		go handleBindConn(conn, cfg, log)
	}
}

func handleBindConn(conn net.Conn, cfg Config, log *slog.Logger) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
	reply := map[string]any{
		"dev_id":    cfg.Serial,
		"dev_name":  cfg.Name,
		"dev_model": cfg.Model,
		"name":      cfg.Name,
		"model":     cfg.Model,
		"ok":        true,
	}
	raw, _ := json.Marshal(reply)

	br := bufio.NewReader(conn)
	buf := make([]byte, 0, 512)
	tmp := make([]byte, 256)
	for {
		n, err := br.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "timeout") {
				log.Debug("vp bind read", "err", err)
			}
			break
		}
		if n < len(tmp) {
			break
		}
		if len(buf) > 8192 {
			break
		}
	}
	if len(buf) > 0 {
		log.Debug("vp bind request", "bytes", len(buf), "sn", cfg.Serial)
	}
	_, _ = conn.Write(append(raw, '\n'))
}
