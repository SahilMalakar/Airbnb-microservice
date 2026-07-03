package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/app"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
)

// main is the entry point: it loads environment config, builds the
// application, and starts the server, exiting non-zero on unexpected
// startup errors.
func main() {

	config.LoadEnv()

	cfg := app.NewConfig()

	application := app.NewApplication(cfg)

	err := application.RunServer()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}

}