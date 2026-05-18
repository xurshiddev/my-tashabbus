package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/config"
	"github.com/my-tashabbus/api/internal/http/requestcontext"
	"github.com/my-tashabbus/api/internal/modules/auth"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func TestRequireAuthMissingToken(t *testing.T) {
	middleware := RequireAuth(auth.NewTokenManager("secret", time.Hour), users.NewService(users.NewMemoryStore()))
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/private", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
}

func TestRequireAuthInvalidToken(t *testing.T) {
	middleware := RequireAuth(auth.NewTokenManager("secret", time.Hour), users.NewService(users.NewMemoryStore()))
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer invalid-token")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
}

func TestTelegramLoginAccessTokenCanCallMe(t *testing.T) {
	telegramID := int64(123)
	username := "devuser"
	userService := users.NewService(users.NewMemoryStore())
	createdUser, err := userService.Create(context.Background(), users.CreateUserInput{
		FullName:         "Telegram Admin",
		TelegramID:       &telegramID,
		TelegramUsername: &username,
		Role:             users.RoleSuperAdmin,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	authService, err := auth.NewService(config.Config{
		AppEnv:                   "development",
		JWTSecret:                "dev-secret-change-me",
		JWTAccessTokenTTLMinutes: 60,
		TelegramAuthDevMode:      true,
	}, userService)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	loginResponse, err := authService.TelegramLogin(context.Background(), auth.TelegramAuthRequest{
		DevTelegramID: &telegramID,
		DevUsername:   username,
	})
	if err != nil {
		t.Fatalf("telegram login: %v", err)
	}
	if loginResponse.AccessToken == "" {
		t.Fatal("expected access token")
	}

	authHandler := auth.NewHandler(authService)
	handler := RequireAuth(authService.TokenManager(), userService)(http.HandlerFunc(authHandler.Me))
	request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	request.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected /auth/me status 200, got %d: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, `"id":"`+createdUser.ID.String()+`"`) {
		t.Fatalf("expected current user id in response, got %s", body)
	}
	if !strings.Contains(body, `"role":"SUPER_ADMIN"`) {
		t.Fatalf("expected current user role in response, got %s", body)
	}
}

func TestRequireAnyRoleForbidden(t *testing.T) {
	user := users.User{ID: uuid.New(), Role: users.RoleStreetLeader, IsActive: true}
	handler := RequireAnyRole(users.RoleSuperAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request = request.WithContext(requestcontext.WithCurrentUser(request.Context(), user))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", response.Code)
	}
}
