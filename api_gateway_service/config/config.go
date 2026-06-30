package config

import (
	"log"
	"sync"

	"github.com/joho/godotenv"
	"os"
)

var loadOnce sync.Once

func loadEnv() {
	loadOnce.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("no .env file found, relying on real environment variables")
		}
	})
}

func GetEnvString(key string, fallBack string) string {
	loadEnv()

	val, exists := os.LookupEnv(key)
	if !exists {
		return fallBack
	}

	return val
}