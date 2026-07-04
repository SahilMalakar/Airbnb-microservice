package middleware

import (
	"net/http"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type ValidatedHandler[T any] func(w http.ResponseWriter, r *http.Request, payload T)

// DecodeAndValidate wraps a ValidatedHandler, decoding + validating the
// request body into T before calling the handler with the typed payload.
func DecodeAndValidate[T any](h ValidatedHandler[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload T

		if err := utils.ReadJSONRequest(w, r, &payload); err != nil {
			utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		if err := utils.ValidateStruct(payload); err != nil {
			utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		h(w, r, payload)
	}
}
