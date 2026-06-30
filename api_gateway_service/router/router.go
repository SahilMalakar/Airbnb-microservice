package router

import (
	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
)

func SetUpRouter() *chi.Mux {

	router := chi.NewRouter()

	router.Get("/health", handler.HealthHandler)

	return router
}
