package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"recipebox-backend-go/internal/utils"
)

type fixedWindowLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]windowBucket
}

type windowBucket struct {
	count      int
	windowEnds time.Time
}

func NewAuthRateLimit(limitPerMinute int) func(http.Handler) http.Handler {
	if limitPerMinute <= 0 {
		limitPerMinute = 30
	}
	limiter := &fixedWindowLimiter{
		limit:   limitPerMinute,
		window:  time.Minute,
		buckets: make(map[string]windowBucket),
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractRequestIP(r)
			if !limiter.allow(ip + ":" + r.URL.Path) {
				utils.Error(w, http.StatusTooManyRequests, "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (l *fixedWindowLimiter) allow(key string) bool {
	now := time.Now().UTC()

	l.mu.Lock()
	defer l.mu.Unlock()

	for k, b := range l.buckets {
		if now.After(b.windowEnds) {
			delete(l.buckets, k)
		}
	}

	b := l.buckets[key]
	if b.windowEnds.IsZero() || now.After(b.windowEnds) {
		l.buckets[key] = windowBucket{
			count:      1,
			windowEnds: now.Add(l.window),
		}
		return true
	}

	if b.count >= l.limit {
		return false
	}
	b.count++
	l.buckets[key] = b
	return true
}

func extractRequestIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}
