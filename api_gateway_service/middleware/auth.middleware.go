package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/cache"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type ctxKey string

const (
	CtxUserID       ctxKey = "userID"
	CtxUserEmail    ctxKey = "userEmail"
	CtxUserFamilyID ctxKey = "userFamilyID"
	CtxUserRoles    ctxKey = "userRoles"
	CtxUserPerms    ctxKey = "userPermissions"
)

var (
	refreshStore     cache.RefreshTokenStore
	refreshStoreOnce sync.Once
)

func InitAuthDependencies(store cache.RefreshTokenStore) {
	refreshStoreOnce.Do(func() {
		refreshStore = store
	})
}

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

		familyID, ok := claims["familyId"].(string)
		if !ok {
			utils.SendError(w, http.StatusUnauthorized, "Error on token verification", "invalid token claims")
			return
		}

		if refreshStore != nil {
			denylisted, err := refreshStore.IsFamilyDenylisted(r.Context(), familyID)
			if err != nil {
				utils.Logger.Error("denylist check redis error, allowing request", "error", err, "familyID", familyID)
			} else if denylisted {
				utils.SendError(w, http.StatusUnauthorized, "Error on token verification", "session revoked")
				return
			}
		}

		roles := toStringSlice(claims["roles"])
		permissions := toStringSlice(claims["permissions"])
		userID := int64(idFloat)
		name, _ := claims["name"].(string)

		// still set headers for userID/email/name — useful if you proxy to
		// downstream services that read headers, not Go context
		r.Header.Set("X-User-ID", strconv.FormatInt(userID, 10))
		r.Header.Set("X-User-Email", email)
		r.Header.Set("X-User-Name", name)

		ctx := context.WithValue(r.Context(), CtxUserID, userID)
		ctx = context.WithValue(ctx, CtxUserEmail, email)
		ctx = context.WithValue(ctx, CtxUserFamilyID, familyID)
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
