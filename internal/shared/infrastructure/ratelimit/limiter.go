package ratelimit

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var fixedWindowScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
	redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
return {count, ttl}
`)

type Limiter struct {
	client *redis.Client
}

func New(client *redis.Client) *Limiter {
	return &Limiter{client: client}
}

func (l *Limiter) Middleware(name string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ipHash := sha256.Sum256([]byte(c.ClientIP()))
		key := "rate-limit:" + name + ":" + hex.EncodeToString(ipHash[:])

		result, err := fixedWindowScript.Run(
			c.Request.Context(),
			l.client,
			[]string{key},
			window.Milliseconds(),
		).Slice()
		if err != nil {
			slog.ErrorContext(c.Request.Context(), "rate limiter unavailable", slog.String("error.message", err.Error()), slog.String("rate_limit.name", name))
			response.Error(c, "service unavailable", http.StatusServiceUnavailable, nil)
			return
		}

		count := result[0].(int64)
		ttl := result[1].(int64)
		if count > int64(limit) {
			retryAfterSeconds := (ttl + 999) / 1000
			if retryAfterSeconds < 1 {
				retryAfterSeconds = 1
			}
			c.Header("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
			response.Error(c, "too many requests", http.StatusTooManyRequests, nil)
			return
		}

		c.Next()
	}
}
