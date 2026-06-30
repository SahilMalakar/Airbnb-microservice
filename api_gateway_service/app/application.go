package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/router"
)

type Config struct {
	Address string
}

type Application struct {
	Config Config
}

// NewConfig is the single source of truth for default values.
func NewConfig() *Config {
	port := config.GetEnvString("PORT", "8080")

	return &Config{
		Address: ":" + port,
	}
}

// NewApplication assumes cfg was built via NewConfig and is valid.
// A nil cfg is a programmer error, not a runtime condition to silently fix.
func NewApplication(cfg *Config) *Application {
	if cfg == nil {
		panic("app: NewApplication called with nil config")
	}
	return &Application{
		Config: *cfg,
	}
}

func (a *Application) RunServer() error {
	chi := router.SetUpRouter()

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
