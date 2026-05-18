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
	"time"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/config"
	"github.com/my-tashabbus/api/internal/http/requestcontext"
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

func TestTelegramLoginReturnsSpecificInvalidInitDataCode(t *testing.T) {
	handler := newTelegramTestHandler(t, "token", time.Unix(1700000000, 0))
	request := httptest.NewRequest(http.MethodPost, "/auth/telegram", strings.NewReader(`{"init_data":"auth_date=1700000000&hash=bad&user=%7B%22id%22%3A123%7D"}`))
	response := httptest.NewRecorder()

	handler.TelegramLogin(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "TELEGRAM_INIT_DATA_INVALID") {
		t.Fatalf("expected invalid init data code, got %s", response.Body.String())
	}
}

func TestTelegramLoginReturnsSpecificExpiredInitDataCode(t *testing.T) {
	now := time.Unix(1700000000, 0)
	handler := newTelegramTestHandler(t, "token", now)
	body := `{"init_data":"` + signedInitData(t, "token", now.Add(-48*time.Hour), `{"id":123}`) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/auth/telegram", strings.NewReader(body))
	response := httptest.NewRecorder()

	handler.TelegramLogin(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "TELEGRAM_INIT_DATA_EXPIRED") {
		t.Fatalf("expected expired init data code, got %s", response.Body.String())
	}
}

func TestTelegramLoginReturnsUserNotBoundCode(t *testing.T) {
	now := time.Unix(1700000000, 0)
	handler := newTelegramTestHandler(t, "token", now)
	body := `{"init_data":"` + signedInitData(t, "token", now, `{"id":123}`) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/auth/telegram", strings.NewReader(body))
	response := httptest.NewRecorder()

	handler.TelegramLogin(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "TELEGRAM_USER_NOT_BOUND") {
		t.Fatalf("expected user not bound code, got %s", response.Body.String())
	}
}

func TestTelegramLoginSuccessResponseContainsAccessToken(t *testing.T) {
	now := time.Unix(1700000000, 0)
	telegramID := int64(123)
	username := "devuser"
	userService := users.NewService(users.NewMemoryStore())
	if _, err := userService.Create(context.Background(), users.CreateUserInput{
		FullName:         "Telegram Admin",
		TelegramID:       &telegramID,
		TelegramUsername: &username,
		Role:             users.RoleSuperAdmin,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	service, err := NewService(config.Config{
		AppEnv:                   "development",
		JWTSecret:                "dev-secret-change-me",
		JWTAccessTokenTTLMinutes: 60,
		BotToken:                 "token",
	}, userService)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	service.telegramValidator.now = func() time.Time { return now }
	handler := NewHandler(service)
	body := `{"init_data":"` + signedInitData(t, "token", now, `{"id":123,"username":"devuser"}`) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/auth/telegram", strings.NewReader(body))
	response := httptest.NewRecorder()

	handler.TelegramLogin(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	responseBody := response.Body.String()
	if !strings.Contains(responseBody, `"access_token":"`) {
		t.Fatalf("expected access_token in response, got %s", responseBody)
	}
	if !strings.Contains(responseBody, `"token_type":"Bearer"`) {
		t.Fatalf("expected bearer token type in response, got %s", responseBody)
	}
	if !strings.Contains(responseBody, `"role":"SUPER_ADMIN"`) {
		t.Fatalf("expected user in response, got %s", responseBody)
	}
}

func TestMiniAppMeReturnsDirectUserAndMFYPayload(t *testing.T) {
	service, err := NewService(config.Config{
		AppEnv:                   "development",
		JWTSecret:                "dev-secret-change-me",
		JWTAccessTokenTTLMinutes: 60,
	}, users.NewService(users.NewMemoryStore()))
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	service.SetDeploymentMFY(DeploymentMFY{
		ID:   uuid.New(),
		Name: "Test MFY",
		Slug: "test-mfy",
	})
	handler := NewHandler(service)
	user := users.User{
		ID:       uuid.New(),
		FullName: "Mini App User",
		Role:     users.RoleResponsiblePerson,
		IsActive: true,
	}
	request := httptest.NewRequest(http.MethodGet, "/miniapp/me", nil)
	request = request.WithContext(requestcontext.WithCurrentUser(request.Context(), user))
	response := httptest.NewRecorder()

	handler.MiniAppMe(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	body := response.Body.String()
	if strings.Contains(body, `"data":`) {
		t.Fatalf("expected direct mini app payload without data envelope, got %s", body)
	}
	if !strings.Contains(body, `"user":`) {
		t.Fatalf("expected user payload, got %s", body)
	}
	if !strings.Contains(body, `"mfy":`) {
		t.Fatalf("expected mfy payload, got %s", body)
	}
	if !strings.Contains(body, `"role":"RESPONSIBLE_PERSON"`) {
		t.Fatalf("expected user role, got %s", body)
	}
}

func newTelegramTestHandler(t *testing.T, botToken string, now time.Time) Handler {
	t.Helper()
	userService := users.NewService(users.NewMemoryStore())
	service, err := NewService(config.Config{
		AppEnv:                   "development",
		JWTSecret:                "dev-secret-change-me",
		JWTAccessTokenTTLMinutes: 60,
		BotToken:                 botToken,
	}, userService)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	service.telegramValidator.now = func() time.Time { return now }
	return NewHandler(service)
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
