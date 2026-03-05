package service

import (
	"context"
	"time"
)

// AuthStateStore provides fast auth-state storage for short-lived data.
type AuthStateStore interface {
	StoreAccessSession(ctx context.Context, tokenID string, userID int64, ttl time.Duration) error
	HasAccessSession(ctx context.Context, tokenID string, userID int64) (bool, error)
	RevokeAccessSession(ctx context.Context, tokenID string) error

	BlacklistAccessToken(ctx context.Context, tokenID string, ttl time.Duration) error
	IsAccessTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)

	StoreRefreshToken(ctx context.Context, tokenHash string, userID int64, ttl time.Duration) error
	HasRefreshToken(ctx context.Context, tokenHash string, userID int64) (bool, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error

	StoreOTP(ctx context.Context, purpose, tokenHash string, ttl time.Duration) error
	ConsumeOTP(ctx context.Context, purpose, tokenHash string) (bool, error)
}
