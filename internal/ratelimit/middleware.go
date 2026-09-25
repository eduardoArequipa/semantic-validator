package ratelimit

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type bucket struct {
	started time.Time
	count   int
}

// Limiter enforces a per-client fixed-window request limit in process memory.
type Limiter struct {
	mu                sync.Mutex
	requests          int
	window            time.Duration
	clients           map[string]bucket
	lastCleanup       time.Time
	trustProxyHeaders bool
}

func New(requests int, window time.Duration) *Limiter {
	return NewWithTrustedProxy(requests, window, false)
}

// NewWithTrustedProxy enables X-Forwarded-For only when the application is
// reachable exclusively through a trusted reverse proxy.
func NewWithTrustedProxy(requests int, window time.Duration, trustProxyHeaders bool) *Limiter {
	if requests < 1 {
		requests = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	now := time.Now()
	return &Limiter{
		requests:          requests,
		window:            window,
		clients:           make(map[string]bucket),
		lastCleanup:       now,
		trustProxyHeaders: trustProxyHeaders,
	}
}

func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := clientIP(r.RemoteAddr, r.Header.Get("X-Forwarded-For"), l.trustProxyHeaders)
		allowed, retryAfter := l.allow(client, time.Now())
		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", retryAfter)
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "rate_limit_exceeded",
					"message": "request limit exceeded; retry after the current window",
				},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *Limiter) allow(client string, now time.Time) (bool, string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastCleanup) >= l.window {
		for key, entry := range l.clients {
			if now.Sub(entry.started) >= l.window {
				delete(l.clients, key)
			}
		}
		l.lastCleanup = now
	}

	entry, exists := l.clients[client]
	if !exists || now.Sub(entry.started) >= l.window {
		l.clients[client] = bucket{started: now, count: 1}
		return true, ""
	}
	if entry.count >= l.requests {
		remaining := time.Until(entry.started.Add(l.window))
		if remaining < time.Second {
			remaining = time.Second
		}
		return false, formatSeconds(remaining)
	}
	entry.count++
	l.clients[client] = entry
	return true, ""
}

func clientIP(remoteAddr string, forwardedFor string, trustProxyHeaders bool) string {
	if trustProxyHeaders {
		forwarded := strings.Split(forwardedFor, ",")
		if len(forwarded) > 0 {
			candidate := strings.TrimSpace(forwarded[len(forwarded)-1])
			if ip := net.ParseIP(candidate); ip != nil {
				return ip.String()
			}
		}
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}

func formatSeconds(duration time.Duration) string {
	seconds := int(duration.Seconds())
	if time.Duration(seconds)*time.Second < duration {
		seconds++
	}
	if seconds < 1 {
		seconds = 1
	}
	return strconv.Itoa(seconds)
}
