package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"recipebox-backend-go/internal/dto"
	"recipebox-backend-go/internal/entity"
)

type mockAuthRepo struct {
	createUserFn                     func(ctx context.Context, name, email, passwordHash string) (entity.User, error)
	findUserByEmailFn                func(ctx context.Context, email string) (entity.User, error)
	findUserByIDFn                   func(ctx context.Context, id int64) (entity.User, error)
	updateUserPasswordFn             func(ctx context.Context, userID int64, passwordHash string) error
	saveRefreshTokenFn               func(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, userAgent, ip string) error
	findRefreshOwnerFn               func(ctx context.Context, tokenHash string, now time.Time) (int64, error)
	findRefreshTokenByHashFn         func(ctx context.Context, tokenHash string) (entity.RefreshToken, error)
	rotateRefreshTokenFn             func(ctx context.Context, oldTokenHash, newTokenHash string, newExpiresAt, now time.Time, userAgent, ip string) error
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

func (m mockAuthRepo) FindRefreshTokenOwner(ctx context.Context, tokenHash string, now time.Time) (int64, error) {
	if m.findRefreshOwnerFn == nil {
		return 0, errors.New("unexpected FindRefreshTokenOwner call")
	}
	return m.findRefreshOwnerFn(ctx, tokenHash, now)
}

func (m mockAuthRepo) RotateRefreshToken(ctx context.Context, oldTokenHash, newTokenHash string, newExpiresAt, now time.Time, userAgent, ip string) error {
	if m.rotateRefreshTokenFn == nil {
		return errors.New("unexpected RotateRefreshToken call")
	}
	return m.rotateRefreshTokenFn(ctx, oldTokenHash, newTokenHash, newExpiresAt, now, userAgent, ip)
}

func (m mockAuthRepo) FindRefreshTokenByHash(ctx context.Context, tokenHash string) (entity.RefreshToken, error) {
	if m.findRefreshTokenByHashFn == nil {
		return entity.RefreshToken{}, errors.New("unexpected FindRefreshTokenByHash call")
	}
	return m.findRefreshTokenByHashFn(ctx, tokenHash)
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
	}

	svc := NewAuthService(repo, strings.Repeat("a", 32), 15*time.Minute, 24*time.Hour, 10)

	resp, err := svc.Register(context.Background(), dto.RegisterRequest{
		Name:     "Kahfi Smith",
		Email:    " User@Example.com ",
		Password: "secret123",
	})
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
	})
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
	})
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

