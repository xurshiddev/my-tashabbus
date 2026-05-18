package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestTelegramLoginReturnsUserNotRegisteredWhenTelegramUserNotBound(t *testing.T) {
	now := time.Unix(1700000000, 0)
	userService := users.NewService(users.NewMemoryStore())
	service, err := NewService(config.Config{
		AppEnv:                   "production",
		JWTSecret:                "safe-production-secret",
		JWTAccessTokenTTLMinutes: 60,
		BotToken:                 "token",
	}, userService)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	service.telegramValidator.now = func() time.Time { return now.Add(time.Minute) }

	_, err = service.TelegramLogin(context.Background(), TelegramAuthRequest{
		InitData: signedInitData(t, "token", now, `{"id":123,"username":"dev"}`),
	})
	if err != ErrUserNotRegistered {
		t.Fatalf("expected ErrUserNotRegistered, got %v", err)
	}
}

func TestTelegramLoginReturnsTokenWhenTelegramUserBound(t *testing.T) {
	now := time.Unix(1700000000, 0)
	telegramID := int64(123)
	userService := users.NewService(users.NewMemoryStore())
	created, err := userService.Create(context.Background(), users.CreateUserInput{
		FullName:   "Bound Telegram User",
		TelegramID: &telegramID,
		Role:       users.RoleResponsiblePerson,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	service, err := NewService(config.Config{
		AppEnv:                   "production",
		JWTSecret:                "safe-production-secret",
		JWTAccessTokenTTLMinutes: 60,
		BotToken:                 "token",
	}, userService)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	service.telegramValidator.now = func() time.Time { return now.Add(time.Minute) }

	result, err := service.TelegramLogin(context.Background(), TelegramAuthRequest{
		InitData: signedInitData(t, "token", now, `{"id":123,"username":"dev"}`),
	})
	if err != nil {
		t.Fatalf("telegram login: %v", err)
	}
	if result.AccessToken == "" {
		t.Fatalf("expected access token")
	}
	if result.User.ID != created.ID {
		t.Fatalf("expected user %s, got %s", created.ID, result.User.ID)
	}
}

func TestUserFromTelegramInitDataCreatesChairmanForConfiguredID(t *testing.T) {
	now := time.Unix(1700000000, 0)
	userService := users.NewService(users.NewMemoryStore())
	service := newMiniAppAuthService(t, userService, 123, now)

	user, _, err := service.UserFromTelegramInitData(context.Background(), signedInitData(t, "token", now, `{"id":123,"username":"chairman","first_name":"Ali"}`))
	if err != nil {
		t.Fatalf("user from init data: %v", err)
	}
	if user.Role != users.RoleMFYChairman {
		t.Fatalf("expected MFY_CHAIRMAN, got %s", user.Role)
	}
	if user.TelegramID == nil || *user.TelegramID != 123 {
		t.Fatalf("expected telegram id 123, got %v", user.TelegramID)
	}
	if user.MFYID == nil || *user.MFYID != testDeploymentMFYID {
		t.Fatalf("expected deployment mfy assignment, got %v", user.MFYID)
	}
}

func TestUserFromTelegramInitDataEnsuresExistingChairmanRole(t *testing.T) {
	now := time.Unix(1700000000, 0)
	telegramID := int64(123)
	userService := users.NewService(users.NewMemoryStore())
	if _, err := userService.Create(context.Background(), users.CreateUserInput{
		FullName:   "Existing",
		TelegramID: &telegramID,
		Role:       users.RoleResponsiblePerson,
	}); err != nil {
		t.Fatalf("create existing user: %v", err)
	}
	service := newMiniAppAuthService(t, userService, 123, now)

	user, _, err := service.UserFromTelegramInitData(context.Background(), signedInitData(t, "token", now, `{"id":123,"username":"chairman"}`))
	if err != nil {
		t.Fatalf("user from init data: %v", err)
	}
	if user.Role != users.RoleMFYChairman {
		t.Fatalf("expected role ensured to MFY_CHAIRMAN, got %s", user.Role)
	}
	if user.MFYID == nil || *user.MFYID != testDeploymentMFYID {
		t.Fatalf("expected deployment mfy assignment, got %v", user.MFYID)
	}
}

func TestUserFromTelegramInitDataUnknownNonChairmanReturnsNotAssigned(t *testing.T) {
	now := time.Unix(1700000000, 0)
	userService := users.NewService(users.NewMemoryStore())
	service := newMiniAppAuthService(t, userService, 123, now)

	_, _, err := service.UserFromTelegramInitData(context.Background(), signedInitData(t, "token", now, `{"id":456,"username":"unknown"}`))
	if err != ErrUserNotAssigned {
		t.Fatalf("expected ErrUserNotAssigned, got %v", err)
	}
}

func TestUserFromTelegramInitDataReturnsAssignedStreetLeader(t *testing.T) {
	now := time.Unix(1700000000, 0)
	telegramID := int64(456)
	userService := users.NewService(users.NewMemoryStore())
	if _, err := userService.Create(context.Background(), users.CreateUserInput{
		FullName:   "Street Leader",
		TelegramID: &telegramID,
		Role:       users.RoleStreetLeader,
		MFYID:      &testDeploymentMFYID,
	}); err != nil {
		t.Fatalf("create street leader: %v", err)
	}
	service := newMiniAppAuthService(t, userService, 123, now)

	user, _, err := service.UserFromTelegramInitData(context.Background(), signedInitData(t, "token", now, `{"id":456,"username":"leader"}`))
	if err != nil {
		t.Fatalf("user from init data: %v", err)
	}
	if user.Role != users.RoleStreetLeader {
		t.Fatalf("expected STREET_LEADER, got %s", user.Role)
	}
}

func TestUserFromTelegramInitDataReturnsAssignedResponsiblePerson(t *testing.T) {
	now := time.Unix(1700000000, 0)
	telegramID := int64(789)
	userService := users.NewService(users.NewMemoryStore())
	if _, err := userService.Create(context.Background(), users.CreateUserInput{
		FullName:   "Responsible",
		TelegramID: &telegramID,
		Role:       users.RoleResponsiblePerson,
		MFYID:      &testDeploymentMFYID,
	}); err != nil {
		t.Fatalf("create responsible: %v", err)
	}
	service := newMiniAppAuthService(t, userService, 123, now)

	user, _, err := service.UserFromTelegramInitData(context.Background(), signedInitData(t, "token", now, `{"id":789,"username":"responsible"}`))
	if err != nil {
		t.Fatalf("user from init data: %v", err)
	}
	if user.Role != users.RoleResponsiblePerson {
		t.Fatalf("expected RESPONSIBLE_PERSON, got %s", user.Role)
	}
}

var testDeploymentMFYID = uuid.MustParse("00000000-0000-0000-0000-000000000123")

func newMiniAppAuthService(t *testing.T, userService *users.Service, chairmanTelegramID int64, now time.Time) *Service {
	t.Helper()
	service, err := NewService(config.Config{
		AppEnv:                   "development",
		JWTSecret:                "dev-secret-change-me",
		JWTAccessTokenTTLMinutes: 60,
		BotToken:                 "token",
		MFYOwnerTelegramID:       chairmanTelegramID,
	}, userService)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	service.telegramValidator.now = func() time.Time { return now.Add(time.Minute) }
	service.SetDeploymentMFY(DeploymentMFY{
		ID:   testDeploymentMFYID,
		Name: "Test MFY",
		Slug: "test-mfy",
	})
	return service
}
