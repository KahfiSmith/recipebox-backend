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
	"recipebox-backend-go/internal/models"
	"recipebox-backend-go/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo             repository.AuthRepository
	stateRepo        AuthStateStore
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
		stateRepo:        NewNoopAuthStateStore(),
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

func (s *AuthService) ConfigureAuthStateStore(stateRepo AuthStateStore) {
	if stateRepo == nil {
		s.stateRepo = NewNoopAuthStateStore()
		return
	}
	s.stateRepo = stateRepo
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
		if errors.Is(err, models.ErrConflict) {
			return dto.RegisterResponse{}, models.ErrEmailTaken
		}
		return dto.RegisterResponse{}, err
	}

	var verificationResp *dto.OneTimeTokenResponse
	if issuedResp, err := s.issueEmailVerificationToken(ctx, user.ID, user.Email); err != nil {
		log.Printf("auth: failed to prepare email verification for newly registered user %s: %v", user.Email, err)
	} else if issuedResp.Token != "" {
		verificationResp = &issuedResp
	}

	user.PasswordHash = ""
	return dto.RegisterResponse{User: user, EmailVerification: verificationResp}, nil
}

func (s *AuthService) Login(ctx context.Context, input dto.LoginRequest, userAgent, ip string) (dto.AuthResponse, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return dto.AuthResponse{}, models.ErrInvalidCredentials
		}
		return dto.AuthResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return dto.AuthResponse{}, models.ErrInvalidCredentials
	}
	if user.EmailVerifiedAt == nil {
		return dto.AuthResponse{}, models.ErrEmailNotVerified
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
		return dto.TokenPair{}, models.ErrInvalidRefreshToken
	}

	oldHash := hashToken(refreshToken)
	now := s.now()
	userID, err := s.repo.FindRefreshTokenOwner(ctx, oldHash, now)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			s.detectRefreshTokenReuse(ctx, oldHash)
			return dto.TokenPair{}, models.ErrInvalidRefreshToken
		}
		return dto.TokenPair{}, err
	}

	currentUserAgent := strings.TrimSpace(userAgent)
	currentIP := sanitizeIP(ip)
	storedToken, err := s.repo.FindRefreshTokenByHash(ctx, oldHash)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return dto.TokenPair{}, models.ErrInvalidRefreshToken
		}
		return dto.TokenPair{}, err
	}
	if refreshTokenMetadataMismatch(storedToken, currentUserAgent, currentIP) {
		if err := s.repo.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
			return dto.TokenPair{}, err
		}
		return dto.TokenPair{}, models.ErrInvalidRefreshToken
	}
	isRefreshActive, err := s.stateRepo.HasRefreshToken(ctx, oldHash, userID)
	if err != nil {
		return dto.TokenPair{}, err
	}
	if !isRefreshActive {
		return dto.TokenPair{}, models.ErrInvalidRefreshToken
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return dto.TokenPair{}, models.ErrInvalidRefreshToken
		}
		return dto.TokenPair{}, err
	}

	accessToken, accessExp, accessTokenID, err := s.createAccessToken(user.ID)
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
		if errors.Is(err, models.ErrNotFound) {
			return dto.TokenPair{}, models.ErrInvalidRefreshToken
		}
		return dto.TokenPair{}, err
	}
	if err := s.stateRepo.RevokeRefreshToken(ctx, oldHash); err != nil {
		return dto.TokenPair{}, err
	}
	if err := s.stateRepo.StoreRefreshToken(ctx, newHash, userID, time.Until(refreshExp)); err != nil {
		return dto.TokenPair{}, err
	}
	if err := s.stateRepo.StoreAccessSession(ctx, accessTokenID, userID, time.Until(accessExp)); err != nil {
		return dto.TokenPair{}, err
	}

	return dto.TokenPair{AccessToken: accessToken, AccessTokenExpiresAt: accessExp, RefreshToken: newRefreshToken, RefreshTokenExpiresAt: refreshExp}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken, accessToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	accessToken = strings.TrimSpace(accessToken)

	if refreshToken != "" {
		refreshHash := hashToken(refreshToken)
		if err := s.repo.RevokeRefreshToken(ctx, refreshHash); err != nil {
			return err
		}
		if err := s.stateRepo.RevokeRefreshToken(ctx, refreshHash); err != nil {
			return err
		}
	}

	if accessToken != "" {
		_, tokenID, expiresAt, err := s.parseAccessTokenClaims(accessToken)
		if err == nil {
			if err := s.stateRepo.RevokeAccessSession(ctx, tokenID); err != nil {
				return err
			}
			if err := s.stateRepo.BlacklistAccessToken(ctx, tokenID, time.Until(expiresAt)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *AuthService) GetMe(ctx context.Context, userID int64) (models.User, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return models.User{}, err
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
		if errors.Is(err, models.ErrNotFound) {
			return dto.OneTimeTokenResponse{}, models.ErrNotFound
		}
		return dto.OneTimeTokenResponse{}, err
	}

	if user.EmailVerifiedAt != nil {
		return dto.OneTimeTokenResponse{}, nil
	}

	return s.issueEmailVerificationToken(ctx, user.ID, user.Email)
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return models.ErrInvalidVerifyToken
	}
	tokenHash := hashToken(token)
	if _, err := s.stateRepo.ConsumeOTP(ctx, "verify_email", tokenHash); err != nil {
		return err
	}
	if err := s.repo.ConsumeEmailVerificationToken(ctx, tokenHash, s.now()); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.ErrInvalidVerifyToken
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
		if errors.Is(err, models.ErrNotFound) {
			return dto.OneTimeTokenResponse{}, models.ErrNotFound
		}
		return dto.OneTimeTokenResponse{}, err
	}

	rawToken, err := generateNumericCode(8)
	if err != nil {
		return dto.OneTimeTokenResponse{}, fmt.Errorf("generate password reset token: %w", err)
	}

	expiresAt := s.now().Add(s.resetPasswordTTL)
	tokenHash := hashToken(rawToken)
	if err := s.repo.SavePasswordResetToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return dto.OneTimeTokenResponse{}, err
	}
	if err := s.stateRepo.StoreOTP(ctx, "password_reset", tokenHash, time.Until(expiresAt)); err != nil {
		return dto.OneTimeTokenResponse{}, err
	}

	s.trySendPasswordReset(ctx, user.Email, rawToken, expiresAt)
	if !s.exposeTokens {
		rawToken = ""
	}

	return dto.OneTimeTokenResponse{Token: rawToken, ExpiresAt: expiresAt}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return models.ErrInvalidResetToken
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.bcryptCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	tokenHash := hashToken(token)
	if _, err := s.stateRepo.ConsumeOTP(ctx, "password_reset", tokenHash); err != nil {
		return err
	}
	if err := s.repo.ConsumePasswordResetTokenAndUpdatePassword(ctx, tokenHash, string(newPasswordHash), s.now()); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.ErrInvalidResetToken
		}
		return err
	}

	return nil
}

