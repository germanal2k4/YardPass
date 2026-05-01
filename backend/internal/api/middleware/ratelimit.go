package middleware

import (
	"net/http"
	"strconv"
	"time"

	"yardpass/internal/errors"
	"yardpass/internal/redis"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func InMemoryRateLimit(requestsPerSecond int, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burst)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			errors.ErrorResponseJSON(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}

func CreatePassRateLimit(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			key := "rate_limit:create_pass:" + c.ClientIP()
			allowed, err := redisClient.CheckRateLimit(c.Request.Context(), key, limit, window)
			if err != nil || !allowed {
				if !allowed {
					errors.ErrorResponseJSON(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many pass creation requests")
					c.Abort()
				}
				return
			}
			c.Next()
			return
		}

		key := "rate_limit:create_pass:" + strconv.FormatInt(userID.(int64), 10)
		allowed, err := redisClient.CheckRateLimit(c.Request.Context(), key, limit, window)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			errors.ErrorResponseJSON(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many pass creation requests")
			c.Abort()
			return
		}

		c.Next()
	}
}
