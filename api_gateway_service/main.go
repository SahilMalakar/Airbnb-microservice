package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/app"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
)

func main() {

	config.LoadEnv()

	db, err := config.LoadDb()
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	cfg := app.NewConfig()

	application := app.NewApplication(cfg, db)

	err = application.RunServer()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
