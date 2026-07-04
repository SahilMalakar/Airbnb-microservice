package utils

import (
	"encoding/json"
	"net/http"
)

// convert struct response to json response ,
// use it to send response to client with correct status code and json data
func WriteJSONResponse(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// convert json request to struct ,
// use it to read request from client
func ReadJSONRequest(w http.ResponseWriter, r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}