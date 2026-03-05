package service

import (
	"context"

	"recipebox-backend-go/internal/dto"
)

type DashboardCacheStore interface {
	GetDashboard(ctx context.Context, userID int64) (dto.DashboardResponse, bool, error)
	SetDashboard(ctx context.Context, userID int64, payload dto.DashboardResponse) error
	InvalidateDashboard(ctx context.Context, userID int64) error
}

type noopDashboardCacheStore struct{}

func NewNoopDashboardCacheStore() DashboardCacheStore {
	return noopDashboardCacheStore{}
}

func (noopDashboardCacheStore) GetDashboard(context.Context, int64) (dto.DashboardResponse, bool, error) {
	return dto.DashboardResponse{}, false, nil
}

func (noopDashboardCacheStore) SetDashboard(context.Context, int64, dto.DashboardResponse) error {
	return nil
}

func (noopDashboardCacheStore) InvalidateDashboard(context.Context, int64) error {
	return nil
}
