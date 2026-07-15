package utils

import (
	"log/slog"
	"os"
)

// Logger is the shared structured logger for the API gateway.
var Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))
