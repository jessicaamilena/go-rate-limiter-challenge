# Go Rate Limiter

A high-performance, configurable rate-limiting middleware for Gin-based HTTP servers. Supports multiple storage backends (Redis, Memcached, MySQL and PostgreSQL) with IP and token-based rate limiting.

This repository provides a `Makefile` with handy commands for building the project and managing the Docker services. Run `make help` to see all available targets.
> Before using the Makefile, make sure you have `make` and Docker installed.

## 🚀 Features

- **Dual Rate Limiting**: IP-based and token-based rate limiting with token precedence
- **Configurable Limits**: Different limits for different API tokens
- **Redis Backend**: Persistent, high-performance rate limiting using Redis
- **Memcached Support**: Switch storage backend via configuration
- **MySQL & Postgres Support**: Use relational databases as storage backends
- **Strategy Pattern**: Easy to swap storage backends
- **Ban Duration**: Configurable ban periods when limits are exceeded
- **Rate Limit Headers**: Standard HTTP rate limiting headers
- **Dockerized**: Docker Compose setup with Redis, Memcached, MySQL and PostgreSQL
- **Comprehensive Testing**: Postman collection with extensive test scenarios

## 📋 Quick Start

### Using Docker (Recommended)

When running the project for the first time, it's a good idea to rebuild the Docker services from scratch and start them in the background. This avoids interference from previous containers or volumes:

```bash
docker-compose down
docker-compose build --no-cache
docker-compose up -d # starts app and all databases

# View logs if you want to inspect the startup in detail
docker-compose logs -f

# Test the API
curl http://localhost:8080/ping
```

### Local Development

```bash
# Start a backend (redis, memcached, mysql or postgres)
make dev-redis   # or dev-memcached, dev-mysql, dev-postgres

# Set environment variables and run
RL_IP_LIMIT=5 \
RL_TOKEN_LIMIT_DEFAULT=10 \
RL_CUSTOM_TOKEN_LIMITS="abc123:20,premium:100" \
RL_BLOCK_DURATION_SECONDS=300 \
go run ./cmd/main.go
```

### Testing with Different Storage Backends

1. Start the required storage service using Docker Compose (all services run with `make docker-up`):

```bash
# Redis (default)
docker-compose up -d redis

# Memcached
docker-compose up -d memcached

# MySQL
docker-compose up -d mysql

# PostgreSQL
docker-compose up -d postgres
```
2. Run the application locally with the desired backend:

#### Redis
```bash
STORAGE_BACKEND=redis REDIS_URL=redis://localhost:6379/0 \
RL_IP_LIMIT=5 RL_TOKEN_LIMIT_DEFAULT=10 RL_BLOCK_DURATION_SECONDS=300 \
go run ./cmd/main.go
```

#### Memcached
```bash
STORAGE_BACKEND=memcached MEMCACHED_SERVER=localhost:11211 \
RL_IP_LIMIT=5 RL_TOKEN_LIMIT_DEFAULT=10 RL_BLOCK_DURATION_SECONDS=300 \
go run ./cmd/main.go
```

#### MySQL
```bash
STORAGE_BACKEND=mysql MYSQL_DSN=root:root@tcp(localhost:3306)/go_rate_limiter_db?parseTime=true \
RL_IP_LIMIT=5 RL_TOKEN_LIMIT_DEFAULT=10 RL_BLOCK_DURATION_SECONDS=300 \
go run ./cmd/main.go
```

#### PostgreSQL
```bash
STORAGE_BACKEND=postgres POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/go_rate_limiter_db?sslmode=disable \
RL_IP_LIMIT=5 RL_TOKEN_LIMIT_DEFAULT=10 RL_BLOCK_DURATION_SECONDS=300 \
go run ./cmd/main.go
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Gin Router    │────│  Rate Limiter   │────│   Redis Store   │
│   (Port 8080)   │    │   Middleware    │    │  (Port 6379)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
    ┌────▼────┐             ┌────▼────┐             ┌────▼────┐
    │  /ping  │             │ IP vs   │             │ Counter │
    └─────────┘             │ Token   │             │ & Ban   │
                            │ Logic   │             │ Keys    │
                            └─────────┘             └─────────┘
```

