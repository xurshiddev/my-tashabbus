package logger

import (
	"log/slog"
	"os"
)

func New(botEnv string) *slog.Logger {
	level := slog.LevelInfo
	if botEnv == "development" {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
