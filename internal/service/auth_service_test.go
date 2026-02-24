package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"recipebox-backend-go/internal/dto"
	"recipebox-backend-go/internal/entity"
)

type mockAuthRepo struct {
	createUserFn                     func(ctx context.Context, name, email, passwordHash string) (entity.User, error)
	findUserByEmailFn                func(ctx context.Context, email string) (entity.User, error)
	findUserByIDFn                   func(ctx context.Context, id int64) (entity.User, error)
	updateUserPasswordFn             func(ctx context.Context, userID int64, passwordHash string) error
	saveRefreshTokenFn               func(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, userAgent, ip string) error
	findRefreshOwnerFn               func(ctx context.Context, tokenHash string) (int64, error)
	rotateRefreshTokenFn             func(ctx context.Context, oldTokenHash, newTokenHash string, newExpiresAt time.Time, userAgent, ip string) error
	revokeRefreshTokenFn             func(ctx context.Context, tokenHash string) error
	revokeAllUserRefreshTokensFn     func(ctx context.Context, userID int64) error
	saveEmailVerificationTokenFn     func(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	consumeEmailVerificationTokenFn  func(ctx context.Context, tokenHash string, now time.Time) error
	savePasswordResetTokenFn         func(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	consumePasswordResetAndUpdateFn  func(ctx context.Context, tokenHash, newPasswordHash string, now time.Time) error
}

func (m mockAuthRepo) CreateUser(ctx context.Context, name, email, passwordHash string) (entity.User, error) {
	if m.createUserFn == nil {
		return entity.User{}, errors.New("unexpected CreateUser call")
	}
	return m.createUserFn(ctx, name, email, passwordHash)
}

func (m mockAuthRepo) FindUserByEmail(ctx context.Context, email string) (entity.User, error) {
	if m.findUserByEmailFn == nil {
		return entity.User{}, errors.New("unexpected FindUserByEmail call")
	}
	return m.findUserByEmailFn(ctx, email)
}

func (m mockAuthRepo) FindUserByID(ctx context.Context, id int64) (entity.User, error) {
	if m.findUserByIDFn == nil {
		return entity.User{}, errors.New("unexpected FindUserByID call")
	}
	return m.findUserByIDFn(ctx, id)
}

func (m mockAuthRepo) UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error {
	if m.updateUserPasswordFn == nil {
		return errors.New("unexpected UpdateUserPassword call")
	}
	return m.updateUserPasswordFn(ctx, userID, passwordHash)
}

func (m mockAuthRepo) SaveRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, userAgent, ip string) error {
	if m.saveRefreshTokenFn == nil {
		return errors.New("unexpected SaveRefreshToken call")
	}
	return m.saveRefreshTokenFn(ctx, userID, tokenHash, expiresAt, userAgent, ip)
}

func (m mockAuthRepo) FindRefreshTokenOwner(ctx context.Context, tokenHash string) (int64, error) {
	if m.findRefreshOwnerFn == nil {
		return 0, errors.New("unexpected FindRefreshTokenOwner call")
	}
	return m.findRefreshOwnerFn(ctx, tokenHash)
}

func (m mockAuthRepo) RotateRefreshToken(ctx context.Context, oldTokenHash, newTokenHash string, newExpiresAt time.Time, userAgent, ip string) error {
	if m.rotateRefreshTokenFn == nil {
		return errors.New("unexpected RotateRefreshToken call")
	}
	return m.rotateRefreshTokenFn(ctx, oldTokenHash, newTokenHash, newExpiresAt, userAgent, ip)
}

func (m mockAuthRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	if m.revokeRefreshTokenFn == nil {
		return errors.New("unexpected RevokeRefreshToken call")
	}
	return m.revokeRefreshTokenFn(ctx, tokenHash)
}

func (m mockAuthRepo) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error {
	if m.revokeAllUserRefreshTokensFn == nil {
		return errors.New("unexpected RevokeAllUserRefreshTokens call")
	}
	return m.revokeAllUserRefreshTokensFn(ctx, userID)
}

func (m mockAuthRepo) SaveEmailVerificationToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	if m.saveEmailVerificationTokenFn == nil {
		return errors.New("unexpected SaveEmailVerificationToken call")
	}
	return m.saveEmailVerificationTokenFn(ctx, userID, tokenHash, expiresAt)
}

func (m mockAuthRepo) ConsumeEmailVerificationToken(ctx context.Context, tokenHash string, now time.Time) error {
	if m.consumeEmailVerificationTokenFn == nil {
		return errors.New("unexpected ConsumeEmailVerificationToken call")
	}
	return m.consumeEmailVerificationTokenFn(ctx, tokenHash, now)
}

