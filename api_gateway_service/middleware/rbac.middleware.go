package middleware

import (
	"net/http"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

// RequirePermission passes only if the user holds this exact single permission.
// Use for a route gated by one specific action, e.g. RequirePermission("booking:write").
func RequirePermission(permissionName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, _ := r.Context().Value(CtxUserPerms).([]string)
			if !contains(permissions, permissionName) {
				utils.SendError(w, http.StatusForbidden, "Error on role permission fetching", "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission passes if the user holds AT LEAST ONE of the listed permissions (OR logic).
// Use when several permission levels can satisfy the same route, e.g. RequireAnyPermission("booking:write", "booking:manage").
func RequireAnyPermission(permissionNames ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, _ := r.Context().Value(CtxUserPerms).([]string)
			for _, name := range permissionNames {
				if !contains(permissions, name) {
					utils.SendError(w, http.StatusForbidden, "Error on role permission fetching", "insufficient permissions")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllPermissions passes only if the user holds EVERY listed permission (AND logic).
// Use for sensitive combined actions needing multiple permissions together, e.g. RequireAllPermissions("user:delete", "user:manage")
func RequireAllPermissions(permissionNames ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, _ := r.Context().Value(CtxUserPerms).([]string)
			for _, name := range permissionNames {
				if !contains(permissions, name) {
					utils.SendError(w, http.StatusForbidden, "Error on role permission fetching", "insufficient permissions")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole passes only if the user has this exact single role.
// Use for routes with one clear owner role, e.g. RequireRole("admin").
func RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, _ := r.Context().Value(CtxUserRoles).([]string)
			if !contains(roles, roleName) {
				utils.SendError(w, http.StatusForbidden, "Error on role permission fetching", "insufficient role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole passes if the user has AT LEAST ONE of the listed roles (OR logic).
// Use when multiple different roles should all be allowed, e.g. RequireAnyRole("host", "admin").
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
			utils.SendError(w, http.StatusForbidden, "Error on role permission fetching", "insufficient role")
		})
	}
}

// RequireAllRoles passes only if the user has EVERY listed role at once (AND logic).
// Use for rare cases needing multiple roles together, e.g. RequireAllRoles("host", "verified").
func RequireAllRoles(roleNames ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, _ := r.Context().Value(CtxUserRoles).([]string)
			for _, name := range roleNames {
				if !contains(roles, name) {
					utils.SendError(w, http.StatusForbidden, "Error on role permission fetching", "insufficient role")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// contains is a helper function that checks if a slice of strings contains a target string.
func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}