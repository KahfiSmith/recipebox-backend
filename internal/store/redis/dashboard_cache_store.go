package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"recipebox-backend-go/internal/dto"
	"recipebox-backend-go/internal/service"
)

type DashboardCacheStore struct {
	redisClient *goredis.Client
	ttl         time.Duration
}

func NewDashboardCacheStore(redisClient *goredis.Client, ttl time.Duration) *DashboardCacheStore {
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	return &DashboardCacheStore{redisClient: redisClient, ttl: ttl}
}

var _ service.DashboardCacheStore = (*DashboardCacheStore)(nil)

func (s *DashboardCacheStore) GetDashboard(ctx context.Context, userID int64) (dto.DashboardResponse, bool, error) {
	raw, err := s.redisClient.Get(ctx, dashboardCacheKey(userID)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return dto.DashboardResponse{}, false, nil
		}
		return dto.DashboardResponse{}, false, err
	}

	var payload dto.DashboardResponse
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return dto.DashboardResponse{}, false, fmt.Errorf("decode dashboard cache: %w", err)
	}
	return payload, true, nil
}

func (s *DashboardCacheStore) SetDashboard(ctx context.Context, userID int64, payload dto.DashboardResponse) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode dashboard cache: %w", err)
	}
	return s.redisClient.Set(ctx, dashboardCacheKey(userID), string(encoded), s.ttl).Err()
}

func (s *DashboardCacheStore) InvalidateDashboard(ctx context.Context, userID int64) error {
	_, err := s.redisClient.Del(ctx, dashboardCacheKey(userID)).Result()
	return err
}

func dashboardCacheKey(userID int64) string {
	return "dashboard:overview:" + strconv.FormatInt(userID, 10)
}
