package dto

import (
	"recipebox-backend-go/internal/models"
	"time"
)

type RegisterRequest struct {
	Name     string `json:"name" example:"Kahfi Smith"`
	Email    string `json:"email" example:"alkahfii2018@gmail.com"`
	Password string `json:"password" example:"secret123"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"alkahfii2018@gmail.com"`
	Password string `json:"password" example:"secret123"`
}

type EmailRequest struct {
	Email string `json:"email" example:"alkahfii2018@gmail.com"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" example:"refresh-token-sample"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" example:"verify-token-sample"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" example:"reset-token-sample"`
	NewPassword string `json:"newPassword" example:"newSecret123"`
}

type TokenPair struct {
	AccessToken           string    `json:"accessToken"`
	AccessTokenExpiresAt  time.Time `json:"accessTokenExpiresAt"`
	RefreshToken          string    `json:"refreshToken"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
}

type AuthResponse struct {
	User   models.User `json:"user"`
	Tokens TokenPair   `json:"tokens"`
}

type RegisterResponse struct {
	User models.User `json:"user"`
}

type OneTimeTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}
