package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"recipebox-backend-go/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthGormRepository struct {
	db *gorm.DB
}

func NewAuthGormRepository(db *gorm.DB) *AuthGormRepository {
	return &AuthGormRepository{db: db}
}

func (r *AuthGormRepository) CreateUser(ctx context.Context, name, email, passwordHash string) (models.User, error) {
	user := models.User{Name: name, Email: email, PasswordHash: passwordHash}
	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return models.User{}, models.ErrConflict
		}
		return models.User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (r *AuthGormRepository) FindUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, models.ErrNotFound
		}
		return models.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *AuthGormRepository) FindUserByID(ctx context.Context, id int64) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, models.ErrNotFound
		}
		return models.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (r *AuthGormRepository) UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{
		"password_hash": passwordHash,
		"updated_at":    time.Now().UTC(),
	}).Error; err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

func (r *AuthGormRepository) SaveRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, userAgent, ip string) error {
	model := models.RefreshToken{UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt, UserAgent: userAgent, IPAddress: nullableString(ip)}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

func (r *AuthGormRepository) FindRefreshTokenOwner(ctx context.Context, tokenHash string, now time.Time) (int64, error) {
	var token models.RefreshToken
	if err := r.db.WithContext(ctx).Select("user_id").Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, now).Take(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, models.ErrNotFound
		}
		return 0, fmt.Errorf("find refresh token owner: %w", err)
	}
	return token.UserID, nil
}

func (r *AuthGormRepository) FindRefreshTokenByHash(ctx context.Context, tokenHash string) (models.RefreshToken, error) {
	var token models.RefreshToken
	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).Take(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.RefreshToken{}, models.ErrNotFound
		}
		return models.RefreshToken{}, fmt.Errorf("find refresh token by hash: %w", err)
	}
	return token, nil
}

func (r *AuthGormRepository) RotateRefreshToken(ctx context.Context, oldTokenHash, newTokenHash string, newExpiresAt, now time.Time, userAgent, ip string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var oldToken models.RefreshToken
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", oldTokenHash, now).Take(&oldToken).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.ErrNotFound
			}
			return fmt.Errorf("lock refresh token: %w", err)
		}

		replacedBy := newTokenHash
		if err := tx.Model(&models.RefreshToken{}).Where("id = ?", oldToken.ID).Updates(map[string]any{"revoked_at": now, "replaced_by_token_hash": replacedBy}).Error; err != nil {
			return fmt.Errorf("revoke old refresh token: %w", err)
		}

		newToken := models.RefreshToken{UserID: oldToken.UserID, TokenHash: newTokenHash, ExpiresAt: newExpiresAt, UserAgent: userAgent, IPAddress: nullableString(ip)}
		if err := tx.Create(&newToken).Error; err != nil {
			return fmt.Errorf("create new refresh token: %w", err)
		}
		return nil
	})
}

func (r *AuthGormRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("token_hash = ? AND revoked_at IS NULL", tokenHash).Update("revoked_at", now).Error; err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (r *AuthGormRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error {
	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).Model(&models.RefreshToken{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", now).Error; err != nil {
		return fmt.Errorf("revoke all user refresh tokens: %w", err)
	}
	return nil
}

func (r *AuthGormRepository) SaveEmailVerificationToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	model := models.EmailVerificationToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("save email verification token: %w", err)
	}
	return nil
}

func (r *AuthGormRepository) ConsumeEmailVerificationToken(ctx context.Context, tokenHash string, now time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var token models.EmailVerificationToken
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND consumed_at IS NULL AND expires_at > ?", tokenHash, now).
			Take(&token).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.ErrNotFound
			}
			return fmt.Errorf("find email verification token: %w", err)
		}

		if err := tx.Model(&models.EmailVerificationToken{}).
			Where("id = ?", token.ID).
			Update("consumed_at", now).Error; err != nil {
			return fmt.Errorf("consume email verification token: %w", err)
		}

		if err := tx.Model(&models.User{}).
			Where("id = ? AND email_verified_at IS NULL", token.UserID).
			Update("email_verified_at", now).Error; err != nil {
			return fmt.Errorf("mark email verified: %w", err)
		}
		return nil
	})
}

func (r *AuthGormRepository) SavePasswordResetToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	model := models.PasswordResetToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("save password reset token: %w", err)
	}
	return nil
}

func (r *AuthGormRepository) ConsumePasswordResetTokenAndUpdatePassword(ctx context.Context, tokenHash, newPasswordHash string, now time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var token models.PasswordResetToken
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND consumed_at IS NULL AND expires_at > ?", tokenHash, now).
			Take(&token).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.ErrNotFound
			}
			return fmt.Errorf("find password reset token: %w", err)
		}

		if err := tx.Model(&models.PasswordResetToken{}).
			Where("id = ?", token.ID).
			Update("consumed_at", now).Error; err != nil {
			return fmt.Errorf("consume password reset token: %w", err)
		}

		if err := tx.Model(&models.User{}).Where("id = ?", token.UserID).Updates(map[string]any{
			"password_hash": newPasswordHash,
			"updated_at":    now,
		}).Error; err != nil {
			return fmt.Errorf("update password by reset token: %w", err)
		}

		if err := tx.Model(&models.PasswordResetToken{}).
			Where("user_id = ? AND consumed_at IS NULL AND expires_at > ?", token.UserID, now).
			Update("consumed_at", now).Error; err != nil {
			return fmt.Errorf("consume remaining password reset tokens: %w", err)
		}

		if err := tx.Model(&models.RefreshToken{}).
			Where("user_id = ? AND revoked_at IS NULL", token.UserID).
			Update("revoked_at", now).Error; err != nil {
			return fmt.Errorf("revoke user refresh tokens: %w", err)
		}

		return nil
	})
}

func nullableString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
