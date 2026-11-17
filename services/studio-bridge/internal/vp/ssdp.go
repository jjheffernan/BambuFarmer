package vp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"
)

const ssdpGroup = "239.255.255.250:1900"

// StartSSDP answers M-SEARCH on UDP 2021 (and best-effort 1900) and periodically
// NOTIFYs. Same-LAN discovery only; Docker bridge often needs host networking.
func StartSSDP(ctx context.Context, cfg Config, advertiseIP string, log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}
	addr := getenv("STUDIO_BRIDGE_SSDP_ADDR", ":2021")
	pc, err := net.ListenPacket("udp4", addr)
	if err != nil {
		log.Warn("vp ssdp listen failed", "addr", addr, "err", err)
		return
	}
	log.Info("vp ssdp listening", "addr", addr, "sn", cfg.Serial, "advertise_ip", advertiseIP)
	go func() {
		<-ctx.Done()
		_ = pc.Close()
	}()

	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for {
			sendNotify(cfg, advertiseIP, log)
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()

	buf := make([]byte, 4096)
	for {
		n, raddr, err := pc.ReadFrom(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Warn("vp ssdp read", "err", err)
				return
			}
		}
		msg := string(buf[:n])
		if !strings.Contains(strings.ToUpper(msg), "M-SEARCH") {
			continue
		}
		resp := ssdpResponse(cfg, advertiseIP)
		if _, err := pc.WriteTo([]byte(resp), raddr); err != nil {
			log.Debug("vp ssdp reply", "err", err)
		}
	}
}

func sendNotify(cfg Config, advertiseIP string, log *slog.Logger) {
	c, err := net.Dial("udp4", ssdpGroup)
	if err != nil {
		return
	}
	defer c.Close()
	body := ssdpNotify(cfg, advertiseIP)
	if _, err := c.Write([]byte(body)); err != nil {
		log.Debug("vp ssdp notify", "err", err)
	}
}

func ssdpResponse(cfg Config, ip string) string {
	if ip == "" {
		ip = "127.0.0.1"
	}
	return fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Server: Buildroot/bambu 1.0\r\n"+
			"Location: //%s:0\r\n"+
			"USN: %s\r\n"+
			"Cache-Control: max-age=1800\r\n"+
			"DevModel.bambu.com: %s\r\n"+
			"DevName.bambu.com: %s\r\n"+
			"DevSignal.bambu.com: -40\r\n"+
			"DevConnect.bambu.com: lan\r\n"+
			"DevBind.bambu.com: free\r\n"+
			"Devseclink.bambu.com: secure\r\n"+
			"DevVersion.bambu.com: 01.00.00.00\r\n"+
			"\r\n",
		ip, cfg.Serial, cfg.Model, cfg.Name,
	)
}

func ssdpNotify(cfg Config, ip string) string {
	if ip == "" {
		ip = "127.0.0.1"
	}
	return fmt.Sprintf(
		"NOTIFY * HTTP/1.1\r\n"+
			"HOST: 239.255.255.250:1900\r\n"+
			"Server: Buildroot/bambu 1.0\r\n"+
			"Location: //%s:0\r\n"+
			"NTS: ssdp:alive\r\n"+
			"USN: %s\r\n"+
			"Cache-Control: max-age=1800\r\n"+
			"DevModel.bambu.com: %s\r\n"+
			"DevName.bambu.com: %s\r\n"+
			"DevSignal.bambu.com: -40\r\n"+
			"DevConnect.bambu.com: lan\r\n"+
			"DevBind.bambu.com: free\r\n"+
			"Devseclink.bambu.com: secure\r\n"+
			"\r\n",
		ip, cfg.Serial, cfg.Model, cfg.Name,
	)
}

// AdvertiseIP picks a non-loopback IPv4 for SSDP Location, or override env.
func AdvertiseIP() string {
	if v := strings.TrimSpace(osGetenv("STUDIO_BRIDGE_ADVERTISE_IP")); v != "" {
		return v
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			return ip.String()
		}
	}
	return "127.0.0.1"
}

func osGetenv(k string) string { return getenv(k, "") }