func (m mockAuthRepo) SavePasswordResetToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	if m.savePasswordResetTokenFn == nil {
		return errors.New("unexpected SavePasswordResetToken call")
	}
	return m.savePasswordResetTokenFn(ctx, userID, tokenHash, expiresAt)
}

func (m mockAuthRepo) ConsumePasswordResetTokenAndUpdatePassword(ctx context.Context, tokenHash, newPasswordHash string, now time.Time) error {
	if m.consumePasswordResetAndUpdateFn == nil {
		return errors.New("unexpected ConsumePasswordResetTokenAndUpdatePassword call")
	}
	return m.consumePasswordResetAndUpdateFn(ctx, tokenHash, newPasswordHash, now)
}

func TestRegisterSuccess(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC)
	var savedHash string
	var savedIP string

	repo := mockAuthRepo{
		createUserFn: func(_ context.Context, name, email, passwordHash string) (entity.User, error) {
			if name != "Kahfi Smith" {
				t.Fatalf("unexpected name: %s", name)
			}
			if email != "user@example.com" {
				t.Fatalf("unexpected email: %s", email)
			}
			if passwordHash == "" || passwordHash == "secret123" {
				t.Fatalf("password hash not generated")
			}
			return entity.User{ID: 11, Name: name, Email: email, PasswordHash: passwordHash}, nil
		},
		saveRefreshTokenFn: func(_ context.Context, userID int64, tokenHash string, expiresAt time.Time, userAgent, ip string) error {
			if userID != 11 {
				t.Fatalf("unexpected user id: %d", userID)
			}
			if userAgent != "unit-test" {
				t.Fatalf("unexpected user agent: %s", userAgent)
			}
			savedHash = tokenHash
			savedIP = ip
			if expiresAt.Before(fixedNow) {
				t.Fatalf("refresh expiry should be after now")
			}
			return nil
		},
	}

	svc := NewAuthService(repo, strings.Repeat("a", 32), 15*time.Minute, 24*time.Hour, 10)
	svc.now = func() time.Time { return fixedNow }

	resp, err := svc.Register(context.Background(), dto.RegisterRequest{
		Name:     "Kahfi Smith",
		Email:    " User@Example.com ",
		Password: "secret123",
	}, "unit-test", " 127.0.0.1 ")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if resp.User.Email != "user@example.com" {
		t.Fatalf("unexpected normalized email: %s", resp.User.Email)
	}
	if resp.User.Name != "Kahfi Smith" {
		t.Fatalf("unexpected name: %s", resp.User.Name)
	}
	if resp.User.PasswordHash != "" {
		t.Fatalf("user password hash must be hidden in response")
	}
	if resp.Tokens.AccessToken == "" || resp.Tokens.RefreshToken == "" {
		t.Fatalf("expected token pair to be generated")
	}
	if savedHash != hashToken(resp.Tokens.RefreshToken) {
		t.Fatalf("stored refresh hash does not match token")
	}
	if savedIP != "127.0.0.1" {
		t.Fatalf("ip should be sanitized, got: %q", savedIP)
	}
}

