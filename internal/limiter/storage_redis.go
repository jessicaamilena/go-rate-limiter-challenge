package limiter

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(redisURL string) (*RedisStorage, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed parsing Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed connecting to Redis: %w", err)
	}
	log.Printf("[Redis] Using Redis storage strategy")
	return &RedisStorage{client: client}, nil
}

func (r *RedisStorage) Increment(ctx context.Context, key string, window time.Duration) (int, error) {
	pipe := r.client.Pipeline()

	incrCmd := pipe.Incr(ctx, key)

	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed incrementing counter: %w", err)
	}

	count, err := incrCmd.Result()
	if err != nil {
		return 0, fmt.Errorf("failed getting increment result: %w", err)
	}
	return int(count), nil
}

func (r *RedisStorage) SetBan(ctx context.Context, key string, duration time.Duration) error {
	err := r.client.Set(ctx, key, "banned", duration).Err()
	if err != nil {
		return fmt.Errorf("failed set ban: %w", err)
	}
	return nil
}

func (r *RedisStorage) IsBanned(ctx context.Context, key string) (bool, error) {
	result := r.client.Get(ctx, key)
	if errors.Is(result.Err(), redis.Nil) {
		return false, nil
	}
	if result.Err() != nil {
		return false, fmt.Errorf("failed checking ban status: %w", result.Err())
	}
	return true, nil
}

func (r *RedisStorage) GetBanReset(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed getting ttl: %w", err)
	}
	if ttl < 0 {
		return 0, nil
	}
	return ttl, nil
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}
