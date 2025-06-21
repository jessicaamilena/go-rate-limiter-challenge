package limiter

import (
	"context"
	"fmt"
	"github.com/jessicaamilena/go-rate-limiter-challenge/internal/config"
	"strings"
	"time"
)

type Result struct {
	Allowed   bool
	Reason    string
	ResetTime time.Time
	Limit     int
	Remaining int
}

type RateLimiter struct {
	config  *config.Config
	storage StorageStrategy
	window  time.Duration
}

func NewRateLimiter(cfg *config.Config, storage StorageStrategy) *RateLimiter {
	return &RateLimiter{
		config:  cfg,
		storage: storage,
		window:  time.Second,
	}
}

func (rl *RateLimiter) Check(ctx context.Context, ip, token string) (*Result, error) {
	var (
		key   string
		limit int
		id    string
	)

	if token != "" {
		if customLimit, exists := rl.config.CustomTokenLimit[token]; exists {
			key = fmt.Sprintf("token:%s", hashToken(token))
			limit = customLimit
			id = fmt.Sprintf("token:%s", maskToken(token))
		} else {
			key = fmt.Sprintf("token:%s", hashToken(token))
			limit = rl.config.TokenLimitDefault
			id = fmt.Sprintf("token:%s", maskToken(token))
		}
	} else {
		key = fmt.Sprintf("ip:%s", ip)
		limit = rl.config.TokenLimitDefault
		id = fmt.Sprintf("ip:%s", ip)
	}

	windowKey := fmt.Sprintf("%s:%d", key, time.Now().Unix())
	banKey := fmt.Sprintf("ban:%s", key)

	banned, err := rl.storage.IsBanned(ctx, banKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check ban status: %w", err)
	}

	if banned {
		return &Result{
			Allowed:   false,
			Reason:    fmt.Sprintf("You have reached the maximum number of requests or actions allowed within a certain time frame"),
			ResetTime: time.Now().Add(rl.config.BlockDuration),
			Limit:     limit,
			Remaining: 0,
		}, nil
	}

	count, err := rl.storage.Increment(ctx, windowKey, rl.window)
	if err != nil {
		return nil, fmt.Errorf("failed to increment rate counter: %w", err)
	}

	if count > limit {
		if err := rl.storage.SetBan(ctx, banKey, rl.config.BlockDuration); err != nil {
			fmt.Printf("failed to set ban: %s: %v\n", id, err)
		}

		return &Result{
			Allowed:   false,
			Reason:    fmt.Sprintf("You have reached the maximum number of requests or actions allowed within a certain time frame"),
			ResetTime: time.Now().Add(rl.config.BlockDuration),
			Limit:     limit,
			Remaining: 0,
		}, nil
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return &Result{
		Allowed:   true,
		Reason:    fmt.Sprintf("Request allowed for %s (%d/%d requests)", id, count, limit),
		ResetTime: time.Now().Add(rl.window),
		Limit:     limit,
		Remaining: remaining,
	}, nil

}

func (rl *RateLimiter) Close() error {
	return rl.storage.Close()
}

func hashToken(token string) string {
	if len(token) < 8 {
		return token
	}
	return fmt.Sprintf("%x", []byte(token)[:8])
}

func maskToken(token string) string {
	if len(token) <= 4 {
		return strings.Repeat("*", len(token))
	}
	return token[:2] + strings.Repeat("*", len(token)-4) + token[len(token)-2:]
}
