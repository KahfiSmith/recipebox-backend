package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stubRateLimitStore struct {
	incrementFn func(ctx context.Context, key string, window time.Duration) (int64, error)
}

func (s stubRateLimitStore) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	if s.incrementFn == nil {
		return 0, errors.New("unexpected Increment call")
	}
	return s.incrementFn(ctx, key, window)
}

func TestAuthRateLimitBlocksAfterLimit(t *testing.T) {
	t.Parallel()

	counts := map[string]int64{}
	store := stubRateLimitStore{
		incrementFn: func(_ context.Context, key string, window time.Duration) (int64, error) {
			if window != time.Minute {
				t.Fatalf("unexpected window: %v", window)
			}
			counts[key]++
			return counts[key], nil
		},
	}

	h := NewAuthRateLimit(store, 2)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, req.Clone(req.Context()))
	if rec1.Code != http.StatusNoContent {
		t.Fatalf("first request should pass, got %d", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req.Clone(req.Context()))
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("second request should pass, got %d", rec2.Code)
	}

	rec3 := httptest.NewRecorder()
	h.ServeHTTP(rec3, req.Clone(req.Context()))
	if rec3.Code != http.StatusTooManyRequests {
		t.Fatalf("third request should be rate limited, got %d", rec3.Code)
	}
}

func TestAuthRateLimitReturnsServiceUnavailableOnStoreError(t *testing.T) {
	t.Parallel()

	h := NewAuthRateLimit(stubRateLimitStore{
		incrementFn: func(_ context.Context, _ string, _ time.Duration) (int64, error) {
			return 0, errors.New("redis unavailable")
		},
	}, 2)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected service unavailable, got %d", rec.Code)
	}
}
