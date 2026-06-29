package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/app"
)

func main() {
	cfg := app.NewConfig(":8080")

	application := app.NewApplication(cfg)

	err := application.Run()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}