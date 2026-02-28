package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"recipebox-backend-go/internal/dto"
	"recipebox-backend-go/internal/entity"
	"recipebox-backend-go/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo             repository.AuthRepository
	jwtSecret        []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	verifyEmailTTL   time.Duration
	resetPasswordTTL time.Duration
	bcryptCost       int
	now              func() time.Time
	emailSender      EmailSender
	frontendBaseURL  string
	exposeTokens     bool
}

const (
	accessTokenIssuer   = "recipebox-api"
	accessTokenAudience = "recipebox-client"
)

type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

func NewAuthService(repo repository.AuthRepository, jwtSecret string, accessTokenTTL, refreshTokenTTL time.Duration, bcryptCost int) *AuthService {
	return &AuthService{
		repo:             repo,
		jwtSecret:        []byte(jwtSecret),
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
		verifyEmailTTL:   24 * time.Hour,
		resetPasswordTTL: 1 * time.Hour,
		bcryptCost:       bcryptCost,
		now:              time.Now,
		exposeTokens:     true,
	}
}

func (s *AuthService) ConfigureEmailDelivery(sender EmailSender, frontendBaseURL string, exposeTokens bool) {
	s.emailSender = sender
	s.frontendBaseURL = strings.TrimRight(strings.TrimSpace(frontendBaseURL), "/")
	s.exposeTokens = exposeTokens
}

func (s *AuthService) Register(ctx context.Context, input dto.RegisterRequest) (dto.RegisterResponse, error) {
	name, err := normalizeName(input.Name)
	if err != nil {
		return dto.RegisterResponse{}, err
	}

	email, err := normalizeEmail(input.Email)
	if err != nil {
		return dto.RegisterResponse{}, err
	}
	if err := validatePassword(input.Password); err != nil {
		return dto.RegisterResponse{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.bcryptCost)
	if err != nil {
		return dto.RegisterResponse{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, name, email, string(passwordHash))
	if err != nil {
		if errors.Is(err, entity.ErrConflict) {
			return dto.RegisterResponse{}, entity.ErrEmailTaken
		}
		return dto.RegisterResponse{}, err
	}

	user.PasswordHash = ""
	return dto.RegisterResponse{User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, input dto.LoginRequest, userAgent, ip string) (dto.AuthResponse, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return dto.AuthResponse{}, entity.ErrInvalidCredentials
	}

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return dto.AuthResponse{}, entity.ErrInvalidCredentials
		}
		return dto.AuthResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return dto.AuthResponse{}, entity.ErrInvalidCredentials
	}
	if user.EmailVerifiedAt == nil {
		return dto.AuthResponse{}, entity.ErrEmailNotVerified
	}

	tokens, err := s.issueTokens(ctx, user.ID, userAgent, ip)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	user.PasswordHash = ""
	return dto.AuthResponse{User: user, Tokens: tokens}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken, userAgent, ip string) (dto.TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return dto.TokenPair{}, entity.ErrInvalidRefreshToken
	}

	oldHash := hashToken(refreshToken)
	now := s.now()
	userID, err := s.repo.FindRefreshTokenOwner(ctx, oldHash, now)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			s.detectRefreshTokenReuse(ctx, oldHash)
			return dto.TokenPair{}, entity.ErrInvalidRefreshToken
		}
		return dto.TokenPair{}, err
	}

	currentUserAgent := strings.TrimSpace(userAgent)
	currentIP := sanitizeIP(ip)
	storedToken, err := s.repo.FindRefreshTokenByHash(ctx, oldHash)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return dto.TokenPair{}, entity.ErrInvalidRefreshToken
		}
		return dto.TokenPair{}, err
	}
	if refreshTokenMetadataMismatch(storedToken, currentUserAgent, currentIP) {
		if err := s.repo.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
			return dto.TokenPair{}, err
		}
		return dto.TokenPair{}, entity.ErrInvalidRefreshToken
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return dto.TokenPair{}, entity.ErrInvalidRefreshToken
		}
		return dto.TokenPair{}, err
	}

	accessToken, accessExp, err := s.createAccessToken(user.ID)
	if err != nil {
		return dto.TokenPair{}, err
	}

	newRefreshToken, err := generateTokenString(48)
	if err != nil {
		return dto.TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshExp := now.Add(s.refreshTokenTTL)
	newHash := hashToken(newRefreshToken)
	if err := s.repo.RotateRefreshToken(ctx, oldHash, newHash, refreshExp, now, currentUserAgent, currentIP); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return dto.TokenPair{}, entity.ErrInvalidRefreshToken
		}
		return dto.TokenPair{}, err
	}

	return dto.TokenPair{AccessToken: accessToken, AccessTokenExpiresAt: accessExp, RefreshToken: newRefreshToken, RefreshTokenExpiresAt: refreshExp}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil
	}
	return s.repo.RevokeRefreshToken(ctx, hashToken(refreshToken))
}

