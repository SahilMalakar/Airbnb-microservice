package handler

import (
	"net/http"
)

// HealthHandler responds to health-check requests with a 200 OK and a
// short confirmation message, used to verify the server is up and reachable.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	// array of byte ---> array of character
	w.Write([]byte("All Good 🔥"))
}