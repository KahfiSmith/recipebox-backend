package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"recipebox-backend-go/internal/dto"
	"recipebox-backend-go/internal/entity"
	"recipebox-backend-go/internal/repository"
	"recipebox-backend-go/internal/service"
	"recipebox-backend-go/internal/utils"
)

type mockAuthRepo struct {
	findUserByEmailFn               func(ctx context.Context, email string) (entity.User, error)
	findUserByIDFn                  func(ctx context.Context, id int64) (entity.User, error)
	saveRefreshTokenFn              func(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, userAgent, ip string) error
	findRefreshOwnerFn              func(ctx context.Context, tokenHash string, now time.Time) (int64, error)
	rotateRefreshTokenFn            func(ctx context.Context, oldTokenHash, newTokenHash string, newExpiresAt, now time.Time, userAgent, ip string) error
	revokeRefreshTokenFn            func(ctx context.Context, tokenHash string) error
	findRefreshTokenByHashFn        func(ctx context.Context, tokenHash string) (entity.RefreshToken, error)
	revokeAllUserRefreshTokensFn    func(ctx context.Context, userID int64) error
	consumePasswordResetAndUpdateFn func(ctx context.Context, tokenHash, newPasswordHash string, now time.Time) error
}

var _ repository.AuthRepository = mockAuthRepo{}

func (m mockAuthRepo) CreateUser(context.Context, string, string, string) (entity.User, error) {
	return entity.User{}, nil
}

func (m mockAuthRepo) FindUserByEmail(ctx context.Context, email string) (entity.User, error) {
	if m.findUserByEmailFn == nil {
		return entity.User{}, entity.ErrNotFound
	}
	return m.findUserByEmailFn(ctx, email)
}

func (m mockAuthRepo) FindUserByID(ctx context.Context, id int64) (entity.User, error) {
	if m.findUserByIDFn == nil {
		return entity.User{}, entity.ErrNotFound
	}
	return m.findUserByIDFn(ctx, id)
}

func (m mockAuthRepo) UpdateUserPassword(context.Context, int64, string) error {
	return nil
}

func (m mockAuthRepo) SaveRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, userAgent, ip string) error {
	if m.saveRefreshTokenFn == nil {
		return nil
	}
	return m.saveRefreshTokenFn(ctx, userID, tokenHash, expiresAt, userAgent, ip)
}

func (m mockAuthRepo) FindRefreshTokenOwner(ctx context.Context, tokenHash string, now time.Time) (int64, error) {
	if m.findRefreshOwnerFn == nil {
		return 0, entity.ErrNotFound
	}
	return m.findRefreshOwnerFn(ctx, tokenHash, now)
}

func (m mockAuthRepo) FindRefreshTokenByHash(ctx context.Context, tokenHash string) (entity.RefreshToken, error) {
	if m.findRefreshTokenByHashFn == nil {
		return entity.RefreshToken{}, entity.ErrNotFound
	}
	return m.findRefreshTokenByHashFn(ctx, tokenHash)
}

func (m mockAuthRepo) RotateRefreshToken(ctx context.Context, oldTokenHash, newTokenHash string, newExpiresAt, now time.Time, userAgent, ip string) error {
	if m.rotateRefreshTokenFn == nil {
		return nil
	}
	return m.rotateRefreshTokenFn(ctx, oldTokenHash, newTokenHash, newExpiresAt, now, userAgent, ip)
}

func (m mockAuthRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	if m.revokeRefreshTokenFn == nil {
		return nil
	}
	return m.revokeRefreshTokenFn(ctx, tokenHash)
}

func (m mockAuthRepo) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error {
	if m.revokeAllUserRefreshTokensFn == nil {
		return nil
	}
	return m.revokeAllUserRefreshTokensFn(ctx, userID)
}

func (m mockAuthRepo) SaveEmailVerificationToken(context.Context, int64, string, time.Time) error {
	return nil
}

func (m mockAuthRepo) ConsumeEmailVerificationToken(context.Context, string, time.Time) error {
	return nil
}

func (m mockAuthRepo) SavePasswordResetToken(context.Context, int64, string, time.Time) error {
	return nil
}

func (m mockAuthRepo) ConsumePasswordResetTokenAndUpdatePassword(ctx context.Context, tokenHash, newPasswordHash string, now time.Time) error {
	if m.consumePasswordResetAndUpdateFn == nil {
		return nil
	}
	return m.consumePasswordResetAndUpdateFn(ctx, tokenHash, newPasswordHash, now)
}

