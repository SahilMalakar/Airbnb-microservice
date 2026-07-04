package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

// RateLimiter limits each request to a single global rate.
// Use RateLimiterPerIP below if you want per-client limiting instead.
func RateLimiter(limiter *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}