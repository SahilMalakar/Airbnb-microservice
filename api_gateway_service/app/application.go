package app

import (
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	Address string
}

type Application struct {
	Config Config
}

func (a *Application) Run() error {

	mux := http.NewServeMux() // TODO: swap for chi router

	server := &http.Server{
		Addr:         a.Config.Address,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Println("server started listening on address", a.Config.Address)

	return server.ListenAndServe()
}