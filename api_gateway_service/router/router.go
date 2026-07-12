package router

import (
	"time"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
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

	// CORS: explicit allow-list, read from env so it can differ per
	// environment (local dev vs deployed) without a code change.
	allowedOrigins := config.GetEnvStringList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"})

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-Correlation-Id"},
		AllowCredentials: true, // required since auth relies on HttpOnly cookies
		MaxAge:           300,
	}))

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

	// hotel Service proxy
	hotelProxy := utils.ProxyToService(config.ServicesConfig.HOTEL_SERVICE_URL, "/api/v1")
	router.Route("/api/v1", func(r chi.Router) {
		// public reads -- no token check
		r.Get("/hotels", hotelProxy)
		r.Get("/hotels/{id}", hotelProxy)

		// private writes -- need auth check
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthCookie)
			r.Post("/hotels", hotelProxy)
			r.Patch("/hotels/{id}", hotelProxy)
			r.Patch("/hotels/{id}/restore", hotelProxy)
			r.Delete("/hotels/{id}", hotelProxy)
		})
	})

	// booking service proxy
	bookingProxy := utils.ProxyToService(config.ServicesConfig.BOOKING_SERVICE_URL, "api/v1")
	router.Route("/api/v1", func(r chi.Router) {

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthCookie)
			r.Post("/create", bookingProxy)
			r.Post("/confirm/{key}", bookingProxy)
			r.Post("/cancel/{id}", bookingProxy)
		})
	})

	// review service proxy

	return router
}


