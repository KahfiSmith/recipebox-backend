package models

import (
	"errors"
	"time"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrEmailTaken          = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailNotVerified    = errors.New("email is not verified")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidVerifyToken  = errors.New("invalid email verification token")
	ErrInvalidResetToken   = errors.New("invalid password reset token")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}

type User struct {
	ID              int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name            string     `json:"name" gorm:"column:name;type:text;not null"`
	Email           string     `json:"email" gorm:"column:email;type:text;not null;uniqueIndex"`
	PasswordHash    string     `json:"-" gorm:"column:password_hash;type:text;not null"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty" gorm:"column:email_verified_at"`
	CreatedAt       time.Time  `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time  `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}

type RefreshToken struct {
	ID                  int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID              int64      `gorm:"column:user_id;not null;index:idx_refresh_tokens_user_id"`
	TokenHash           string     `gorm:"column:token_hash;type:char(64);not null;uniqueIndex"`
	ExpiresAt           time.Time  `gorm:"column:expires_at;not null;index:idx_refresh_tokens_expires_at"`
	UserAgent           string     `gorm:"column:user_agent;type:text"`
	IPAddress           *string    `gorm:"column:ip_address;type:inet"`
	RevokedAt           *time.Time `gorm:"column:revoked_at"`
	ReplacedByTokenHash *string    `gorm:"column:replaced_by_token_hash;type:char(64)"`
	CreatedAt           time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

type EmailVerificationToken struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID     int64      `gorm:"column:user_id;not null;index:idx_email_verification_tokens_user_id"`
	TokenHash  string     `gorm:"column:token_hash;type:char(64);not null;uniqueIndex"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;not null;index:idx_email_verification_tokens_expires_at"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}

type PasswordResetToken struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID     int64      `gorm:"column:user_id;not null;index:idx_password_reset_tokens_user_id"`
	TokenHash  string     `gorm:"column:token_hash;type:char(64);not null;uniqueIndex"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;not null;index:idx_password_reset_tokens_expires_at"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
