package service

import (
	"context"
	"time"
)

type noopAuthStateStore struct{}

func NewNoopAuthStateStore() AuthStateStore {
	return noopAuthStateStore{}
}

func (noopAuthStateStore) StoreAccessSession(context.Context, string, int64, time.Duration) error {
	return nil
}

func (noopAuthStateStore) HasAccessSession(context.Context, string, int64) (bool, error) {
	return true, nil
}

func (noopAuthStateStore) RevokeAccessSession(context.Context, string) error {
	return nil
}

func (noopAuthStateStore) BlacklistAccessToken(context.Context, string, time.Duration) error {
	return nil
}

func (noopAuthStateStore) IsAccessTokenBlacklisted(context.Context, string) (bool, error) {
	return false, nil
}

func (noopAuthStateStore) StoreRefreshToken(context.Context, string, int64, time.Duration) error {
	return nil
}

func (noopAuthStateStore) HasRefreshToken(context.Context, string, int64) (bool, error) {
	return true, nil
}

func (noopAuthStateStore) RevokeRefreshToken(context.Context, string) error {
	return nil
}

func (noopAuthStateStore) StoreOTP(context.Context, string, string, time.Duration) error {
	return nil
}

func (noopAuthStateStore) ConsumeOTP(context.Context, string, string) (bool, error) {
	return false, nil
}
