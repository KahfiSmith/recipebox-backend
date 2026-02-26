package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthRateLimitBlocksAfterLimit(t *testing.T) {
	t.Parallel()

	h := NewAuthRateLimit(2)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
