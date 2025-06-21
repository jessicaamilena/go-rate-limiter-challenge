package limiter

import (
	"context"
	"time"
)

type StorageStrategy interface {
	Increment(ctx context.Context, key string, window time.Duration) (int, error)
	SetBan(ctx context.Context, key string, duration time.Duration) error
	IsBanned(ctx context.Context, key string) (bool, error)
	GetBanReset(ctx context.Context, key string) (time.Duration, error)
	Close() error
}