## 📁 Project Structure

```
rate-limiter-go/
├── cmd/
│   └── main.go                     # Application entry point
├── internal/
│   ├── config/                     # Configuration management
│   │   └── config.go
│   ├── limiter/                    # Core rate limiting logic
│   │   ├── limiter.go              # Rate limiter implementation
│   │   ├── storage.go              # Strategy interface
│   │   ├── storage_memcached.go    # Memcached persistence
│   │   ├── storage_mysql.go        # MySQL persistence
│   │   ├── storage_postgres.go     # PostgreSQL persistence
│   │   └── storage_redis.go        # Redis persistence
│   └── middleware/                 # Gin middleware
│       └── ratelimit.go
├── handlers/                       # HTTP handlers
│   └── ping.go
├── tests
│   ├── postman/                # Postman test collection
│   ├── Rate-Limiter-Tests.postman_collection.json
│   ├── Rate-Limiter-Environment.postman_environment.json
│   └── README.md   
├── docker-compose.yml          # Docker services
├── Dockerfile                  # Application container
├── Makefile                    # Build and run commands
└── README.md                   # This file
```

## ⚙️ Configuration

Configure the rate limiter using environment variables:

| Variable                    | Default | Description |
|-----------------------------|---------|-------------|
| `SERVER_PORT`               | `8080` | HTTP server port |
| `RL_IP_LIMIT`               | `10` | Max requests per second per IP |
| `RL_TOKEN_LIMIT_DEFAULT`    | `50` | Default token limit per second |
| `RL_CUSTOM_TOKEN_LIMITS`    | `""` | Custom token limits (`token:limit,token:limit`) |
| `RL_BLOCK_DURATION_SECONDS` | `300` | Ban duration in seconds |
| `STORAGE_BACKEND`           | `redis` | `redis`, `memcached`, `mysql`, or `postgres` |
| `REDIS_URL`                 | `redis://localhost:6379/0` | Redis connection string |
| `MEMCACHED_SERVER`          | `localhost:11211` | Comma-separated memcached servers |
| `MYSQL_DSN`                 | `root:root@tcp(mysql:3306)/go_rate_limiter_db` | MySQL DSN |
| `POSTGRES_DSN`              | `postgres://postgres:postgres@postgres:5432/go_rate_limiter_db?sslmode=disable` | PostgreSQL DSN |

### Example Configuration

```bash
# .env file
SERVER_PORT=8080
RL_IP_LIMIT=5
RL_TOKEN_LIMIT_DEFAULT=10
RL_CUSTOM_TOKEN_LIMITS="abc123:20,premium:100,enterprise:500"
RL_BLOCK_DURATION_SECONDS=300
REDIS_URL=redis://localhost:6379/0
STORE_BACKEND=redis
MEMCACHED_SERVER=localhost:11211
MYSQL_DSN=root:root@tcp(mysql:3306)/go_rate_limiter_db?parseTime=true
POSTGRES_DSN=postgres://postgres:postgres@postgres:5432/go_rate_limiter_db?sslmode=disable
```

```bash
# IDE environment mode - Do not 
SERVER_PORT=8080;RL_IP_LIMIT=5;RL_TOKEN_LIMIT_DEFAULT=10;RL_CUSTOM_TOKEN_LIMITS=abc123:20,premium:100,enterprise:500;RL_BLOCK_DURATION_SECONDS=300;REDIS_URL=redis://localhost:6379/0;STORE_BACKEND=redis;MEMCACHED_SERVER=localhost:11211;MYSQL_DSN=root:root@tcp(mysql:3306)/go_rate_limiter_db?parseTime=true;POSTGRES_DSN=postgres://postgres:postgres@postgres:5432/go_rate_limiter_db?sslmode=disable
```

## 🔧 Usage

### Basic Request (IP Rate Limiting)

```bash
curl http://localhost:8080/ping
# Response: pong
# Headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
```

