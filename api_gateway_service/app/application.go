package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/router"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/service"
)

// Config holds runtime configuration for the application.
type Config struct {
	Address string
}

// Application is the top-level struct wiring together config and storage
// dependencies needed to run the server.
type Application struct {
	Config Config
	Store  db.Storage
}

// NewConfig builds a Config using environment variables, falling back to
// somee defaults when they aren't set. This is the single source of truth
// for default values.
func NewConfig() *Config {
	port := config.GetEnvString("PORT", "8080")

	return &Config{
		Address: ":" + port,
	}
}

// NewApplication constructs an Application from a valid Config.
// NewApplication assumes cfg was built via NewConfig and is valid.
// A nil cfg is a programmer error, not a runtime condition to silently fix.
func NewApplication(cfg *Config) *Application {
	if cfg == nil {
		panic("app: NewApplication called with nil config")
	}
	return &Application{
		Config: *cfg,
		Store:  *db.NewStorage(),
	}
}

// RunServer starts the HTTP server using the configured router and address,
// and blocks until the server stops or errors out.
func (a *Application) RunServer() error {
	// Wire dependencies: Storage → Service → Controller → Router
	userService := service.NewUserService(a.Store.UserRepository)
	userController := handler.NewUserController(userService)
	userRouter := router.NewUserRouter(userController)

	chi := router.SetUpRouter(userRouter)

	server := &http.Server{
		Addr:         a.Config.Address,
		Handler:      chi,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Println("server started listening on address", a.Config.Address)

	return server.ListenAndServe()
}
