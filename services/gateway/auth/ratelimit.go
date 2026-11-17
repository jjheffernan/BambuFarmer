package auth

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/httpserver"
)

// LoginRateLimit caps POST /api/v1/auth/login attempts per client IP.
func LoginRateLimit(maxPerMinute int, next http.Handler) http.Handler {
	if maxPerMinute <= 0 {
		maxPerMinute = 20
	}
	lim := newWindowLimiter(maxPerMinute, time.Minute)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/auth/login" {
			ip := clientIP(r)
			if !lim.allow(ip) {
				httpserver.WriteError(w, api.New(http.StatusTooManyRequests, "rate_limited", "too many login attempts — try again shortly"))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

type windowLimiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	buckets map[string]*windowBucket
}

type windowBucket struct {
	reset time.Time
	count int
}

func newWindowLimiter(max int, window time.Duration) *windowLimiter {
	return &windowLimiter{
		max:     max,
		window:  window,
		buckets: map[string]*windowBucket{},
	}
}

func (l *windowLimiter) allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[key]
	if !ok || now.After(b.reset) {
		l.buckets[key] = &windowBucket{reset: now.Add(l.window), count: 1}
		return true
	}
	if b.count >= l.max {
		return false
	}
	b.count++
	return true
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