### Token-Based Rate Limiting

```bash
# Using API_KEY header
curl -H "API_KEY: abc123" http://localhost:8080/ping

# Using Authorization header
curl -H "Authorization: Bearer abc123" http://localhost:8080/ping
```

### Rate Limit Headers

Every response includes rate limiting information:

```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 7
X-RateLimit-Reset: 1749922520
```

### Rate Limit Exceeded (429 Response)

```json
{
  "error": "Rate Limit Exceeded",
  "message": "You have reached the maximum number of requests or actions allowed within a certain time frame",
  "retry_after_seconds": 300
}
```

## 🧪 Testing

### Using Postman

1. Import the Postman collection from `postman` folder
2. Set up the environment variables
3. Run individual tests or the entire collection
4. See detailed test results and console logs

### Using Make Commands

```bash
# Test basic functionality
make test-ping

# Test rate limiting
make test-rate-limit

# Load testing
make docker-up  # starts app and all databases
# Then use Postman Runner or Newman
```

### Manual Testing

```bash
# Send multiple requests quickly to trigger rate limiting
for i in {1..15}; do
  echo "Request $i:"
  curl -s -w "Status: %{http_code}\n" http://localhost:8080/ping
done
```

## 🔀 Rate Limiting Logic

### Precedence Rules

1. **Token with Custom Limit**: If a valid API token with custom limit is provided
2. **Token with Default Limit**: If a valid API token without custom limit is provided  
3. **IP Rate Limiting**: Fallback to IP-based limiting

### Example Scenarios

```bash
# Scenario 1: No token (uses IP limit: 5 req/sec)
curl http://localhost:8080/ping

# Scenario 2: Default token (uses default limit: 10 req/sec)
curl -H "API_KEY: unknown-token" http://localhost:8080/ping

# Scenario 3: Custom token (uses custom limit: 20 req/sec)
curl -H "API_KEY: abc123" http://localhost:8080/ping

# Scenario 4: Premium token (uses premium limit: 100 req/sec)
curl -H "API_KEY: premium" http://localhost:8080/ping
```

## 📊 Monitoring & Observability

### Redis Keys Structure

```
# Rate limiting counters
ip:127.0.0.1:1749922520          # IP-based counter with timestamp
token:616263313233:1749922520    # Token-based counter (hashed)

# Ban keys
ban:ip:127.0.0.1                 # IP ban key
ban:token:616263313233           # Token ban key
```

### Monitoring Commands

```bash
# Watch Redis operations
redis-cli monitor

# Check current keys
redis-cli keys "*"

# Check specific key
redis-cli get "ip:127.0.0.1:$(date +%s)"

# Check TTL
redis-cli ttl "ban:ip:127.0.0.1"
```

## 🐳 Docker Commands

```bash
# Build and start app and all databases
make docker-up 

# View logs
make docker-logs

# Stop services
make docker-down

# Rebuild images
make docker-build
```

## 🚀 Production Considerations

### Performance Optimizations

- **Redis Pipelining**: Implemented for atomic operations
- **Connection Pooling**: Redis client uses connection pooling
- **Efficient Key Structure**: Time-based bucketing for automatic cleanup

### Security

- **Token Hashing**: API tokens are hashed in Redis keys for privacy
- **Token Masking**: Tokens are masked in logs
- **Non-root Container**: Docker runs as non-root user

### Scalability

- **Horizontal Scaling**: Multiple app instances can share Redis
- **Redis Clustering**: Can use Redis Cluster for high availability
- **Load Balancing**: Supports X-Forwarded-For headers

## 🔧 Development

### Building

```bash
# Build binary
make build

# Clean up
make clean
```

### Adding New Storage Backend

1. Implement the `StorageStrategy` interface in `internal/limiter/storage.go`
2. Add constructor function
3. Update configuration to support new backend
4. Add tests

Example:
```go
type MemcachedStorage struct {
    client *memcache.Client
}

func (m *MemcachedStorage) Increment(ctx context.Context, key string, window time.Duration) (int, error) {
    // Implementation
}
```
