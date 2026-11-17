package vp

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PlateHandler receives an uploaded plate from FTPS STOR.
type PlateHandler func(ctx context.Context, filename string, body []byte) error

// StartFTPS runs an implicit-TLS FTP server (LAN-mode shape: user bblp, pass = access code).
func StartFTPS(ctx context.Context, cfg Config, tlsCfg *tls.Config, onPlate PlateHandler, log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}
	ln, err := tls.Listen("tcp", cfg.FTPSAddr, tlsCfg)
	if err != nil {
		log.Warn("vp ftps listen failed", "addr", cfg.FTPSAddr, "err", err)
		return
	}
	log.Info("vp ftps listening", "addr", cfg.FTPSAddr, "user", "bblp")
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()
	var pasvMu sync.Mutex
	nextPasv := cfg.PASVStart
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Warn("vp ftps accept", "err", err)
				return
			}
		}
		go handleFTPS(ctx, conn, cfg, tlsCfg, onPlate, log, &pasvMu, &nextPasv)
	}
}

type ftpsSession struct {
	cfg     Config
	tlsCfg  *tls.Config
	onPlate PlateHandler
	log     *slog.Logger
	user    string
	authed  bool
	cwd     string
	pasvLn  net.Listener
	pasvMu  *sync.Mutex
	next    *int
}

func handleFTPS(ctx context.Context, conn net.Conn, cfg Config, tlsCfg *tls.Config, onPlate PlateHandler, log *slog.Logger, pasvMu *sync.Mutex, next *int) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Minute))
	s := &ftpsSession{cfg: cfg, tlsCfg: tlsCfg, onPlate: onPlate, log: log, cwd: "/", pasvMu: pasvMu, next: next}
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	s.reply(rw, 220, "BambuFarmer studio-bridge ready")
	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		cmd, arg, _ := strings.Cut(line, " ")
		cmd = strings.ToUpper(strings.TrimSpace(cmd))
		arg = strings.TrimSpace(arg)
		switch cmd {
		case "USER":
			s.user = arg
			s.reply(rw, 331, "password required")
		case "PASS":
			if s.user == "bblp" && arg == cfg.AccessCode {
				s.authed = true
				s.reply(rw, 230, "logged in")
			} else {
				s.reply(rw, 530, "login incorrect")
			}
		case "QUIT":
			s.reply(rw, 221, "bye")
			return
		case "SYST":
			s.reply(rw, 215, "UNIX Type: L8")
		case "FEAT":
			s.replyMultiline(rw, 211, "FEAT", "PASV\nEPSV\nSIZE\nUTF8")
		case "TYPE":
			s.reply(rw, 200, "type set")
		case "PWD", "XPWD":
			s.reply(rw, 257, `"`+s.cwd+`"`)
		case "CWD", "XCWD":
			if !s.authed {
				s.reply(rw, 530, "login first")
				continue
			}
			s.cwd = "/" + strings.Trim(arg, "/")
			if s.cwd == "/" || arg == "" {
				s.cwd = "/"
			}
			s.reply(rw, 250, "cwd ok")
		case "PASV":
			if !s.authed {
				s.reply(rw, 530, "login first")
				continue
			}
			host, port, err := s.openPASV(conn)
			if err != nil {
				s.reply(rw, 425, "pasv failed")
				continue
			}
			h := strings.ReplaceAll(host, ".", ",")
			p1, p2 := port/256, port%256
			s.reply(rw, 227, fmt.Sprintf("Entering Passive Mode (%s,%d,%d)", h, p1, p2))
		case "EPSV":
			if !s.authed {
				s.reply(rw, 530, "login first")
				continue
			}
			_, port, err := s.openPASV(conn)
			if err != nil {
				s.reply(rw, 425, "epsv failed")
				continue
			}
			s.reply(rw, 229, fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", port))
		case "LIST", "NLST":
			if !s.authed {
				s.reply(rw, 530, "login first")
				continue
			}
			data, err := s.acceptData()
			if err != nil {
				s.reply(rw, 425, "no data connection")
				continue
			}
			s.reply(rw, 150, "here comes the directory listing")
			_, _ = data.Write([]byte(""))
			_ = data.Close()
			s.reply(rw, 226, "directory send OK")
		case "SIZE":
			s.reply(rw, 550, "file not found")
		case "STOR":
			if !s.authed {
				s.reply(rw, 530, "login first")
				continue
			}
			data, err := s.acceptData()
			if err != nil {
				s.reply(rw, 425, "no data connection")
				continue
			}
			s.reply(rw, 150, "ok to send data")
			body, err := io.ReadAll(io.LimitReader(data, 64<<20))
			_ = data.Close()
			if err != nil {
				s.reply(rw, 426, "transfer aborted")
				continue
			}
			name := filepath.Base(arg)
			if name == "" || name == "." {
				name = "plate.gcode.3mf"
			}
			if s.onPlate != nil {
				if err := s.onPlate(ctx, name, body); err != nil {
					s.log.Warn("vp ftps stor ingest", "err", err, "file", name)
					s.reply(rw, 451, "local error")
					continue
				}
			}
			s.log.Info("vp ftps stor ok", "file", name, "bytes", len(body))
			s.reply(rw, 226, "transfer complete")
		case "NOOP":
			s.reply(rw, 200, "ok")
		case "OPTS":
			s.reply(rw, 200, "ok")
		default:
			s.reply(rw, 502, "command not implemented")
		}
	}
}

