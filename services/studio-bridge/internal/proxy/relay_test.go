package proxy

import (
	"io"
	"log/slog"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestConfigReady(t *testing.T) {
	cfg := Config{Enabled: true, TargetIP: "1.2.3.4", Ports: []int{8883}}
	if !cfg.Ready() {
		t.Fatal("ready")
	}
	bad := Config{Enabled: true, Ports: []int{8883}}
	if bad.Ready() {
		t.Fatal("missing target")
	}
}

func TestPipeCopiesBytes(t *testing.T) {
	backend, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	_, bport, _ := net.SplitHostPort(backend.Addr().String())
	backendPort, _ := strconv.Atoi(bport)

	go func() {
		c, err := backend.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		buf := make([]byte, 4)
		if _, err := io.ReadFull(c, buf); err != nil {
			return
		}
		_, _ = c.Write([]byte("pong"))
	}()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	proxyAddr := ln.Addr().String()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go pipe(c, "127.0.0.1", backendPort, log)
		}
	}()
	time.Sleep(20 * time.Millisecond)

	client, err := net.DialTimeout("tcp", proxyAddr, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err := client.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 4)
	_ = client.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := io.ReadFull(client, buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != "pong" {
		t.Fatalf("%q", buf)
	}
}
