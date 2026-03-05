package middleware

import (
	"context"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"recipebox-backend-go/internal/utils"
)

type AuthRateLimitStore interface {
	Increment(ctx context.Context, key string, window time.Duration) (int64, error)
}

type redisAuthRateLimitStore struct {
	redisClient *redis.Client
}

const rateLimitIncrementScript = "local current = redis.call('INCR', KEYS[1]); if current == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]); end; return current"

func NewRedisAuthRateLimitStore(redisClient *redis.Client) AuthRateLimitStore {
	return &redisAuthRateLimitStore{redisClient: redisClient}
}

func NewAuthRateLimit(store AuthRateLimitStore, limitPerMinute int, trustedProxies []*net.IPNet) func(http.Handler) http.Handler {
	if limitPerMinute <= 0 {
		limitPerMinute = 30
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractRequestIP(r, trustedProxies)
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
	result, err := s.redisClient.Eval(ctx, rateLimitIncrementScript, []string{key}, strconv.FormatInt(window.Milliseconds(), 10)).Int64()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func extractRequestIP(r *http.Request, trustedProxies []*net.IPNet) string {
	return utils.ClientIP(r, trustedProxies)
}
