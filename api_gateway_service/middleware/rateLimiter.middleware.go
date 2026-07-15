package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

// rateLimitScript atomically increments a fixed-window counter and sets
// its expiry only on the first hit in that window — avoids a separate
// GET-then-EXPIRE round trip and the race that comes with it. Same
// atomic-Lua pattern as cache/refreshToken.go's rotateScript.
var rateLimitScript = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
if tonumber(current) == 1 then
	redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return current
`)

// RedisRateLimiter enforces a fixed-window request limit per caller,
// keyed by authenticated user ID when available (parsed best-effort
// from the access_token cookie, no error if missing/invalid) or by
// client IP otherwise. Redis-backed so the limit holds correctly
// across multiple gateway instances.
func RedisRateLimiter(client *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "srv:gateway:ratelimit:" + rateLimitIdentity(r)

			count, err := rateLimitScript.Run(
				r.Context(), client,
				[]string{key},
				int(window.Seconds()),
			).Int()
			if err != nil {
				utils.Logger.Error("rate limiter redis error, allowing request", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			if count > limit {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func rateLimitIdentity(r *http.Request) string {
	if cookie, err := r.Cookie("access_token"); err == nil {
		if claims, err := utils.VerifyAccessToken(cookie.Value); err == nil {
			if idFloat, ok := claims["id"].(float64); ok {
				return fmt.Sprintf("user:%d", int64(idFloat))
			}
		}
	}
	return "ip:" + clientIP(r)
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}