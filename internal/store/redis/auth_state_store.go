package redisstore

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"recipebox-backend-go/internal/service"
)

type AuthStateStore struct {
	redisClient *goredis.Client
}

func NewAuthStateStore(redisClient *goredis.Client) *AuthStateStore {
	return &AuthStateStore{redisClient: redisClient}
}

var _ service.AuthStateStore = (*AuthStateStore)(nil)

func (s *AuthStateStore) StoreAccessSession(ctx context.Context, tokenID string, userID int64, ttl time.Duration) error {
	return s.redisClient.Set(ctx, accessSessionKey(tokenID), strconv.FormatInt(userID, 10), ttl).Err()
}

func (s *AuthStateStore) HasAccessSession(ctx context.Context, tokenID string, userID int64) (bool, error) {
	value, err := s.redisClient.Get(ctx, accessSessionKey(tokenID)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return false, nil
		}
		return false, err
	}
	storedUserID, convErr := strconv.ParseInt(value, 10, 64)
	if convErr != nil {
		return false, fmt.Errorf("parse access session user id: %w", convErr)
	}
	return storedUserID == userID, nil
}

func (s *AuthStateStore) RevokeAccessSession(ctx context.Context, tokenID string) error {
	_, err := s.redisClient.Del(ctx, accessSessionKey(tokenID)).Result()
	return err
}

func (s *AuthStateStore) BlacklistAccessToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return s.redisClient.Set(ctx, accessBlacklistKey(tokenID), "1", ttl).Err()
}

func (s *AuthStateStore) IsAccessTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	count, err := s.redisClient.Exists(ctx, accessBlacklistKey(tokenID)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *AuthStateStore) StoreRefreshToken(ctx context.Context, tokenHash string, userID int64, ttl time.Duration) error {
	return s.redisClient.Set(ctx, refreshTokenKey(tokenHash), strconv.FormatInt(userID, 10), ttl).Err()
}

func (s *AuthStateStore) HasRefreshToken(ctx context.Context, tokenHash string, userID int64) (bool, error) {
	value, err := s.redisClient.Get(ctx, refreshTokenKey(tokenHash)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return false, nil
		}
		return false, err
	}
	storedUserID, convErr := strconv.ParseInt(value, 10, 64)
	if convErr != nil {
		return false, fmt.Errorf("parse refresh token user id: %w", convErr)
	}
	return storedUserID == userID, nil
}

func (s *AuthStateStore) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := s.redisClient.Del(ctx, refreshTokenKey(tokenHash)).Result()
	return err
}

func (s *AuthStateStore) StoreOTP(ctx context.Context, purpose, tokenHash string, ttl time.Duration) error {
	return s.redisClient.Set(ctx, otpKey(purpose, tokenHash), "1", ttl).Err()
}

func (s *AuthStateStore) ConsumeOTP(ctx context.Context, purpose, tokenHash string) (bool, error) {
	deleted, err := s.redisClient.Del(ctx, otpKey(purpose, tokenHash)).Result()
	if err != nil {
		return false, err
	}
	return deleted > 0, nil
}

func accessSessionKey(tokenID string) string {
	return "auth:sess:access:" + tokenID
}

func accessBlacklistKey(tokenID string) string {
	return "auth:blacklist:access:" + tokenID
}

func refreshTokenKey(tokenHash string) string {
	return "auth:refresh:" + tokenHash
}

func otpKey(purpose, tokenHash string) string {
	return "auth:otp:" + strings.TrimSpace(purpose) + ":" + tokenHash
}
