package middleware

import (
	"context"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"recipebox-backend-go/internal/utils"
)

type AuthRateLimitStore interface {
	Increment(ctx context.Context, key string, window time.Duration) (int64, error)
}

type redisEvaler interface {
	EvalInt(ctx context.Context, script string, keys []string, args []string) (int64, error)
}

type redisAuthRateLimitStore struct {
	client redisEvaler
}

const rateLimitIncrementScript = "local current = redis.call('INCR', KEYS[1]); if current == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]); end; return current"

func NewRedisAuthRateLimitStore(client redisEvaler) AuthRateLimitStore {
	return &redisAuthRateLimitStore{client: client}
}

func NewAuthRateLimit(store AuthRateLimitStore, limitPerMinute int) func(http.Handler) http.Handler {
	if limitPerMinute <= 0 {
		limitPerMinute = 30
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractRequestIP(r)
			if store == nil {
				utils.Error(w, http.StatusServiceUnavailable, "rate limit unavailable")
				return
			}

			allowed, err := allowRequest(r.Context(), store, ip+":"+r.URL.Path, limitPerMinute, time.Minute)
			if err != nil {
				log.Printf("rate limit error: %v", err)
				utils.Error(w, http.StatusServiceUnavailable, "rate limit unavailable")
				return
			}
			if !allowed {
				utils.Error(w, http.StatusTooManyRequests, "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func allowRequest(ctx context.Context, store AuthRateLimitStore, key string, limit int, window time.Duration) (bool, error) {
	current, err := store.Increment(ctx, "rl:auth:"+key, window)
	if err != nil {
		return false, err
	}
	return current <= int64(limit), nil
}

func (s *redisAuthRateLimitStore) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	result, err := s.client.EvalInt(ctx, rateLimitIncrementScript, []string{key}, []string{strconv.FormatInt(window.Milliseconds(), 10)})
	if err != nil {
		return 0, err
	}
	return result, nil
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
