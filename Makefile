.PHONY: help build run test clean docker-build docker-up docker-down docker-logs deps

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the Go application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-up    - Start services with docker-compose"
	@echo "  docker-down  - Stop services with docker-compose"
	@echo "  docker-logs  - View logs from docker-compose"

# Build the application
build:
	go build -o bin/rate-limiter ./cmd/main.go

# Run the application locally (requires Redis to be running)
run:
	go run ./cmd/main.go

# Run tests
test:
	go test -v ./...

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
	@echo "Services started. App available at http://localhost:8080"
	@echo "Redis available at localhost:6379"

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Development with local Redis
dev-redis:
	docker run -d --name redis-dev -p 6379:6379 redis:7-alpine

dev-redis-stop:
	docker stop redis-dev && docker rm redis-dev

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