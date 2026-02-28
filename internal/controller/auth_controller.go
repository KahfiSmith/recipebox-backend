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
	"recipebox-backend-go/internal/entity"
	"recipebox-backend-go/internal/middleware"
	"recipebox-backend-go/internal/service"
	"recipebox-backend-go/internal/utils"
)

type AuthController struct {
	service             *service.AuthService
	refreshCookieSecure bool
	refreshCookieTTL    time.Duration
}

const refreshTokenCookieName = "refresh_token"

func NewAuthController(service *service.AuthService, refreshCookieSecure bool, refreshCookieTTL time.Duration) *AuthController {
	return &AuthController{
		service:             service,
		refreshCookieSecure: refreshCookieSecure,
		refreshCookieTTL:    refreshCookieTTL,
	}
}

func (h *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var input dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Register(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrEmailTaken):
			utils.Error(w, http.StatusConflict, err.Error())
		case entity.IsValidationError(err):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]any{"data": resp})
}

func (h *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Login(r.Context(), input, r.UserAgent(), extractIP(r.RemoteAddr))
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidCredentials):
			utils.Error(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, entity.ErrEmailNotVerified):
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

func (h *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := refreshTokenFromRequest(r)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Refresh(r.Context(), refreshToken, r.UserAgent(), extractIP(r.RemoteAddr))
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidRefreshToken):
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

func (h *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := refreshTokenFromRequest(r)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.Logout(r.Context(), refreshToken); err != nil {
		utils.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	clearRefreshTokenCookie(w, h.refreshCookieSecure)
	utils.JSON(w, http.StatusOK, map[string]any{"message": "logged out"})
}

func (h *AuthController) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "user not found")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"data": user})
}

func (h *AuthController) RequestEmailVerification(w http.ResponseWriter, r *http.Request) {
	var input dto.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.RequestEmailVerification(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			utils.JSON(w, http.StatusOK, map[string]any{"message": "if the email exists, verification instructions have been generated"})
		case entity.IsValidationError(err):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if resp.ExpiresAt.IsZero() {
		utils.JSON(w, http.StatusOK, map[string]any{"message": "email is already verified"})
		return
	}

	payload := map[string]any{"message": "verification instructions sent"}
	if resp.Token != "" {
		payload["data"] = resp
	}
	utils.JSON(w, http.StatusOK, payload)
}

func (h *AuthController) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var input dto.VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.VerifyEmail(r.Context(), input.Token); err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidVerifyToken):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"message": "email verified"})
}

func (h *AuthController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input dto.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.RequestPasswordReset(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			utils.JSON(w, http.StatusOK, map[string]any{"message": "if the email exists, reset instructions have been generated"})
		case entity.IsValidationError(err):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	payload := map[string]any{"message": "if the email exists, reset instructions have been generated"}
	if resp.Token != "" {
		payload["data"] = resp
	}
	utils.JSON(w, http.StatusOK, payload)
}

func (h *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input dto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.ResetPassword(r.Context(), input.Token, input.NewPassword); err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidResetToken), entity.IsValidationError(err):
			utils.Error(w, http.StatusBadRequest, err.Error())
		default:
			utils.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"message": "password has been reset"})
}

func extractIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return ""
	}
	return host
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
