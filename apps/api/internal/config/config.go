package config

import (
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv                   string
	APIHost                  string
	APIPort                  string
	DatabaseURL              string
	CORSAllowedOrigins       []string
	JWTSecret                string
	JWTAccessTokenTTLMinutes int
	BotToken                 string
	TelegramBotToken         string
	TelegramAuthDevMode      bool
	AppMFYName               string
	AppMFYSlug               string
	MFYOwnerTelegramID       int64
}

func Load() Config {
	return Config{
		AppEnv:                   env("APP_ENV", "development"),
		APIHost:                  env("API_HOST", "0.0.0.0"),
		APIPort:                  env("API_PORT", "8080"),
		DatabaseURL:              env("DATABASE_URL", ""),
		CORSAllowedOrigins:       splitCSV(env("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174")),
		JWTSecret:                env("JWT_SECRET", "dev-secret-change-me"),
		JWTAccessTokenTTLMinutes: envInt("JWT_ACCESS_TOKEN_TTL_MINUTES", 1440),
		BotToken:                 envFirst([]string{"TELEGRAM_BOT_TOKEN", "BOT_TOKEN"}, ""),
		TelegramBotToken:         envFirst([]string{"TELEGRAM_BOT_TOKEN", "BOT_TOKEN"}, ""),
		TelegramAuthDevMode:      envBool("TELEGRAM_AUTH_DEV_MODE", true),
		AppMFYName:               env("APP_MFY_NAME", "My Tashabbus MFY"),
		AppMFYSlug:               env("APP_MFY_SLUG", "my-tashabbus-mfy"),
		MFYOwnerTelegramID:       envInt64First([]string{"MFY_CHAIRMAN_TELEGRAM_ID", "ADMIN_TELEGRAM_ID", "MFY_OWNER_TELEGRAM_ID"}, 0),
	}
}

func (c Config) Address() string {
	return net.JoinHostPort(c.APIHost, c.APIPort)
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

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
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

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "true" || value == "1" || value == "yes"
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
