package entity

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
	ID              int64      `json:"id" gorm:"column:id;primaryKey"`
	Name            string     `json:"name" gorm:"column:name"`
	Email           string     `json:"email" gorm:"column:email"`
	PasswordHash    string     `json:"-" gorm:"column:password_hash"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty" gorm:"column:email_verified_at"`
	CreatedAt       time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt       time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}

type RefreshToken struct {
	ID                  int64      `gorm:"column:id;primaryKey"`
	UserID              int64      `gorm:"column:user_id"`
	TokenHash           string     `gorm:"column:token_hash"`
	ExpiresAt           time.Time  `gorm:"column:expires_at"`
	UserAgent           string     `gorm:"column:user_agent"`
	IPAddress           *string    `gorm:"column:ip_address"`
	RevokedAt           *time.Time `gorm:"column:revoked_at"`
	ReplacedByTokenHash *string    `gorm:"column:replaced_by_token_hash"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

type EmailVerificationToken struct {
	ID         int64      `gorm:"column:id;primaryKey"`
	UserID     int64      `gorm:"column:user_id"`
	TokenHash  string     `gorm:"column:token_hash"`
	ExpiresAt  time.Time  `gorm:"column:expires_at"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}

type PasswordResetToken struct {
	ID         int64      `gorm:"column:id;primaryKey"`
	UserID     int64      `gorm:"column:user_id"`
	TokenHash  string     `gorm:"column:token_hash"`
	ExpiresAt  time.Time  `gorm:"column:expires_at"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
