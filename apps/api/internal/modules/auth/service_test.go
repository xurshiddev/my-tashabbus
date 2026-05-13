package auth

import (
	"context"
	"testing"

	"github.com/my-tashabbus/api/internal/config"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func TestDevLoginDisabledInProduction(t *testing.T) {
	userService := users.NewService(users.NewMemoryStore())
	service, err := NewService(config.Config{
		AppEnv:                   "production",
		JWTSecret:                "safe-production-secret",
		JWTAccessTokenTTLMinutes: 60,
	}, userService)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	_, err = service.DevLogin(context.Background(), DevLoginRequest{
		FullName: "Dev",
		Role:     users.RoleSuperAdmin,
	})
	if err != ErrDevLoginDisabled {
		t.Fatalf("expected ErrDevLoginDisabled, got %v", err)
	}
}

func TestDevLoginWorksInDevelopment(t *testing.T) {
	userService := users.NewService(users.NewMemoryStore())
	service, err := NewService(config.Config{
		AppEnv:                   "development",
		JWTSecret:                "dev-secret-change-me",
		JWTAccessTokenTTLMinutes: 60,
	}, userService)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	result, err := service.DevLogin(context.Background(), DevLoginRequest{
		FullName: "Dev Super Admin",
		Role:     users.RoleSuperAdmin,
	})
	if err != nil {
		t.Fatalf("dev login: %v", err)
	}
	if result.AccessToken == "" {
		t.Fatalf("expected access token")
	}
	if result.User.Role != users.RoleSuperAdmin {
		t.Fatalf("expected role SUPER_ADMIN, got %s", result.User.Role)
	}
}
