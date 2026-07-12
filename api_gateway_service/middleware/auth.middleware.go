package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type ctxKey string

const (
	CtxUserID      ctxKey = "userID"
	CtxUserEmail   ctxKey = "userEmail"
	CtxUserRoles   ctxKey = "userRoles"
	CtxUserPerms   ctxKey = "userPermissions"
)

// AuthCookie verifies the JWT from the access_token cookie and extracts
// identity + authorization data from its claims — no DB call, everything
// needed lives in the signed token (roles/permissions embedded at login).
//
// Sets X-User-ID / X-User-Email as HEADERS: use headers only for values that
// must cross a process boundary — e.g. forwarded onward by ReverseProxy to
// downstream services (Hotel/Booking/Review) that can't read this Go
// process's context.Context.
//
// Sets roles/permissions on request CONTEXT: use context for values only
// needed within this same process/binary — e.g. by RequireRole/RequirePermission
// middleware further down the chain. Context avoids stringly-typed array
// serialization (join/split) and stays properly typed as []string.
func AuthCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			utils.SendError(w, http.StatusUnauthorized, "Error on token verification", "no token found")
			return
		}

		claims, err := utils.VerifyAccessToken(cookie.Value)
		if err != nil {
			utils.SendError(w, http.StatusUnauthorized, "Error on token verification", "unauthorized access")
			return
		}

		idFloat, ok := claims["id"].(float64)
		if !ok {
			utils.SendError(w, http.StatusUnauthorized, "Error on token verification", "invalid token claims")
			return
		}
		email, ok := claims["email"].(string)
		if !ok {
			utils.SendError(w, http.StatusUnauthorized, "Error on token verification", "invalid token claims")
			return
		}

		roles := toStringSlice(claims["roles"])
		permissions := toStringSlice(claims["permissions"])
		userID := int64(idFloat)

		// still set headers for userID/email — useful if you proxy to
		// downstream services that read headers, not Go context
		r.Header.Set("X-User-ID", strconv.FormatInt(userID, 10))
		r.Header.Set("X-User-Email", email)

		ctx := context.WithValue(r.Context(), CtxUserID, userID)
		ctx = context.WithValue(ctx, CtxUserEmail, email)
		ctx = context.WithValue(ctx, CtxUserRoles, roles)
		ctx = context.WithValue(ctx, CtxUserPerms, permissions)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// toStringSlice safely converts a JWT claim (decoded as []interface{}) to []string.
func toStringSlice(v interface{}) []string {
	raw, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}