func (s *ftpsSession) reply(rw *bufio.ReadWriter, code int, msg string) {
	_, _ = fmt.Fprintf(rw, "%d %s\r\n", code, msg)
	_ = rw.Flush()
}

func (s *ftpsSession) replyMultiline(rw *bufio.ReadWriter, code int, label, body string) {
	_, _ = fmt.Fprintf(rw, "%d-%s\r\n", code, label)
	for _, line := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		_, _ = fmt.Fprintf(rw, " %s\r\n", line)
	}
	_, _ = fmt.Fprintf(rw, "%d End\r\n", code)
	_ = rw.Flush()
}

func (s *ftpsSession) openPASV(ctrl net.Conn) (host string, port int, err error) {
	if s.pasvLn != nil {
		_ = s.pasvLn.Close()
		s.pasvLn = nil
	}
	host = pasvHost(ctrl)
	s.pasvMu.Lock()
	port = *s.next
	if port < s.cfg.PASVStart || port > s.cfg.PASVEnd {
		port = s.cfg.PASVStart
	}
	*s.next = port + 1
	if *s.next > s.cfg.PASVEnd {
		*s.next = s.cfg.PASVStart
	}
	s.pasvMu.Unlock()
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return "", 0, err
	}
	s.pasvLn = ln
	return host, port, nil
}

func (s *ftpsSession) acceptData() (net.Conn, error) {
	if s.pasvLn == nil {
		return nil, fmt.Errorf("no pasv")
	}
	_ = s.pasvLn.(*net.TCPListener).SetDeadline(time.Now().Add(30 * time.Second))
	c, err := s.pasvLn.Accept()
	_ = s.pasvLn.Close()
	s.pasvLn = nil
	if err != nil {
		return nil, err
	}
	// Optional wrap with TLS for some clients; many Studio paths use clear data after PASV on LAN FTPS.
	return c, nil
}

func pasvHost(ctrl net.Conn) string {
	a := ctrl.LocalAddr().String()
	host, _, err := net.SplitHostPort(a)
	if err != nil || host == "" || host == "::" || host == "0.0.0.0" {
		return "127.0.0.1"
	}
	if strings.Contains(host, ":") {
		// IPv6 PASV is awkward; prefer EPSV clients. Fall back to loopback for tests.
		return "127.0.0.1"
	}
	return host
}
