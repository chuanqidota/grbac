package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig holds the configuration for a rate limiter.
type RateLimitConfig struct {
	// KeyFunc returns the rate limit key from the request context.
	KeyFunc func(c *gin.Context) string
	// MaxRequests is the maximum number of requests allowed in the window.
	MaxRequests int
	// Window is the sliding window duration.
	Window time.Duration
}

// RateLimitMiddleware returns a Gin middleware that enforces a sliding-window
// rate limit using Redis sorted sets.
func RateLimitMiddleware(rdb *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ratelimit:" + cfg.KeyFunc(c)
		now := time.Now().UnixMicro()
		windowStart := now - cfg.Window.Microseconds()

		pipe := rdb.Pipeline()
		// Remove expired entries.
		pipe.ZRemRangeByScore(c.Request.Context(), key, "0",
			strconv.FormatInt(windowStart, 10))
		// Count current entries.
		countCmd := pipe.ZCard(c.Request.Context(), key)
		// Add current request.
		pipe.ZAdd(c.Request.Context(), key, redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%d", now),
		})
		// Set TTL so keys don't accumulate forever.
		pipe.Expire(c.Request.Context(), key, cfg.Window+time.Minute)
		pipe.Exec(c.Request.Context())

		if countCmd.Val() > int64(cfg.MaxRequests) {
			c.Header("Retry-After", cfg.Window.String())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后重试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimit returns a rate limiter for the login endpoint:
// 10 requests per minute per IP.
func LoginRateLimit(rdb *redis.Client) gin.HandlerFunc {
	return RateLimitMiddleware(rdb, RateLimitConfig{
		KeyFunc: func(c *gin.Context) string {
			return "login:" + c.ClientIP()
		},
		MaxRequests: 10,
		Window:      time.Minute,
	})
}

// ExternalRateLimit returns a rate limiter for external API endpoints:
// 100 requests per second per system code.
func ExternalRateLimit(rdb *redis.Client) gin.HandlerFunc {
	return RateLimitMiddleware(rdb, RateLimitConfig{
		KeyFunc: func(c *gin.Context) string {
			code := c.GetHeader("X-System-Code")
			if code == "" {
				code = c.ClientIP()
			}
			return "external:" + code
		},
		MaxRequests: 100,
		Window:      time.Second,
	})
}
