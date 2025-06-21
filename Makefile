.PHONY: help build run clean docker-build docker-up docker-down docker-logs deps \
	dev-redis dev-redis-stop dev-memcached dev-memcached-stop \
    dev-mysql dev-mysql-stop dev-postgres dev-postgres-stop \
    test-ping test-rate-limit

# Default target
help:
	@echo "Available commands:"
	@echo "  build        	 		 - Build the Go application"
	@echo "  run          	 		 - Run the application locally"
	@echo "  clean        	 		 - Clean build artifacts"
	@echo "  deps         	 		 - Download dependencies"
	@echo "  docker-build 	 		 - Build Docker image"
	@echo "  docker-up    	 		 - Start services with docker-compose (all backends)"
	@echo "  docker-down  	 		 - Stop services with docker-compose"
	@echo "  docker-compose logs -f	 - View logs from docker-compose"
	@echo "  dev-redis    	 		 - Run a local Redis instance for development"
	@echo "  dev-redis-stop  		 - Stop the local Redis instance"
	@echo "  dev-memcached           - Run a local Memcached instance for development"
	@echo "  dev-memcached-stop      - Stop the local Memcached instance"
	@echo "  dev-mysql               - Run a local MySQL instance for development"
	@echo "  dev-mysql-stop          - Stop the local MySQL instance"
	@echo "  dev-postgres            - Run a local PostgreSQL instance for development"
	@echo "  dev-postgres-stop       - Stop the local PostgreSQL instance"
	@echo "  test-ping    	 		 - Test the /ping endpoint"
	@echo "  test-rate-limit 		 - Send 15 quick requests to test rate limiting"

# Build the application
build:
	go build -o bin/rate-limiter ./cmd/main.go

# Run the application locally (requires Redis to be running)
run:
	go run ./cmd/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	docker-compose down --volumes --remove-orphans

# Download dependencies
deps:
	go mod download
	go mod tidy

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

	docker-compose ps
	@echo "App available at http://localhost:8080"
	@echo "Redis: localhost:6379 | Memcached: localhost:11211 | MySQL: localhost:3306 | Postgres: localhost:5432"

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Development with local Redis
dev-redis:
	docker run -d --name redis-dev -p 6379:6379 redis:7-alpine

dev-redis-stop:
	docker stop redis-dev && docker rm redis-dev

# Development with local Memcached
dev-memcached:
	docker run -d --name memcached-dev -p 11211:11211 memcached:alpine

dev-memcached-stop:
	docker stop memcached-dev && docker rm memcached-dev

# Development with local MySQL
dev-mysql:
	docker run -d --name mysql-dev -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=go_rate_limiter_db -p 3306:3306 mysql:8

dev-mysql-stop:
	docker stop mysql-dev && docker rm mysql-dev

# Development with local PostgreSQL
dev-postgres:
	docker run -d --name postgres-dev -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=go_rate_limiter_db -p 5432:5432 postgres:16-alpine

dev-postgres-stop:
	docker stop postgres-dev && docker rm postgres-dev

# Test the rate limiter
test-ping:
	@echo "Testing /ping endpoint..."
	curl -v http://localhost:8080/ping

test-rate-limit:
	@echo "Testing rate limiting (sending 15 requests quickly)..."
	@for i in $$(seq 1 15); do \
		echo "Request $$i:"; \
		curl -s -w "Status: %{http_code}\n" http://localhost:8080/ping; \
	done 