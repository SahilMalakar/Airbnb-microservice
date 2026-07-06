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

func AuthCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			utils.WriteJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "no token found"})
			return
		}

		claims, err := utils.VerifyAccessToken(cookie.Value)
		if err != nil {
			utils.WriteJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized access"})
			return
		}

		idFloat, ok := claims["id"].(float64)
		if !ok {
			utils.WriteJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
			return
		}
		email, ok := claims["email"].(string)
		if !ok {
			utils.WriteJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
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