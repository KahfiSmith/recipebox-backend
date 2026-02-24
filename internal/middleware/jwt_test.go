package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"recipebox-backend-go/internal/service"
	"recipebox-backend-go/internal/utils"
)

func TestAuthJWTMissingBearer(t *testing.T) {
	t.Parallel()

	authService := service.NewAuthService(nil, strings.Repeat("a", 32), 15*time.Minute, 24*time.Hour, 10)
	h := AuthJWT(authService)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var payload utils.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error != "missing bearer token" {
		t.Fatalf("unexpected error message: %s", payload.Error)
	}
}

func TestAuthJWTInvalidToken(t *testing.T) {
	t.Parallel()

	authService := service.NewAuthService(nil, strings.Repeat("a", 32), 15*time.Minute, 24*time.Hour, 10)
	h := AuthJWT(authService)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var payload utils.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error != "invalid access token" {
		t.Fatalf("unexpected error message: %s", payload.Error)
	}
}

func TestAuthJWTValidTokenSetsContext(t *testing.T) {
	t.Parallel()

	secret := strings.Repeat("a", 32)
	authService := service.NewAuthService(nil, secret, 15*time.Minute, 24*time.Hour, 10)
	token := makeToken(t, secret, 99)

	h := AuthJWT(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatalf("expected user id in context")
		}
		if userID != 99 {
			t.Fatalf("expected user ID 99, got %d", userID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func makeToken(t *testing.T, secret string, userID int64) string {
	t.Helper()

	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Issuer:    "recipebox-api",
		Subject:   fmt.Sprintf("%d", userID),
		Audience:  []string{"recipebox-client"},
		ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        "acc_test-token-id",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}
