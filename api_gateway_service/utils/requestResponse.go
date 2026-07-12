package utils

import (
	"encoding/json"
	"net/http"
)

// Response is the standard envelope every handler returns, mirroring
// the Node services' sendSuccess/sendError/sendPaginated shape so
type Response struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       any    `json:"data,omitempty"`
	ErrorCode  string `json:"errorCode,omitempty"`
	Meta       *Meta  `json:"meta,omitempty"`
}

type Meta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
}

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


// SendSuccess writes a successful response with the standard envelope.
func SendSuccess(w http.ResponseWriter, statusCode int, message string, data any) error {
	return WriteJSONResponse(w, statusCode, Response{
		Success:    true,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	})
}

// SendError writes an error response with the standard envelope.
// errorCode is a short machine-readable string (e.g. "INVALID_USER_ID"),
func SendError(w http.ResponseWriter, statusCode int, message string, errorCode string) error {
	return WriteJSONResponse(w, statusCode, Response{
		Success:    false,
		StatusCode: statusCode,
		Message:    message,
		ErrorCode:  errorCode,
	})
}

// SendPaginated writes a successful list response with pagination meta.
func SendPaginated(w http.ResponseWriter, statusCode int, message string, data any, total, page, limit int) error {
	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}
	return WriteJSONResponse(w, statusCode, Response{
		Success:    true,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
		Meta: &Meta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}