func TestRegisterEmailTaken(t *testing.T) {
	t.Parallel()

	repo := mockAuthRepo{
		createUserFn: func(_ context.Context, _, _, _ string) (entity.User, error) {
			return entity.User{}, entity.ErrConflict
		},
		saveRefreshTokenFn: func(_ context.Context, _ int64, _ string, _ time.Time, _, _ string) error {
			return nil
		},
	}

	svc := NewAuthService(repo, strings.Repeat("b", 32), 15*time.Minute, 24*time.Hour, 10)

	_, err := svc.Register(context.Background(), dto.RegisterRequest{
		Name:     "Kahfi Smith",
		Email:    "user@example.com",
		Password: "secret123",
	}, "", "")
	if !errors.Is(err, entity.ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRegisterNameRequired(t *testing.T) {
	t.Parallel()

	svc := NewAuthService(mockAuthRepo{}, strings.Repeat("z", 32), 15*time.Minute, 24*time.Hour, 10)

	_, err := svc.Register(context.Background(), dto.RegisterRequest{
		Name:     "   ",
		Email:    "user@example.com",
		Password: "secret123",
	}, "", "")
	if err == nil || err.Error() != "name is required" {
		t.Fatalf("expected name required validation error, got %v", err)
	}
}

func TestLoginInvalidCredentialsWhenUserMissing(t *testing.T) {
	t.Parallel()

	repo := mockAuthRepo{
		findUserByEmailFn: func(_ context.Context, _ string) (entity.User, error) {
			return entity.User{}, entity.ErrNotFound
		},
	}

	svc := NewAuthService(repo, strings.Repeat("c", 32), 15*time.Minute, 24*time.Hour, 10)

	_, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "user@example.com",
		Password: "secret123",
	}, "", "")
	if !errors.Is(err, entity.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestRefreshRejectsEmptyToken(t *testing.T) {
	t.Parallel()

	svc := NewAuthService(mockAuthRepo{}, strings.Repeat("d", 32), 15*time.Minute, 24*time.Hour, 10)
	_, err := svc.Refresh(context.Background(), "   ", "", "")
	if !errors.Is(err, entity.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefreshSuccessRotation(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC)
	oldToken := "refresh-token-old"
	expectedOldHash := hashToken(oldToken)

	var rotatedOldHash string
	var rotatedNewHash string
	var rotatedIP string

	repo := mockAuthRepo{
		findRefreshOwnerFn: func(_ context.Context, tokenHash string) (int64, error) {
			if tokenHash != expectedOldHash {
				t.Fatalf("unexpected old hash: %s", tokenHash)
			}
			return 7, nil
		},
		findUserByIDFn: func(_ context.Context, id int64) (entity.User, error) {
			if id != 7 {
				t.Fatalf("unexpected user id: %d", id)
			}
			return entity.User{ID: 7, Email: "user@example.com"}, nil
		},
		rotateRefreshTokenFn: func(_ context.Context, oldTokenHash, newTokenHash string, _ time.Time, _ string, ip string) error {
			rotatedOldHash = oldTokenHash
			rotatedNewHash = newTokenHash
			rotatedIP = ip
			return nil
		},
	}

	svc := NewAuthService(repo, strings.Repeat("e", 32), 15*time.Minute, 24*time.Hour, 10)
	svc.now = func() time.Time { return fixedNow }

	tokens, err := svc.Refresh(context.Background(), oldToken, "unit-test", "::1")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected new token pair")
	}
	if rotatedOldHash != expectedOldHash {
		t.Fatalf("unexpected rotated old hash")
	}
	if rotatedNewHash != hashToken(tokens.RefreshToken) {
		t.Fatalf("new hash should match returned refresh token")
	}
	if rotatedIP != "::1" {
		t.Fatalf("expected sanitized ipv6, got %q", rotatedIP)
	}
}

func TestParseAccessToken(t *testing.T) {
	t.Parallel()

	svc := NewAuthService(mockAuthRepo{}, strings.Repeat("f", 32), 15*time.Minute, 24*time.Hour, 10)
	svc.now = func() time.Time { return time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC) }

	token, _, err := svc.createAccessToken(42)
	if err != nil {
		t.Fatalf("createAccessToken() error = %v", err)
	}

	userID, err := svc.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}
	if userID != 42 {
		t.Fatalf("expected user ID 42, got %d", userID)
	}

	if _, err := svc.ParseAccessToken("not-a-token"); err == nil {
		t.Fatalf("expected parse error for malformed token")
	}
}

func TestRequestEmailVerificationSuccess(t *testing.T) {
	t.Parallel()

	var savedHash string
	repo := mockAuthRepo{
		findUserByEmailFn: func(_ context.Context, email string) (entity.User, error) {
			if email != "user@example.com" {
				t.Fatalf("unexpected email %q", email)
			}
			return entity.User{ID: 5, Email: email}, nil
		},
		saveEmailVerificationTokenFn: func(_ context.Context, userID int64, tokenHash string, _ time.Time) error {
			if userID != 5 {
				t.Fatalf("unexpected userID %d", userID)
			}
			savedHash = tokenHash
			return nil
		},
	}
	svc := NewAuthService(repo, strings.Repeat("g", 32), 15*time.Minute, 24*time.Hour, 10)

	resp, err := svc.RequestEmailVerification(context.Background(), dto.EmailRequest{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("RequestEmailVerification() error = %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("expected verify token")
	}
	if savedHash != hashToken(resp.Token) {
		t.Fatalf("stored hash mismatch")
	}
}

func TestResetPasswordInvalidToken(t *testing.T) {
	t.Parallel()

	svc := NewAuthService(mockAuthRepo{}, strings.Repeat("h", 32), 15*time.Minute, 24*time.Hour, 10)
	err := svc.ResetPassword(context.Background(), "", "newPassword123")
	if !errors.Is(err, entity.ErrInvalidResetToken) {
		t.Fatalf("expected ErrInvalidResetToken, got %v", err)
	}
}
