package limiter

import (
	"context"
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"strconv"
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
	expiresAt := time.Now().Add(duration).Unix()
	item := &memcache.Item{
		Key:        key,
		Value:      []byte(strconv.FormatInt(expiresAt, 10)),
		Expiration: int32(duration.Seconds()),
	}
	if err := m.client.Set(item); err != nil {
		return fmt.Errorf("failed set ban: %w", err)
	}
	return nil
}

func (m *MemcachedStorage) IsBanned(ctx context.Context, key string) (bool, error) {
	item, err := m.client.Get(key)
	if err == memcache.ErrCacheMiss {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed checking ban status: %w", err)
	}
	ts, parseErr := strconv.ParseInt(string(item.Value), 10, 64)
	if parseErr != nil {
		return false, fmt.Errorf("failed parsing ban timestamp: %w", parseErr)
	}
	if time.Now().After(time.Unix(ts, 0)) {
		return false, nil
	}
	log.Printf("[Memcached] Using Memcached storage strategy")
	return true, nil
}

func (m *MemcachedStorage) GetBanReset(ctx context.Context, key string) (time.Duration, error) {
	item, err := m.client.Get(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed checking ban reset: %w", err)
	}
	ts, parseErr := strconv.ParseInt(string(item.Value), 10, 64)
	if parseErr != nil {
		return 0, fmt.Errorf("failed parsing ban timestamp: %w", parseErr)
	}
	ttl := time.Until(time.Unix(ts, 0))
	if ttl < 0 {
		return 0, nil
	}
	return ttl, nil
}

func (m *MemcachedStorage) Close() error {
	return nil
}
