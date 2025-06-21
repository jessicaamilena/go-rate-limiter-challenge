package limiter

import (
	"context"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"time"
)

type MemcachedStorage struct {
	client *memcache.Client
}

func NewMemcachedStorage(servers []string) (*MemcachedStorage, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no memcached servers provided")
	}

	client := memcache.New(servers...)
	if err := client.Set(&memcache.Item{Key: "__ping__", Value: []byte("1"), Expiration: 1}); err != nil {
		return nil, fmt.Errorf("failed connecting to memcached: %w", err)
	}
	return &MemcachedStorage{client: client}, nil
}

func (m *MemcachedStorage) Increment(ctx context.Context, key string, window time.Duration) (int, error) {
	newVal, err := m.client.Increment(key, 1)
	if err == memcache.ErrCacheMiss {
		item := &memcache.Item{
			Key:        key,
			Value:      []byte("1"),
			Expiration: int32(window.Seconds()),
		}
		if addErr := m.client.Add(item); addErr != nil {
			if addErr == memcache.ErrNotStored {
				newVal, err = m.client.Increment(key, 1)
				if err != nil {
					return 0, fmt.Errorf("increment failed after add: %w", err)
				}
				return int(newVal), nil
			}
			return 0, fmt.Errorf("failed adding key: %w", addErr)
		}
		return 1, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed incrementing key: %w", err)
	}
	_ = m.client.Touch(key, int32(window.Seconds()))
	return int(newVal), nil
}

func (m *MemcachedStorage) SetBan(ctx context.Context, key string, duration time.Duration) error {
	item := &memcache.Item{
		Key:        key,
		Value:      []byte("banned"),
		Expiration: int32(duration.Seconds()),
	}
	if err := m.client.Set(item); err != nil {
		return fmt.Errorf("failed set ban: %w", err)
	}
	return nil
}

func (m *MemcachedStorage) IsBanned(ctx context.Context, key string) (bool, error) {
	_, err := m.client.Get(key)
	if err == memcache.ErrCacheMiss {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed checking ban status: %w", err)
	}
	return true, nil
}

func (m *MemcachedStorage) Close() error {
	return nil
}
