package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BotEnv             string
	BotToken           string
	APIBaseURL         string
	MiniAppURL         string
	MFYOwnerTelegramID int64
}

func Load() Config {
	return Config{
		BotEnv:             env("BOT_ENV", "development"),
		BotToken:           envFirst([]string{"TELEGRAM_BOT_TOKEN", "BOT_TOKEN"}, ""),
		APIBaseURL:         env("API_BASE_URL", "http://localhost:8080"),
		MiniAppURL:         env("MINI_APP_URL", ""),
		MFYOwnerTelegramID: envInt64First([]string{"MFY_CHAIRMAN_TELEGRAM_ID", "ADMIN_TELEGRAM_ID", "MFY_OWNER_TELEGRAM_ID"}, 0),
	}
}

func (c Config) Validate() error {
	if c.BotToken == "" {
		return ErrMissingBotToken
	}
	return nil
}

var ErrMissingBotToken = configError("TELEGRAM_BOT_TOKEN is required; BOT_TOKEN is supported only as a legacy fallback")

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

func envFirst(keys []string, fallback string) string {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}
	return fallback
}

func envInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt64First(keys []string, fallback int64) int64 {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			continue
		}
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return parsed
		}
	}
	return fallback
}
