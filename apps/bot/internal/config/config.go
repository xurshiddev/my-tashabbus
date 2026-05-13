package config

import (
	"os"
	"strings"
)

type Config struct {
	BotEnv     string
	BotToken   string
	APIBaseURL string
	MiniAppURL string
}

func Load() Config {
	return Config{
		BotEnv:     env("BOT_ENV", "development"),
		BotToken:   env("BOT_TOKEN", ""),
		APIBaseURL: env("API_BASE_URL", "http://localhost:8080"),
		MiniAppURL: env("MINI_APP_URL", ""),
	}
}

func (c Config) Validate() error {
	if c.BotToken == "" {
		return ErrMissingBotToken
	}
	return nil
}

var ErrMissingBotToken = configError("BOT_TOKEN is required")

type configError string

func (e configError) Error() string {
	return string(e)
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
