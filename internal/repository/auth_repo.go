package repository

import (
	"context"
	"recipebox-backend-go/internal/models"
	"time"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, name, email, passwordHash string) (models.User, error)
	FindUserByEmail(ctx context.Context, email string) (models.User, error)
	FindUserByID(ctx context.Context, id int64) (models.User, error)
	UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error
	SaveRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, userAgent, ip string) error
	FindRefreshTokenOwner(ctx context.Context, tokenHash string, now time.Time) (int64, error)
	FindRefreshTokenByHash(ctx context.Context, tokenHash string) (models.RefreshToken, error)
	RotateRefreshToken(ctx context.Context, oldTokenHash, newTokenHash string, newExpiresAt, now time.Time, userAgent, ip string) error
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error
	SaveEmailVerificationToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	ConsumeEmailVerificationToken(ctx context.Context, tokenHash string, now time.Time) error
	SavePasswordResetToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	ConsumePasswordResetTokenAndUpdatePassword(ctx context.Context, tokenHash, newPasswordHash string, now time.Time) error
}
