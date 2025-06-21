# Go Rate Limiter with Gin & Redis

A high-performance, configurable rate-limiting middleware for Gin-based HTTP servers. Features IP and token-based rate limiting with Redis persistence and strategy pattern for easy backend swapping.

## 🚀 Features

- **Dual Rate Limiting**: IP-based and token-based rate limiting with token precedence
- **Configurable Limits**: Different limits for different API tokens
- **Redis Backend**: Persistent, high-performance rate limiting using Redis
- **Memcached Support**: Switch storage backend via configuration
- **MySQL & Postgres Support**: Use relational databases as storage backends
- **Strategy Pattern**: Easy to swap storage backends
- **Ban Duration**: Configurable ban periods when limits are exceeded
- **Rate Limit Headers**: Standard HTTP rate limiting headers
- **Dockerized**: Full Docker Compose setup with Redis
- **Comprehensive Testing**: Postman collection with extensive test scenarios

## 📋 Quick Start

### Using Docker (Recommended)

```bash
# Start the services
make docker-up

# Test the API
curl http://localhost:8080/ping
```

### Local Development

```bash
# Start Redis
make dev-redis

# Set environment variables and run
RL_IP_LIMIT=5 \
RL_TOKEN_LIMIT_DEFAULT=10 \
RL_CUSTOM_TOKEN_LIMITS="abc123:20,premium:100" \
RL_BLOCK_DURATION_SECONDS=60 \
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
    │ /ping   │             │ IP vs   │             │ Counter │
    │ /health │             │ Token   │             │ & Ban   │
    └─────────┘             │ Logic   │             │ Keys    │
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

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `RL_IP_LIMIT` | `10` | Max requests per second per IP |
| `RL_TOKEN_LIMIT_DEFAULT` | `50` | Default token limit per second |
| `RL_CUSTOM_TOKEN_LIMITS` | `""` | Custom token limits (`token:limit,token:limit`) |
| `RL_BLOCK_DURATION_SECONDS` | `300` | Ban duration in seconds |
| `STORE_BACKEND` | `redis` | `redis`, `memcached`, `mysql`, or `postgres` |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection string |
| `MEMCACHED_SERVER` | `localhost:11211` | Comma-separated memcached servers |
| `MYSQL_DSN` | `root:root@tcp(mysql:3306)/ratelimiter` | MySQL DSN |
| `POSTGRES_DSN` | `postgres://postgres:postgres@postgres:5432/ratelimiter?sslmode=disable` | PostgreSQL DSN |

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
MYSQL_DSN=root:root@tcp(mysql:3306)/ratelimiter
POSTGRES_DSN=postgres://postgres:postgres@postgres:5432/ratelimiter?sslmode=disable
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
  "error": "Rate limit exceeded",
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
make test-health

# Test rate limiting
make test-rate-limit

# Load testing
make docker-up
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
# Build and start services
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

# Run locally
make run

# Run tests
make test

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

## 🔗 References

- [Gin Web Framework](https://gin-gonic.com/)
- [Redis](https://redis.io/)
- [Rate Limiting Best Practices](https://www.figma.com/blog/an-alternative-approach-to-rate-limiting/)
- [HTTP Rate Limiting Headers](https://tools.ietf.org/id/draft-polli-ratelimit-headers-00.html) 