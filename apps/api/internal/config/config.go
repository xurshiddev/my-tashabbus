package config

import (
	"net"
	"os"
	"strings"
)

type Config struct {
	AppEnv             string
	APIHost            string
	APIPort            string
	DatabaseURL        string
	CORSAllowedOrigins []string
}

func Load() Config {
	return Config{
		AppEnv:             env("APP_ENV", "development"),
		APIHost:            env("API_HOST", "0.0.0.0"),
		APIPort:            env("API_PORT", "8080"),
		DatabaseURL:        env("DATABASE_URL", ""),
		CORSAllowedOrigins: splitCSV(env("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174")),
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
