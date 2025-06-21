package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerPort        string
	IPLimit           int
	TokenLimitDefault int
	CustomTokenLimit  map[string]int
	BlockDurationSec  int
	BlockDuration     time.Duration

	// Storage config
	StorageBackend  string
	RedisURL        string
	MemcachedServer string
	MySQLDSN        string
	PostgresDSN     string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		ServerPort:        getEnvWithDefault("SERVER_PORT", "8080"),
		IPLimit:           getEnvAsIntWithDefault("RL_IP_LIMIT", 10),
		TokenLimitDefault: getEnvAsIntWithDefault("RL_TOKEN_LIMIT_DEFAULT", 50),
		BlockDurationSec:  getEnvAsIntWithDefault("RL_BLOCK_DURATION_SECONDS", 60),
		StorageBackend:    getEnvWithDefault("STORAGE_BACKEND", "redis"),
		RedisURL:          getEnvWithDefault("REDIS_URL", "redis://localhost:6379/0"),
		MemcachedServer:   getEnvWithDefault("MEMCACHED_SERVER", "localhost:11211"),
		MySQLDSN:          getEnvWithDefault("MYSQL_DSN", "root:root@tcp(mysql:3306)/go_rate_limiter_db"),
		PostgresDSN:       getEnvWithDefault("POSTGRES_DSN", "postgres://postgres:postgres@postgres:5432/go_rate_limiter_db?sslmode=disable"),
	}

	cfg.BlockDuration = time.Duration(cfg.BlockDurationSec) * time.Second

	var err error
	cfg.CustomTokenLimit, err = parseCustomTokenLimit(os.Getenv("RL_CUSTOM_TOKEN_LIMITS"))
	if err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.IPLimit <= 0 {
		return fmt.Errorf("IP limit must be positive, got %d", c.IPLimit)
	}
	if c.TokenLimitDefault <= 0 {
		return fmt.Errorf("token limit default must be positive, got %d", c.TokenLimitDefault)
	}
	if c.BlockDurationSec <= 0 {
		return fmt.Errorf("block duration seconds must be positive, got %d", c.BlockDurationSec)
	}
	if c.StorageBackend == "" {
		return fmt.Errorf("storage backend is required")
	}
	if c.StorageBackend != "redis" && c.StorageBackend != "memcached" && c.StorageBackend != "mysql" && c.StorageBackend != "postgres" {
		return fmt.Errorf("unknown storage backend: %s", c.StorageBackend)
	}
	return nil
}

// parseCustomTokenLimits parses comma-separated token:limit pairs
// Example: "abc123:100,xyz999:200"
func parseCustomTokenLimit(envValue string) (map[string]int, error) {
	result := make(map[string]int)

	if envValue == "" {
		return result, nil
	}

	pairs := strings.Split(envValue, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid custom token limit format: %s (expected toke:limit)", pair)
		}

		token := strings.TrimSpace(parts[0])
		limitStr := strings.TrimSpace(parts[1])
		if token == "" {
			return nil, fmt.Errorf("empty token: %s", pair)
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid limit value '%s' for token '%s': %w", limitStr, token, err)
		}
		if limit <= 0 {
			return nil, fmt.Errorf("limit must be positive for token '%s', got %d", token, limit)
		}

		result[token] = limit
	}
	return result, nil
}

func getEnvWithDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsIntWithDefault(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
