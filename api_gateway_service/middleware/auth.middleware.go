package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

func AuthCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("[DEBUG] AuthCookie: Request received")

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

		// we need to send the role also


		userID := int64(idFloat)

		r.Header.Set("X-User-ID", strconv.FormatInt(userID, 10))
		r.Header.Set("X-User-Email", email)

		next.ServeHTTP(w, r)
	})
}