func TestLoginRejectsUnverifiedEmail(t *testing.T) {
	t.Parallel()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret123"), 10)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	repo := mockAuthRepo{
		findUserByEmailFn: func(_ context.Context, _ string) (entity.User, error) {
			return entity.User{
				ID:           5,
				Email:        "user@example.com",
				PasswordHash: string(passwordHash),
				// EmailVerifiedAt nil means not yet verified.
			}, nil
		},
		saveRefreshTokenFn: func(_ context.Context, _ int64, _ string, _ time.Time, _, _ string) error {
			t.Fatalf("SaveRefreshToken() should not be called for unverified email")
			return nil
		},
	}

	svc := NewAuthService(repo, strings.Repeat("c", 32), 15*time.Minute, 24*time.Hour, 10)

	_, err = svc.Login(context.Background(), dto.LoginRequest{
		Email:    "user@example.com",
		Password: "secret123",
	}, "", "")
	if !errors.Is(err, entity.ErrEmailNotVerified) {
		t.Fatalf("expected ErrEmailNotVerified, got %v", err)
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
		findRefreshOwnerFn: func(_ context.Context, tokenHash string, now time.Time) (int64, error) {
			if tokenHash != expectedOldHash {
				t.Fatalf("unexpected old hash: %s", tokenHash)
			}
			if !now.Equal(fixedNow) {
				t.Fatalf("unexpected now: %v", now)
			}
			return 7, nil
		},
		findRefreshTokenByHashFn: func(_ context.Context, tokenHash string) (entity.RefreshToken, error) {
			if tokenHash != expectedOldHash {
				t.Fatalf("unexpected lookup hash: %q", tokenHash)
			}
			ip := "::1"
			return entity.RefreshToken{
				UserID:    7,
				UserAgent: "unit-test",
				IPAddress: &ip,
			}, nil
		},
		findUserByIDFn: func(_ context.Context, id int64) (entity.User, error) {
			if id != 7 {
				t.Fatalf("unexpected user id: %d", id)
			}
			return entity.User{ID: 7, Email: "user@example.com"}, nil
		},
		rotateRefreshTokenFn: func(_ context.Context, oldTokenHash, newTokenHash string, newExpiresAt, now time.Time, _ string, ip string) error {
			rotatedOldHash = oldTokenHash
			rotatedNewHash = newTokenHash
			rotatedIP = ip
			if !now.Equal(fixedNow) {
				t.Fatalf("unexpected rotate time: %v", now)
			}
			if !newExpiresAt.Equal(fixedNow.Add(24 * time.Hour)) {
				t.Fatalf("unexpected refresh expiry: %v", newExpiresAt)
			}
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

func TestRefreshReuseDetectionRevokesAllUserTokens(t *testing.T) {
	t.Parallel()

	oldToken := "refresh-token-old"
	oldHash := hashToken(oldToken)

	revokedAll := false
	repo := mockAuthRepo{
		findRefreshOwnerFn: func(_ context.Context, tokenHash string, _ time.Time) (int64, error) {
			if tokenHash != oldHash {
				t.Fatalf("unexpected hash %q", tokenHash)
			}
			return 0, entity.ErrNotFound
		},
		findRefreshTokenByHashFn: func(_ context.Context, tokenHash string) (entity.RefreshToken, error) {
			if tokenHash != oldHash {
				t.Fatalf("unexpected hash %q", tokenHash)
			}
			replacedBy := "new-hash"
			now := time.Now().UTC()
			return entity.RefreshToken{
				UserID:              77,
				RevokedAt:           &now,
				ReplacedByTokenHash: &replacedBy,
			}, nil
		},
		revokeAllUserRefreshTokensFn: func(_ context.Context, userID int64) error {
			if userID != 77 {
				t.Fatalf("unexpected userID %d", userID)
			}
			revokedAll = true
			return nil
		},
	}
	svc := NewAuthService(repo, strings.Repeat("e", 32), 15*time.Minute, 24*time.Hour, 10)

	_, err := svc.Refresh(context.Background(), oldToken, "unit-test", "127.0.0.1")
	if !errors.Is(err, entity.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
	}
	if !revokedAll {
		t.Fatalf("expected all user refresh tokens to be revoked")
	}
}

func TestRefreshRejectsMetadataMismatch(t *testing.T) {
	t.Parallel()

	oldToken := "refresh-token-old"
	oldHash := hashToken(oldToken)
	revokedAll := false

	repo := mockAuthRepo{
		findRefreshOwnerFn: func(_ context.Context, tokenHash string, _ time.Time) (int64, error) {
			if tokenHash != oldHash {
				t.Fatalf("unexpected hash %q", tokenHash)
			}
			return 55, nil
		},
		findRefreshTokenByHashFn: func(_ context.Context, tokenHash string) (entity.RefreshToken, error) {
			if tokenHash != oldHash {
				t.Fatalf("unexpected hash %q", tokenHash)
			}
			ip := "127.0.0.1"
			return entity.RefreshToken{
				UserID:    55,
				UserAgent: "stored-agent",
				IPAddress: &ip,
			}, nil
		},
		revokeAllUserRefreshTokensFn: func(_ context.Context, userID int64) error {
			if userID != 55 {
				t.Fatalf("unexpected userID %d", userID)
			}
			revokedAll = true
			return nil
		},
		findUserByIDFn: func(_ context.Context, _ int64) (entity.User, error) {
			t.Fatalf("FindUserByID should not be reached on metadata mismatch")
			return entity.User{}, nil
		},
		rotateRefreshTokenFn: func(_ context.Context, _, _ string, _, _ time.Time, _, _ string) error {
			t.Fatalf("RotateRefreshToken should not be reached on metadata mismatch")
			return nil
		},
	}

	svc := NewAuthService(repo, strings.Repeat("e", 32), 15*time.Minute, 24*time.Hour, 10)

	_, err := svc.Refresh(context.Background(), oldToken, "new-agent", "127.0.0.1")
	if !errors.Is(err, entity.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
	}
	if !revokedAll {
		t.Fatalf("expected all user refresh tokens to be revoked")
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

func TestParseAccessTokenRejectsWrongAudience(t *testing.T) {
	t.Parallel()

	secret := strings.Repeat("f", 32)
	svc := NewAuthService(mockAuthRepo{}, secret, 15*time.Minute, 24*time.Hour, 10)

	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Issuer:    accessTokenIssuer,
		Subject:   "42",
		Audience:  []string{"another-client"},
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

	if _, err := svc.ParseAccessToken(signed); err == nil {
		t.Fatalf("expected parse error for wrong audience")
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

func TestResetPasswordConsumesAllActiveTokens(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
	var capturedHash string
	var capturedPasswordHash string

	repo := mockAuthRepo{
		consumePasswordResetAndUpdateFn: func(_ context.Context, tokenHash, newPasswordHash string, now time.Time) error {
			capturedHash = tokenHash
			capturedPasswordHash = newPasswordHash
			if !now.Equal(fixedNow) {
				t.Fatalf("unexpected reset time: %v", now)
			}
			return nil
		},
	}

	svc := NewAuthService(repo, strings.Repeat("i", 32), 15*time.Minute, 24*time.Hour, 10)
	svc.now = func() time.Time { return fixedNow }

	err := svc.ResetPassword(context.Background(), "reset-token", "newPassword123")
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}
	if capturedHash != hashToken("reset-token") {
		t.Fatalf("unexpected token hash: %q", capturedHash)
	}
	if capturedPasswordHash == "" || capturedPasswordHash == "newPassword123" {
		t.Fatalf("expected hashed password")
	}
	if bcrypt.CompareHashAndPassword([]byte(capturedPasswordHash), []byte("newPassword123")) != nil {
		t.Fatalf("stored password hash does not match new password")
	}
}
