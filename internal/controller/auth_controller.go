package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"recipebox-backend-go/internal/dto"
	"recipebox-backend-go/internal/models"
	"recipebox-backend-go/internal/middleware"
	"recipebox-backend-go/internal/service"
	"recipebox-backend-go/internal/utils"
)

type AuthController struct {
	service             *service.AuthService
	refreshCookieSecure bool
	refreshCookieTTL    time.Duration
	trustedProxies      []*net.IPNet
}

const refreshTokenCookieName = "refresh_token"

func NewAuthController(service *service.AuthService, refreshCookieSecure bool, refreshCookieTTL time.Duration, trustedProxies []*net.IPNet) *AuthController {
	return &AuthController{
		service:             service,
		refreshCookieSecure: refreshCookieSecure,
		refreshCookieTTL:    refreshCookieTTL,
		trustedProxies:      trustedProxies,
	}
}

// Register godoc
// @Summary Register user
// @Description Register a new user account.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.RegisterRequest true "Register payload"
// @Success 201 {object} dto.RegisterEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var input dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Register(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEmailTaken):
			utils.Error(w, http.StatusConflict, err.Error())
		case models.IsValidationError(err):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]any{"data": resp})
}

// Login godoc
// @Summary Login user
// @Description Login with email and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.LoginRequest true "Login payload"
// @Success 200 {object} dto.AuthEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Login(r.Context(), input, r.UserAgent(), extractRequestIP(r, h.trustedProxies))
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidCredentials):
			utils.Error(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, models.ErrEmailNotVerified):
			utils.Error(w, http.StatusForbidden, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	setRefreshTokenCookie(w, resp.Tokens.RefreshToken, h.refreshCookieTTL, h.refreshCookieSecure)
	resp.Tokens.RefreshToken = ""
	resp.Tokens.RefreshTokenExpiresAt = time.Time{}
	utils.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

// Refresh godoc
// @Summary Refresh access token
// @Description Refresh access token using refresh token (cookie or request body).
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.RefreshRequest false "Refresh payload"
// @Success 200 {object} dto.TokenEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := refreshTokenFromRequest(r)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Refresh(r.Context(), refreshToken, r.UserAgent(), extractRequestIP(r, h.trustedProxies))
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidRefreshToken):
			utils.Error(w, http.StatusUnauthorized, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	setRefreshTokenCookie(w, resp.RefreshToken, h.refreshCookieTTL, h.refreshCookieSecure)
	resp.RefreshToken = ""
	resp.RefreshTokenExpiresAt = time.Time{}
	utils.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

// Logout godoc
// @Summary Logout
// @Description Revoke refresh token and clear cookie.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.RefreshRequest false "Refresh payload"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := refreshTokenFromRequest(r)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			utils.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	if err := h.service.Logout(r.Context(), refreshToken, accessTokenFromAuthorizationHeader(r)); err != nil {
		utils.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	clearRefreshTokenCookie(w, h.refreshCookieSecure)
	utils.JSON(w, http.StatusOK, map[string]any{"message": "logged out"})
}

// Me godoc
// @Summary Get current user profile
// @Description Return current authenticated user profile.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserEnvelope
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthController) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "user not found")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"data": user})
}

// RequestEmailVerification godoc
// @Summary Request email verification token
// @Description Request email verification flow.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.EmailRequest true "Email payload"
// @Success 200 {object} dto.OneTimeTokenEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/verify-email/request [post]
func (h *AuthController) RequestEmailVerification(w http.ResponseWriter, r *http.Request) {
	var input dto.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.RequestEmailVerification(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			utils.JSON(w, http.StatusOK, map[string]any{"message": "if the email exists, verification instructions have been generated"})
		case models.IsValidationError(err):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeOneTimeTokenMessage(w, "if the email exists, verification instructions have been generated", resp)
}

// VerifyEmail godoc
// @Summary Confirm email verification
// @Description Verify email using one-time token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.VerifyEmailRequest true "Verify email payload"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/verify-email/confirm [post]
func (h *AuthController) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var input dto.VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.VerifyEmail(r.Context(), input.Token); err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidVerifyToken):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"message": "email verified"})
}

// ForgotPassword godoc
// @Summary Request password reset token
// @Description Request reset password flow.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.EmailRequest true "Email payload"
// @Success 200 {object} dto.OneTimeTokenEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/password/forgot [post]
func (h *AuthController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input dto.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.RequestPasswordReset(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			utils.JSON(w, http.StatusOK, map[string]any{"message": "if the email exists, reset instructions have been generated"})
		case models.IsValidationError(err):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeOneTimeTokenMessage(w, "if the email exists, reset instructions have been generated", resp)
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset password using one-time token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.ResetPasswordRequest true "Reset password payload"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/password/reset [post]
func (h *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input dto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.ResetPassword(r.Context(), input.Token, input.NewPassword); err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidResetToken), models.IsValidationError(err):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"message": "password has been reset"})
}

func extractRequestIP(r *http.Request, trustedProxies []*net.IPNet) string {
	return utils.ClientIP(r, trustedProxies)
}

func refreshTokenFromRequest(r *http.Request) (string, error) {
	if c, err := r.Cookie(refreshTokenCookieName); err == nil {
		if v := strings.TrimSpace(c.Value); v != "" {
			return v, nil
		}
	}

	var input dto.RefreshRequest
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	return strings.TrimSpace(input.RefreshToken), nil
}

func setRefreshTokenCookie(w http.ResponseWriter, refreshToken string, ttl time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	})
}

func clearRefreshTokenCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func writeOneTimeTokenMessage(w http.ResponseWriter, message string, resp dto.OneTimeTokenResponse) {
	payload := map[string]any{"message": message}
	if resp.Token != "" {
		payload["data"] = resp
	}
	utils.JSON(w, http.StatusOK, payload)
}

func accessTokenFromAuthorizationHeader(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
}
