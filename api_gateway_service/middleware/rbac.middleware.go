package middleware

import (
	"net/http"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

func RequirePermission(permissionName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, _ := r.Context().Value(CtxUserPerms).([]string)
			if !contains(permissions, permissionName) {
				utils.WriteJSONResponse(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAnyPermission(permissionNames ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, _ := r.Context().Value(CtxUserPerms).([]string)
			for _, name := range permissionNames {
				if contains(permissions, name) {
					next.ServeHTTP(w, r)
					return
				}
			}
			utils.WriteJSONResponse(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		})
	}
}

func RequireAllPermissions(permissionNames ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, _ := r.Context().Value(CtxUserPerms).([]string)
			for _, name := range permissionNames {
				if !contains(permissions, name) {
					utils.WriteJSONResponse(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, _ := r.Context().Value(CtxUserRoles).([]string)
			if !contains(roles, roleName) {
				utils.WriteJSONResponse(w, http.StatusForbidden, map[string]string{"error": "insufficient role"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAnyRole(roleNames ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, _ := r.Context().Value(CtxUserRoles).([]string)
			for _, name := range roleNames {
				if contains(roles, name) {
					next.ServeHTTP(w, r)
					return
				}
			}
			utils.WriteJSONResponse(w, http.StatusForbidden, map[string]string{"error": "insufficient role"})
		})
	}
}

func RequireAllRoles(roleNames ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, _ := r.Context().Value(CtxUserRoles).([]string)
			for _, name := range roleNames {
				if !contains(roles, name) {
					utils.WriteJSONResponse(w, http.StatusForbidden, map[string]string{"error": "insufficient role"})
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}