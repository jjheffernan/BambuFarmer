package proxy

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"time"
)

// Start launches transparent TCP relays for configured ports (+ PASV range).
func Start(ctx context.Context, cfg Config, log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}
	if !cfg.Ready() {
		log.Warn("proxy not ready — set STUDIO_BRIDGE_MODE=proxy and STUDIO_BRIDGE_PROXY_TARGET")
		return
	}
	ports := append([]int{}, cfg.Ports...)
	for p := cfg.PASVStart; p <= cfg.PASVEnd && p-cfg.PASVStart < 101; p++ {
		ports = append(ports, p)
	}
	seen := map[int]bool{}
	for _, port := range ports {
		if seen[port] {
			continue
		}
		seen[port] = true
		go listenRelay(ctx, cfg, port, log)
	}
	log.Info("proxy started", "target", cfg.TargetIP, "ports", len(seen))
}

func listenRelay(ctx context.Context, cfg Config, port int, log *slog.Logger) {
	host := cfg.ListenHost
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Warn("proxy listen failed", "addr", addr, "err", err)
		return
	}
	log.Info("proxy listening", "local", addr, "remote", net.JoinHostPort(cfg.TargetIP, strconv.Itoa(port)))
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
				log.Warn("proxy accept", "port", port, "err", err)
				return
			}
		}
		go pipe(conn, cfg.TargetIP, port, log)
	}
}

func pipe(client net.Conn, targetIP string, port int, log *slog.Logger) {
	defer client.Close()
	_ = client.SetDeadline(time.Now().Add(30 * time.Minute))
	remoteAddr := net.JoinHostPort(targetIP, strconv.Itoa(port))
	remote, err := net.DialTimeout("tcp", remoteAddr, 10*time.Second)
	if err != nil {
		log.Warn("proxy dial", "remote", remoteAddr, "err", err)
		return
	}
	defer remote.Close()
	_ = remote.SetDeadline(time.Now().Add(30 * time.Minute))

	var wg sync.WaitGroup
	wg.Add(2)
	copyDir := func(dst, src net.Conn) {
		defer wg.Done()
		_, _ = io.Copy(dst, src)
		_ = closeWrite(dst)
	}
	go copyDir(remote, client)
	go copyDir(client, remote)
	wg.Wait()
}

type closeWriter interface {
	CloseWrite() error
}

func closeWrite(c net.Conn) error {
	if cw, ok := c.(closeWriter); ok {
		return cw.CloseWrite()
	}
	return fmt.Errorf("no close write")
}
