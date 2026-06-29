package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/app"
)

func main() {
	cfg := app.Config{
		Address: ":8080",
	}

	server := app.Application{
		Config: cfg,
	}

	err := server.Run()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}