func (s *AuthService) GetMe(ctx context.Context, userID int64) (entity.User, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return entity.User{}, err
	}
	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) RequestEmailVerification(ctx context.Context, input dto.EmailRequest) (dto.OneTimeTokenResponse, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return dto.OneTimeTokenResponse{}, err
	}

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return dto.OneTimeTokenResponse{}, entity.ErrNotFound
		}
		return dto.OneTimeTokenResponse{}, err
	}

	if user.EmailVerifiedAt != nil {
		return dto.OneTimeTokenResponse{}, nil
	}

	rawToken, err := generateTokenString(48)
	if err != nil {
		return dto.OneTimeTokenResponse{}, fmt.Errorf("generate verify token: %w", err)
	}

	expiresAt := s.now().Add(s.verifyEmailTTL)
	if err := s.repo.SaveEmailVerificationToken(ctx, user.ID, hashToken(rawToken), expiresAt); err != nil {
		return dto.OneTimeTokenResponse{}, err
	}

	if err := s.sendEmailVerification(ctx, user.Email, rawToken, expiresAt); err != nil {
		return dto.OneTimeTokenResponse{}, err
	}
	if !s.exposeTokens {
		rawToken = ""
	}

	return dto.OneTimeTokenResponse{Token: rawToken, ExpiresAt: expiresAt}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return entity.ErrInvalidVerifyToken
	}

	if err := s.repo.ConsumeEmailVerificationToken(ctx, hashToken(token), s.now()); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return entity.ErrInvalidVerifyToken
		}
		return err
	}
	return nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, input dto.EmailRequest) (dto.OneTimeTokenResponse, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return dto.OneTimeTokenResponse{}, err
	}

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return dto.OneTimeTokenResponse{}, entity.ErrNotFound
		}
		return dto.OneTimeTokenResponse{}, err
	}

	rawToken, err := generateTokenString(48)
	if err != nil {
		return dto.OneTimeTokenResponse{}, fmt.Errorf("generate password reset token: %w", err)
	}

	expiresAt := s.now().Add(s.resetPasswordTTL)
	if err := s.repo.SavePasswordResetToken(ctx, user.ID, hashToken(rawToken), expiresAt); err != nil {
		return dto.OneTimeTokenResponse{}, err
	}

	if err := s.sendPasswordReset(ctx, user.Email, rawToken, expiresAt); err != nil {
		return dto.OneTimeTokenResponse{}, err
	}
	if !s.exposeTokens {
		rawToken = ""
	}

	return dto.OneTimeTokenResponse{Token: rawToken, ExpiresAt: expiresAt}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return entity.ErrInvalidResetToken
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.bcryptCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	if err := s.repo.ConsumePasswordResetTokenAndUpdatePassword(ctx, hashToken(token), string(newPasswordHash), s.now()); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return entity.ErrInvalidResetToken
		}
		return err
	}

	return nil
}

func (s *AuthService) ParseAccessToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithLeeway(5*time.Second),
		jwt.WithIssuer(accessTokenIssuer),
		jwt.WithAudience(accessTokenAudience),
	)
	if err != nil {
		return 0, fmt.Errorf("parse access token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid access token")
	}
	if claims.Subject == "" {
		return 0, errors.New("missing subject")
	}

	var userID int64
	if _, err := fmt.Sscan(claims.Subject, &userID); err != nil {
		return 0, errors.New("invalid subject")
	}

	if claims.ID == "" || !strings.HasPrefix(claims.ID, "acc_") {
		return 0, errors.New("invalid token id")
	}

	return userID, nil
}

