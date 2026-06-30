package config

import (
	"log"
	"strconv"
	"sync"

	"os"

	"github.com/joho/godotenv"
)

var loadOnce sync.Once

func LoadEnv() {
	loadOnce.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("no .env file found, relying on real environment variables")
		}
	})
}

func GetEnvString(key string, fallBack string) string {

	val, exists := os.LookupEnv(key)
	if !exists {
		return fallBack
	}

	return val
}

func GetEnvInt(key string, fallBack int) int {

	val, exists := os.LookupEnv(key)
	if !exists {
		return fallBack
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return fallBack
	}

	return valInt
}

func GetEnvBool(key string, fallBack bool) bool {

	val, exists := os.LookupEnv(key)
	if !exists {
		return fallBack
	}

	valBool, err := strconv.ParseBool(val)
	if err != nil {
		return fallBack
	}

	return valBool
}