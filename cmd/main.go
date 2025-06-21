package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jessicaamilena/go-rate-limiter-challenge/handlers"
	"github.com/jessicaamilena/go-rate-limiter-challenge/internal/config"
	"github.com/jessicaamilena/go-rate-limiter-challenge/internal/limiter"
	"github.com/jessicaamilena/go-rate-limiter-challenge/internal/middleware"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	var storage limiter.StorageStrategy
	switch cfg.StorageBackend {
	case "redis":
		storage, err = limiter.NewRedisStorage(cfg.RedisURL)
	case "memcached":
		server := strings.Split(cfg.MemcachedServer, ",")
		for i := range server {
			server[i] = strings.TrimSpace(server[i])
		}
		storage, err = limiter.NewMemcachedStorage(server)
	case "mysql":
		storage, err = limiter.NewMySQLStorage(cfg.MySQLDSN)
	case "postgres":
		storage, err = limiter.NewPostgresStorage(cfg.PostgresDSN)
	default:
		err = fmt.Errorf("unknown storage backend: %s", cfg.StorageBackend)
	}
	if err != nil {
		log.Fatalf("failed initializing storage: %v", err)
	}
	defer storage.Close()

	rateLimiter := limiter.NewRateLimiter(cfg, storage)
	defer rateLimiter.Close()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(middleware.RateLimitMiddleware(rateLimiter))

	pingHandler := handlers.NewPingHandler()

	router.GET("/ping", pingHandler.Ping)

	fmt.Printf("Rate Limiter Service starting on port %s\n", cfg.ServerPort)
	fmt.Printf("Configuration:\n")
	fmt.Printf("   - IP Limit: %d requests/second\n", cfg.IPLimit)
	fmt.Printf("   - Token Default Limit: %d requests/second\n", cfg.TokenLimitDefault)
	fmt.Printf("   - Custom Token Limits: %d tokens configured\n", len(cfg.CustomTokenLimit))
	fmt.Printf("   - Block Duration: %v\n", cfg.BlockDuration)
	fmt.Printf("   - Storage Backend: %s\n", cfg.StorageBackend)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := router.Run(":" + cfg.ServerPort); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-quit
	fmt.Println("Shutting down server...")
	fmt.Println("Server stopped gracefully")
}
