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
	TelegramAuthDevMode      bool
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
		BotToken:                 env("BOT_TOKEN", ""),
		TelegramAuthDevMode:      envBool("TELEGRAM_AUTH_DEV_MODE", true),
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