func TestLoginSetsRefreshCookieAndHidesTokenInBody(t *testing.T) {
	t.Parallel()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}

	repo := mockAuthRepo{
		findUserByEmailFn: func(_ context.Context, email string) (entity.User, error) {
			if email != "user@example.com" {
				t.Fatalf("unexpected email %q", email)
			}
			now := time.Date(2026, 2, 23, 8, 0, 0, 0, time.UTC)
			return entity.User{
				ID:              42,
				Name:            "User",
				Email:           email,
				PasswordHash:    string(passwordHash),
				EmailVerifiedAt: &now,
			}, nil
		},
		saveRefreshTokenFn: func(_ context.Context, userID int64, tokenHash string, expiresAt time.Time, userAgent, ip string) error {
			if userID != 42 {
				t.Fatalf("unexpected userID %d", userID)
			}
			if tokenHash == "" {
				t.Fatalf("expected token hash")
			}
			if userAgent != "test-agent" {
				t.Fatalf("unexpected userAgent %q", userAgent)
			}
			if ip != "127.0.0.1" {
				t.Fatalf("unexpected ip %q", ip)
			}
			if expiresAt.IsZero() {
				t.Fatalf("expected refresh token expiry")
			}
			return nil
		},
	}

	authService := service.NewAuthService(repo, strings.Repeat("a", 32), 15*time.Minute, 24*time.Hour, bcrypt.MinCost)
	controller := NewAuthController(authService, true, 24*time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"user@example.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:4321"
	rec := httptest.NewRecorder()

	controller.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	res := rec.Result()
	cookie := findCookie(t, res.Cookies(), refreshTokenCookieName)
	if cookie.Value == "" {
		t.Fatalf("expected refresh token cookie value")
	}
	if !cookie.HttpOnly || !cookie.Secure {
		t.Fatalf("expected secure httpOnly refresh cookie")
	}
	if cookie.Path != "/api/v1/auth" {
		t.Fatalf("unexpected cookie path %q", cookie.Path)
	}

	var payload struct {
		Data dto.AuthResponse `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Tokens.AccessToken == "" {
		t.Fatalf("expected access token in response body")
	}
	if payload.Data.Tokens.RefreshToken != "" {
		t.Fatalf("refresh token should be omitted from response body")
	}
	if !payload.Data.Tokens.RefreshTokenExpiresAt.IsZero() {
		t.Fatalf("refresh token expiry should be omitted from response body")
	}
}

func TestRefreshUsesCookieAndRotatesRefreshToken(t *testing.T) {
	t.Parallel()

	oldTokenHash := ""
	start := time.Now()

	repo := mockAuthRepo{
		findRefreshOwnerFn: func(_ context.Context, tokenHash string, now time.Time) (int64, error) {
			oldTokenHash = tokenHash
			if now.Before(start.Add(-time.Second)) {
				t.Fatalf("unexpected stale now %v", now)
			}
			return 7, nil
		},
		findUserByIDFn: func(_ context.Context, id int64) (entity.User, error) {
			if id != 7 {
				t.Fatalf("unexpected userID %d", id)
			}
			return entity.User{ID: 7, Email: "user@example.com"}, nil
		},
		rotateRefreshTokenFn: func(_ context.Context, oldTokenHashArg, newTokenHash string, newExpiresAt, now time.Time, userAgent, ip string) error {
			if oldTokenHashArg != oldTokenHash {
				t.Fatalf("expected old token hash to match lookup")
			}
			if newTokenHash == "" {
				t.Fatalf("expected new token hash")
			}
			if !newExpiresAt.After(now) {
				t.Fatalf("expected expiry after rotation time")
			}
			if newExpiresAt.Before(now.Add(23 * time.Hour)) || newExpiresAt.After(now.Add(24*time.Hour+time.Minute)) {
				t.Fatalf("unexpected new expiry %v", newExpiresAt)
			}
			if userAgent != "refresh-agent" {
				t.Fatalf("unexpected userAgent %q", userAgent)
			}
			if ip != "127.0.0.1" {
				t.Fatalf("unexpected ip %q", ip)
			}
			return nil
		},
	}

	authService := service.NewAuthService(repo, strings.Repeat("b", 32), 15*time.Minute, 24*time.Hour, bcrypt.MinCost)
	controller := NewAuthController(authService, false, 24*time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", http.NoBody)
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh-token-old"})
	req.Header.Set("User-Agent", "refresh-agent")
	req.RemoteAddr = "127.0.0.1:9876"
	rec := httptest.NewRecorder()

	controller.Refresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	res := rec.Result()
	cookie := findCookie(t, res.Cookies(), refreshTokenCookieName)
	if cookie.Value == "" {
		t.Fatalf("expected rotated refresh token cookie")
	}
	if cookie.Secure {
		t.Fatalf("did not expect secure cookie in this test")
	}

	var payload struct {
		Data dto.TokenPair `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.AccessToken == "" {
		t.Fatalf("expected access token in response")
	}
	if payload.Data.RefreshToken != "" {
		t.Fatalf("refresh token should be omitted from response body")
	}
}

func TestLogoutRevokesRefreshTokenAndClearsCookie(t *testing.T) {
	t.Parallel()

	var revokedHash string
	repo := mockAuthRepo{
		revokeRefreshTokenFn: func(_ context.Context, tokenHash string) error {
			revokedHash = tokenHash
			return nil
		},
	}

	authService := service.NewAuthService(repo, strings.Repeat("c", 32), 15*time.Minute, 24*time.Hour, bcrypt.MinCost)
	controller := NewAuthController(authService, true, 24*time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", http.NoBody)
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "logout-token"})
	rec := httptest.NewRecorder()

	controller.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if revokedHash == "" {
		t.Fatalf("expected refresh token to be revoked")
	}

	res := rec.Result()
	cookie := findCookie(t, res.Cookies(), refreshTokenCookieName)
	if cookie.MaxAge != -1 {
		t.Fatalf("expected cookie to be cleared, got MaxAge %d", cookie.MaxAge)
	}
	if !cookie.Expires.Equal(time.Unix(0, 0)) {
		t.Fatalf("expected cookie expiry to be reset")
	}

	var payload map[string]string
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["message"] != "logged out" {
		t.Fatalf("unexpected message %q", payload["message"])
	}
}

func TestRegisterValidationReturnsBadRequest(t *testing.T) {
	t.Parallel()

	authService := service.NewAuthService(mockAuthRepo{}, strings.Repeat("d", 32), 15*time.Minute, 24*time.Hour, bcrypt.MinCost)
	controller := NewAuthController(authService, false, 24*time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"name":"","email":"invalid","password":"123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	controller.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var payload utils.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error == "" {
		t.Fatalf("expected validation error message")
	}
}

func findCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}