func (s *AuthService) ParseAccessToken(ctx context.Context, tokenString string) (int64, error) {
	userID, tokenID, _, err := s.parseAccessTokenClaims(tokenString)
	if err != nil {
		return 0, err
	}
	isBlacklisted, err := s.stateRepo.IsAccessTokenBlacklisted(ctx, tokenID)
	if err != nil {
		return 0, err
	}
	if isBlacklisted {
		return 0, errors.New("access token is blacklisted")
	}
	hasSession, err := s.stateRepo.HasAccessSession(ctx, tokenID, userID)
	if err != nil {
		return 0, err
	}
	if !hasSession {
		return 0, errors.New("access session not found")
	}

	return userID, nil
}

func (s *AuthService) issueTokens(ctx context.Context, userID int64, userAgent, ip string) (dto.TokenPair, error) {
	accessToken, accessExp, accessTokenID, err := s.createAccessToken(userID)
	if err != nil {
		return dto.TokenPair{}, err
	}

	refreshToken, err := generateTokenString(48)
	if err != nil {
		return dto.TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshExp := s.now().Add(s.refreshTokenTTL)
	refreshHash := hashToken(refreshToken)
	if err := s.repo.SaveRefreshToken(ctx, userID, refreshHash, refreshExp, userAgent, sanitizeIP(ip)); err != nil {
		return dto.TokenPair{}, err
	}
	if err := s.stateRepo.StoreRefreshToken(ctx, refreshHash, userID, time.Until(refreshExp)); err != nil {
		return dto.TokenPair{}, err
	}
	if err := s.stateRepo.StoreAccessSession(ctx, accessTokenID, userID, time.Until(accessExp)); err != nil {
		return dto.TokenPair{}, err
	}

	return dto.TokenPair{AccessToken: accessToken, AccessTokenExpiresAt: accessExp, RefreshToken: refreshToken, RefreshTokenExpiresAt: refreshExp}, nil
}

func (s *AuthService) createAccessToken(userID int64) (string, time.Time, string, error) {
	now := s.now()
	expiresAt := now.Add(s.accessTokenTTL)
	jti, err := generateTokenString(24)
	if err != nil {
		return "", time.Time{}, "", fmt.Errorf("generate token id: %w", err)
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
		return "", time.Time{}, "", fmt.Errorf("sign access token: %w", err)
	}
	return tokenString, expiresAt, claims.ID, nil
}

func (s *AuthService) parseAccessTokenClaims(tokenString string) (int64, string, time.Time, error) {
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
		return 0, "", time.Time{}, fmt.Errorf("parse access token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, "", time.Time{}, errors.New("invalid access token")
	}
	if claims.Subject == "" {
		return 0, "", time.Time{}, errors.New("missing subject")
	}
	var userID int64
	if _, err := fmt.Sscan(claims.Subject, &userID); err != nil {
		return 0, "", time.Time{}, errors.New("invalid subject")
	}
	if claims.ID == "" || !strings.HasPrefix(claims.ID, "acc_") {
		return 0, "", time.Time{}, errors.New("invalid token id")
	}
	if claims.ExpiresAt == nil {
		return 0, "", time.Time{}, errors.New("missing expiry")
	}

	return userID, claims.ID, claims.ExpiresAt.Time, nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", models.ValidationError{Message: "email is required"}
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", models.ValidationError{Message: "invalid email"}
	}
	return email, nil
}

func normalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", models.ValidationError{Message: "name is required"}
	}
	if len(name) > 100 {
		return "", models.ValidationError{Message: "name must be at most 100 characters"}
	}
	return name, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return models.ValidationError{Message: "password must be at least 8 characters"}
	}
	if len(password) > 72 {
		return models.ValidationError{Message: "password must be at most 72 characters"}
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

func generateNumericCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("numeric code length must be positive")
	}

	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	code := make([]byte, length)
	for i, b := range buf {
		code[i] = '0' + (b % 10)
	}

	return string(code), nil
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

func refreshTokenMetadataMismatch(token models.RefreshToken, currentUserAgent, currentIP string) bool {
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

func (s *AuthService) issueEmailVerificationToken(ctx context.Context, userID int64, email string) (dto.OneTimeTokenResponse, error) {
	rawToken, err := generateNumericCode(8)
	if err != nil {
		return dto.OneTimeTokenResponse{}, fmt.Errorf("generate verify token: %w", err)
	}

	expiresAt := s.now().Add(s.verifyEmailTTL)
	tokenHash := hashToken(rawToken)
	if err := s.repo.SaveEmailVerificationToken(ctx, userID, tokenHash, expiresAt); err != nil {
		return dto.OneTimeTokenResponse{}, err
	}
	if err := s.stateRepo.StoreOTP(ctx, "verify_email", tokenHash, time.Until(expiresAt)); err != nil {
		return dto.OneTimeTokenResponse{}, err
	}

	s.trySendEmailVerification(ctx, email, rawToken, expiresAt)
	if !s.exposeTokens {
		rawToken = ""
	}

	return dto.OneTimeTokenResponse{Token: rawToken, ExpiresAt: expiresAt}, nil
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
			"Use this verification code: %s\n\nOr verify your RecipeBox account by opening this link:\n%s\n\nThis code and link expire at %s.",
			token,
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
	if link := s.buildActionLink("/auth/reset-password", token); link != "" {
		body = fmt.Sprintf(
			"Use this reset code: %s\n\nOr reset your RecipeBox password by opening this link:\n%s\n\nThis code and link expire at %s.",
			token,
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

func (s *AuthService) trySendEmailVerification(ctx context.Context, to, token string, expiresAt time.Time) {
	if err := s.sendEmailVerification(ctx, to, token, expiresAt); err != nil {
		log.Printf("auth: failed to send email verification to %s: %v", to, err)
	}
}

func (s *AuthService) trySendPasswordReset(ctx context.Context, to, token string, expiresAt time.Time) {
	if err := s.sendPasswordReset(ctx, to, token, expiresAt); err != nil {
		log.Printf("auth: failed to send password reset to %s: %v", to, err)
	}
}

func (s *AuthService) buildActionLink(path, token string) string {
	if s.frontendBaseURL == "" {
		return ""
	}
	return s.frontendBaseURL + path + "?token=" + url.QueryEscape(token)
}
