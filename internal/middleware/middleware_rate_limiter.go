package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jessicaamilena/go-rate-limiter-challenge/internal/limiter"
	"net/http"
	"strconv"
	"time"
)

func RateLimitMiddleware(rateLimiter *limiter.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		apiToken := c.GetHeader("API_KEY")
		if apiToken == "" {
			apiToken = c.GetHeader("Authorization")
			if len(apiToken) > 7 && apiToken[:7] == "Bearer " {
				apiToken = apiToken[7:]
			}
		}

		result, err := rateLimiter.Check(c.Request.Context(), clientIP, apiToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Internal Server Error",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

		if !result.Allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(result.ResetTime.Sub(time.Now()).Seconds()), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               "Rate Limit Exceeded",
				"message":             result.Reason,
				"retry_after_seconds": uint64(result.ResetTime.Sub(time.Now()).Seconds()),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
