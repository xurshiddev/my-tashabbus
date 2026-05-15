package auth

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/config"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func TestDevLoginLogsInternalErrors(t *testing.T) {
	userService := users.NewService(failingUserStore{err: errors.New("database insert failed")})
	service, err := NewService(config.Config{
		AppEnv:                   "development",
		JWTSecret:                "dev-secret-change-me",
		JWTAccessTokenTTLMinutes: 60,
	}, userService)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}

	var logs bytes.Buffer
	handler := NewHandler(service, slog.New(slog.NewJSONHandler(&logs, nil)))
	request := httptest.NewRequest(http.MethodPost, "/auth/dev-login", strings.NewReader(`{"full_name":"Dev Super Admin","role":"SUPER_ADMIN"}`))
	response := httptest.NewRecorder()

	handler.DevLogin(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "Internal server error") {
		t.Fatalf("expected safe generic client error, got %s", response.Body.String())
	}
	if !strings.Contains(logs.String(), "database insert failed") {
		t.Fatalf("expected internal error in logs, got %s", logs.String())
	}
	if !strings.Contains(logs.String(), "dev_login") {
		t.Fatalf("expected operation in logs, got %s", logs.String())
	}
}

type failingUserStore struct {
	err error
}

func (s failingUserStore) Create(context.Context, users.CreateUserInput) (users.User, error) {
	return users.User{}, s.err
}

func (s failingUserStore) GetByID(context.Context, uuid.UUID) (users.User, error) {
	return users.User{}, s.err
}

func (s failingUserStore) GetByTelegramID(context.Context, int64) (users.User, error) {
	return users.User{}, s.err
}

func (s failingUserStore) List(context.Context, int, int) ([]users.User, error) {
	return nil, s.err
}

func (s failingUserStore) Update(context.Context, uuid.UUID, users.UpdateUserInput) (users.User, error) {
	return users.User{}, s.err
}

func (s failingUserStore) SetTelegramIdentity(context.Context, uuid.UUID, users.SetTelegramIdentityInput) (users.User, error) {
	return users.User{}, s.err
}

func (s failingUserStore) AssignToMFY(context.Context, uuid.UUID, uuid.UUID) (users.User, error) {
	return users.User{}, s.err
}

func (s failingUserStore) Deactivate(context.Context, uuid.UUID) (users.User, error) {
	return users.User{}, s.err
}

func (s failingUserStore) Activate(context.Context, uuid.UUID) (users.User, error) {
	return users.User{}, s.err
}
