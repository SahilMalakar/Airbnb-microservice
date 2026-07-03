package router

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
)

type Router interface {
	Register(r chi.Router)
}

// SetUpRouter builds and returns the chi router with all application
// routes registered.
func SetUpRouter(UserRouter Router) *chi.Mux {

	router := chi.NewRouter()

	// Middleware: log every request (method, path, status, duration)
	router.Use(middleware.Logger)
	// Middleware: recover from panics and log the stack trace
	router.Use(middleware.Recoverer)

	router.Get("/health", handler.HealthHandler)

	UserRouter.Register(router)

	return router
}
