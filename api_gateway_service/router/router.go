package router

import (
	"time"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/middleware"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
	"golang.org/x/time/rate"
)

type Router interface {
	Register(r chi.Router)
}

// SetUpRouter builds and returns the chi router with all application
// routes registered.
func SetUpRouter(
	UserRouter Router,
	RoleRouter Router,
	PermissionRouter Router,
	RolePermissionRouter Router,
	UserRoleRouter Router,
) *chi.Mux {

	router := chi.NewRouter()

	// Middleware: log every request (method, path, status, duration)
	router.Use(chimiddleware.Logger)
	// Middleware: recover from panics and log the stack trace
	router.Use(chimiddleware.Recoverer)
	// Middleware: rate limiting
	limiter := rate.NewLimiter(rate.Every(1*time.Minute), 20)
	router.Use(middleware.RateLimiter(limiter))

	// routes
	router.Get("/health", handler.HealthHandler)

	// proxy the request
	fakeAPIProxy := utils.ProxyToService("https://fakeapi.net", "/fakeapi")
	router.Get("/fakeapi", fakeAPIProxy)
	router.Get("/fakeapi/*", fakeAPIProxy)

	UserRouter.Register(router)
	RoleRouter.Register(router)
	PermissionRouter.Register(router)
	RolePermissionRouter.Register(router)
	UserRoleRouter.Register(router)

	return router
}

// https://fakeapi.net/products
