# Go Rate Limiter - Postman Test Suite

This folder contains a comprehensive Postman collection for testing the Go Rate Limiter API with multiple storage backends (Redis, Memcached, MySQL and PostgreSQL).

## 📁 Files

- `Rate-Limiter-Tests.postman_collection.json` - Main test collection
- `Rate-Limiter-Environment.postman_environment.json` - Environment variables
- `README.md` - This documentation

## 🚀 Quick Start

### 1. Import into Postman

1. Open Postman
2. Click **Import** button
3. Drag and drop both JSON files or click **Upload Files**
4. Select both files:
   - `Rate-Limiter-Tests.postman_collection.json`
   - `Rate-Limiter-Environment.postman_environment.json`
5. Click **Import**

### 2. Set Environment

1. In Postman, select the **Rate Limiter - Local Development** environment from the dropdown (top right)
2. Verify the environment variables are correct for your setup

### 3. Start the Service

Make sure your rate limiter service is running:

```bash
# Option 1: Using Docker Compose
make docker-up

# Option 2: Local development (requires Redis running)
RL_IP_LIMIT=5 RL_TOKEN_LIMIT_DEFAULT=10 RL_CUSTOM_TOKEN_LIMITS="abc123:20,premium:100" go run ./cmd/main.go
```

## 📋 Test Collection Structure

### 1. **Basic Endpoints**
- **Ping Endpoint**: Tests the `/ping` endpoint functionality

### 2. **IP Rate Limiting Tests**
- **Rapid Requests**: Tests IP-based rate limiting by sending multiple requests

### 3. **Token-Based Rate Limiting**
- **Default Token Limit Test**: Tests tokens using the default limit (10 req/sec)
- **Custom Token Limit Test**: Tests the `abc123` token with custom limit (20 req/sec)
- **Premium Token Limit Test**: Tests the `premium` token with high limit (100 req/sec)
- **Authorization Header Test**: Tests Bearer token authentication

### 4. **Load Testing Scenarios**
- **Burst Test**: Designed to be run multiple times to trigger rate limits
- **Token vs IP Precedence**: Verifies token limits take precedence over IP limits

### 5. **Ban Duration Tests**
- **Check If Banned**: Tests the ban functionality
- **Test Different IP**: Simulates different clients to test isolation

## 🧪 Running Tests

### Individual Tests
1. Select any request from the collection
2. Click **Send**
3. Check the **Test Results** tab to see assertions
4. View the **Console** (bottom panel) for detailed logs

### Collection Runner
1. Right-click on the collection name
2. Select **Run collection**
3. Configure iterations and delay between requests
4. Click **Run Rate Limiter API Tests**

### Automated Testing
For burst testing and rate limit validation:
1. Select the **Burst Test** request
2. Use **Runner** with multiple iterations (e.g., 10 iterations, 100ms delay)
3. Watch the console output to see when rate limits are triggered

## 🔧 Environment Variables

| Variable | Value | Description |
|----------|-------|-------------|
| `base_url` | `http://localhost:8080` | API base URL |
| `custom_token` | `abc123` | Token with 20 req/sec limit |
| `premium_token` | `premium` | Token with 100 req/sec limit |
| `default_token` | `test123` | Token using default limit |
| `invalid_token` | `invalid-token-xyz` | Token without custom limits |

## 📊 Test Scenarios

### Scenario 1: Basic Functionality
```
1. Run "Ping Endpoint" ✅
```

### Scenario 2: IP Rate Limiting
```
1. Run "Rapid Requests" multiple times quickly
2. Observe rate limit headers decrease
3. Eventually get 429 response with ban
```

### Scenario 3: Token Precedence
```
1. Get rate limited without token (IP limit: 5)
2. Use premium token and get higher limit (100)
3. Verify token limits override IP limits
```

### Scenario 4: Different Token Limits
```
1. Test default_token → 10 req/sec limit
2. Test custom_token → 20 req/sec limit
3. Test premium_token → 100 req/sec limit
```

## 🎯 Expected Behaviors

### ✅ Successful Requests
- Status: `200 OK`
- Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`
- Body: `"pong"` (for ping) or JSON (for health)

### 🚫 Rate Limited Requests
- Status: `429 Too Many Requests`
- Headers: `Retry-After`
- Body: JSON with error details and retry information

### Key Assertions
- Rate limit headers are always present
- Token limits take precedence over IP limits
- Ban duration is enforced
- Different tokens have different limits

## 🐛 Troubleshooting

### Common Issues

1. **Connection Refused**
   - Make sure the service is running on port 8080
   - Check if your required storage is running

2. **No Rate Limiting**
   - Verify environment variables are set correctly
   - Check storage connection
   - Look at application logs

3. **Wrong Limits Applied**
   - Verify token configuration in environment variables
   - Check the service configuration

### Debug Commands
```bash
# Check Redis
redis-cli ping

# View service logs
docker-compose logs app

# Check rate limit configuration
echo $RL_CUSTOM_TOKEN_LIMITS
```

## 📈 Advanced Testing

### Load Testing with Newman (CLI)
Install Newman: `npm install -g newman`

```bash
# Run entire collection
newman run Rate-Limiter-Tests.postman_collection.json \
  -e Rate-Limiter-Environment.postman_environment.json

# Run with multiple iterations
newman run Rate-Limiter-Tests.postman_collection.json \
  -e Rate-Limiter-Environment.postman_environment.json \
  -n 10 --delay-request 100
```

### Monitoring
- Watch Redis keys: `redis-cli monitor`
- Check rate limit headers in responses
- Monitor application logs for ban events
- Use Postman Console for detailed test output

## 🔍 Test Coverage

- ✅ Basic endpoint functionality
- ✅ IP-based rate limiting
- ✅ Token-based rate limiting
- ✅ Custom token limits
- ✅ Token precedence over IP
- ✅ Ban functionality
- ✅ Rate limit headers validation
- ✅ Error response format
- ✅ Bearer token authentication
- ✅ Multiple authentication methods

## 📝 Notes

- Tests include comprehensive assertions and logging
- Console output provides detailed information about each test
- Environment can be easily modified for different configurations
- Collection supports both manual testing and automated runs
- All rate limit scenarios from the original design document are covered 