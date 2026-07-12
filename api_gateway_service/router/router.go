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

func SetUpRouter(
	UserRouter Router,
	RoleRouter Router,
	PermissionRouter Router,
	RolePermissionRouter Router,
	UserRoleRouter Router,
) *chi.Mux {

	router := chi.NewRouter()

	// --- Global middleware (applies to every request) ---
	allowedOrigins := config.GetEnvStringList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"})
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-Correlation-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)

	limiter := rate.NewLimiter(rate.Every(1*time.Minute), 20)
	router.Use(middleware.RateLimiter(limiter))

	// --- Unversioned ---
	router.Get("/health", handler.HealthHandler)

	// --- Versioned API ---
	router.Route("/api/v1", func(r chi.Router) {

		fakeAPIProxy := utils.ProxyToService("https://fakeapi.net", "/api/v1/fakeapi")
		r.Get("/fakeapi", fakeAPIProxy)
		r.Get("/fakeapi/*", fakeAPIProxy)

		// Auth + RBAC management routes (signup, login, roles, permissions...)
		UserRouter.Register(r)
		RoleRouter.Register(r)
		PermissionRouter.Register(r)
		RolePermissionRouter.Register(r)
		UserRoleRouter.Register(r)

		// --- Hotel Service proxy ---
		hotelProxy := utils.ProxyToService(config.ServicesConfig.HOTEL_SERVICE_URL, "/api/v1")

		// Public reads — anyone can browse/search, no login needed.
		r.Get("/hotels", hotelProxy)
		r.Get("/hotel/{id}", hotelProxy)

		// Writes — require login AND the "listing" permission that
		// actually matches what's seeded (host/admin have it, plain
		// "user" role only has listing:read, not write/delete).
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthCookie)
			r.Use(middleware.RequirePermission("listing:write"))
			r.Post("/hotel", hotelProxy)
			r.Patch("/hotel/{id}", hotelProxy)
			r.Patch("/hotel/{id}/restore", hotelProxy)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthCookie)
			r.Use(middleware.RequirePermission("listing:delete"))
			r.Delete("/hotel/{id}", hotelProxy)
		})

		// --- Booking Service proxy ---
		bookingProxy := utils.ProxyToService(config.ServicesConfig.BOOKING_SERVICE_URL, "/api/v1")

		// All booking actions require login.
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthCookie)
			r.Use(middleware.RequirePermission("booking:write"))
			r.Post("/booking/create", bookingProxy)
			r.Post("/booking/confirm/{key}", bookingProxy)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthCookie)
			r.Use(middleware.RequirePermission("booking:delete"))
			r.Post("/booking/cancel/{id}", bookingProxy)
		})

		// --- Review Service proxy ---
		// Not wired yet — Review Service is empty.
	})

	return router
}
