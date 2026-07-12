package config

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"os"

	"github.com/joho/godotenv"
)

var loadOnce sync.Once

// LoadEnv loads variables from a .env file into the process environment.
// It only runs once per process, even if called multiple times, since
// loadOnce guards it.
func LoadEnv() {
	loadOnce.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("no .env file found, relying on real environment variables")
		}
	})
}

// getEnv is a generic helper that looks up an environment variable and
// parses it into type T using the supplied parse function. If the variable
// isn't set, or parsing fails, fallback is returned instead. This centralizes
// the "lookup -> parse -> fallback on error" pattern shared by GetEnvString,
// GetEnvInt, and GetEnvBool below.
func getEnv[T any](key string, fallback T, parse func(string) (T, error)) T {
	val, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	parsed, err := parse(val)
	if err != nil {
		return fallback
	}

	return parsed
}

// GetEnvString returns the string value of key, or fallBack if unset.
func GetEnvString(key string, fallBack string) string {
	return getEnv(key, fallBack, func(s string) (string, error) {
		return s, nil
	})
}

// GetEnvInt returns the integer value of key, or fallBack if unset or
// unparsable as an int.
func GetEnvInt(key string, fallBack int) int {
	return getEnv(key, fallBack, strconv.Atoi)
}

// GetEnvBool returns the boolean value of key, or fallBack if unset or
// unparsable as a bool.
func GetEnvBool(key string, fallBack bool) bool {
	return getEnv(key, fallBack, strconv.ParseBool)
}

func RequireEnvString(key string) string {
	val, exists := os.LookupEnv(key)
	if !exists || val == "" {
		log.Fatalf("missing required env variable: %s", key)
	}
	return val
}

// GetEnvStringList returns a comma-separated env var split into a slice,
// or fallback if unset. Whitespace around each entry is trimmed.
func GetEnvStringList(key string, fallback []string) []string {
	val, exists := os.LookupEnv(key)
	if !exists || val == "" {
		return fallback
	}
	parts := strings.Split(val, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}


type ServicesConfigType struct {
	HOTEL_SERVICE_URL   string
	BOOKING_SERVICE_URL string
}

var ServicesConfig = ServicesConfigType{
	HOTEL_SERVICE_URL:   RequireEnvString("HOTEL_SERVICE_URL"),
	BOOKING_SERVICE_URL: RequireEnvString("BOOKING_SERVICE_URL"),
}