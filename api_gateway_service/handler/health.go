package handler

import (
	"net/http"
)

// health check handler
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All Good 🔥"))
}