func (s *AuthService) issueTokens(ctx context.Context, userID int64, userAgent, ip string) (dto.TokenPair, error) {
	accessToken, accessExp, err := s.createAccessToken(userID)
	if err != nil {
		return dto.TokenPair{}, err
	}

	refreshToken, err := generateTokenString(48)
	if err != nil {
		return dto.TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshExp := s.now().Add(s.refreshTokenTTL)
	if err := s.repo.SaveRefreshToken(ctx, userID, hashToken(refreshToken), refreshExp, userAgent, sanitizeIP(ip)); err != nil {
		return dto.TokenPair{}, err
	}

	return dto.TokenPair{AccessToken: accessToken, AccessTokenExpiresAt: accessExp, RefreshToken: refreshToken, RefreshTokenExpiresAt: refreshExp}, nil
}

func (s *AuthService) createAccessToken(userID int64) (string, time.Time, error) {
	now := s.now()
	expiresAt := now.Add(s.accessTokenTTL)
	jti, err := generateTokenString(24)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate token id: %w", err)
	}

	claims := jwt.RegisteredClaims{
		Issuer:    accessTokenIssuer,
		Subject:   fmt.Sprintf("%d", userID),
		Audience:  []string{accessTokenAudience},
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        "acc_" + jti,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return tokenString, expiresAt, nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", entity.ValidationError{Message: "email is required"}
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", entity.ValidationError{Message: "invalid email"}
	}
	return email, nil
}

func normalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", entity.ValidationError{Message: "name is required"}
	}
	if len(name) > 100 {
		return "", entity.ValidationError{Message: "name must be at most 100 characters"}
	}
	return name, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return entity.ValidationError{Message: "password must be at least 8 characters"}
	}
	if len(password) > 72 {
		return entity.ValidationError{Message: "password must be at most 72 characters"}
	}
	return nil
}

func generateTokenString(numBytes int) (string, error) {
	buf := make([]byte, numBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func sanitizeIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	return parsed.String()
}

func refreshTokenMetadataMismatch(token entity.RefreshToken, currentUserAgent, currentIP string) bool {
	if storedUserAgent := strings.TrimSpace(token.UserAgent); storedUserAgent != "" && currentUserAgent != "" && storedUserAgent != currentUserAgent {
		return true
	}
	if token.IPAddress != nil && *token.IPAddress != "" && currentIP != "" && *token.IPAddress != currentIP {
		return true
	}
	return false
}

func (s *AuthService) detectRefreshTokenReuse(ctx context.Context, tokenHash string) {
	token, err := s.repo.FindRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return
	}
	if token.RevokedAt == nil || token.ReplacedByTokenHash == nil {
		return
	}
	if err := s.repo.RevokeAllUserRefreshTokens(ctx, token.UserID); err != nil {
		log.Printf("auth: failed to revoke all user refresh tokens after reuse detection: %v", err)
	}
}

func (s *AuthService) sendEmailVerification(ctx context.Context, to, token string, expiresAt time.Time) error {
	subject := "Verify your RecipeBox account"
	body := fmt.Sprintf(
		"Use this token to verify your email: %s\nExpires at: %s",
		token,
		expiresAt.UTC().Format(time.RFC3339),
	)
	if link := s.buildActionLink("/verify-email", token); link != "" {
		body = fmt.Sprintf(
			"Verify your RecipeBox account by opening this link:\n%s\n\nThis link expires at %s.",
			link,
			expiresAt.UTC().Format(time.RFC3339),
		)
	}
	return s.sendEmail(ctx, to, subject, body)
}

func (s *AuthService) sendPasswordReset(ctx context.Context, to, token string, expiresAt time.Time) error {
	subject := "Reset your RecipeBox password"
	body := fmt.Sprintf(
		"Use this token to reset your password: %s\nExpires at: %s",
		token,
		expiresAt.UTC().Format(time.RFC3339),
	)
	if link := s.buildActionLink("/reset-password", token); link != "" {
		body = fmt.Sprintf(
			"Reset your RecipeBox password by opening this link:\n%s\n\nThis link expires at %s.",
			link,
			expiresAt.UTC().Format(time.RFC3339),
		)
	}
	return s.sendEmail(ctx, to, subject, body)
}

func (s *AuthService) sendEmail(ctx context.Context, to, subject, body string) error {
	if s.emailSender == nil {
		if s.exposeTokens {
			return nil
		}
		return errors.New("email sender is not configured")
	}
	return s.emailSender.Send(ctx, to, subject, body)
}

func (s *AuthService) buildActionLink(path, token string) string {
	if s.frontendBaseURL == "" {
		return ""
	}
	return s.frontendBaseURL + path + "?token=" + url.QueryEscape(token)
}
