package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/config"
	"github.com/my-tashabbus/api/internal/http/requestcontext"
	"github.com/my-tashabbus/api/internal/modules/auth"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func TestRequireTelegramInitDataMissingHeader(t *testing.T) {
	handler := newTelegramMiddlewareHandler(t, users.NewService(users.NewMemoryStore()), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/miniapp/me", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "TELEGRAM_INIT_DATA_MISSING") {
		t.Fatalf("expected missing init data code, got %s", response.Body.String())
	}
}

func TestRequireTelegramInitDataInvalidInitData(t *testing.T) {
	handler := newTelegramMiddlewareHandler(t, users.NewService(users.NewMemoryStore()), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/miniapp/me", nil)
	request.Header.Set(TelegramInitDataHeader, "auth_date="+strconv.FormatInt(time.Now().Unix(), 10)+"&hash=bad&user=%7B%22id%22%3A123%7D")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "TELEGRAM_INIT_DATA_INVALID") {
		t.Fatalf("expected invalid init data code, got %s", response.Body.String())
	}
}

func TestRequireTelegramInitDataExpiredInitData(t *testing.T) {
	handler := newTelegramMiddlewareHandler(t, users.NewService(users.NewMemoryStore()), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/miniapp/me", nil)
	request.Header.Set(TelegramInitDataHeader, signedMiniAppInitData(t, "token", time.Now().Add(-48*time.Hour), `{"id":123}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "TELEGRAM_INIT_DATA_EXPIRED") {
		t.Fatalf("expected expired init data code, got %s", response.Body.String())
	}
}

func TestRequireTelegramInitDataUnassignedUser(t *testing.T) {
	handler := newTelegramMiddlewareHandler(t, users.NewService(users.NewMemoryStore()), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/miniapp/me", nil)
	request.Header.Set(TelegramInitDataHeader, signedMiniAppInitData(t, "token", time.Now(), `{"id":123}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "USER_NOT_ASSIGNED") {
		t.Fatalf("expected user not assigned code, got %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "Telegram ID: 123") {
		t.Fatalf("expected telegram id in response, got %s", response.Body.String())
	}
}

func TestRequireTelegramInitDataOwnerAutoCreatesChairman(t *testing.T) {
	userService := users.NewService(users.NewMemoryStore())
	authService, err := auth.NewService(config.Config{
		AppEnv:                   "development",
		JWTSecret:                "dev-secret-change-me",
		JWTAccessTokenTTLMinutes: 60,
		BotToken:                 "token",
		MFYOwnerTelegramID:       123,
	}, userService)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	authService.SetDeploymentMFY(auth.DeploymentMFY{
		ID:   mustUUID(t, "00000000-0000-0000-0000-000000000123"),
		Name: "Test MFY",
		Slug: "test-mfy",
	})
	handler := RequireTelegramInitData(authService, nil, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current, err := requestcontext.CurrentUser(r.Context())
		if err != nil {
			t.Fatalf("current user: %v", err)
		}
		if current.Role != users.RoleMFYChairman {
			t.Fatalf("expected role MFY_CHAIRMAN, got %s", current.Role)
		}
		if current.MFYID == nil || current.MFYID.String() != "00000000-0000-0000-0000-000000000123" {
			t.Fatalf("expected owner assigned to deployment mfy, got %v", current.MFYID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/miniapp/me", nil)
	request.Header.Set(TelegramInitDataHeader, signedMiniAppInitData(t, "token", time.Now(), `{"id":123,"username":"owner","first_name":"Owner"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d: %s", response.Code, response.Body.String())
	}
}

func TestRequireTelegramInitDataBoundUserCanAccessProtectedEndpoint(t *testing.T) {
	telegramID := int64(123)
	userService := users.NewService(users.NewMemoryStore())
	if _, err := userService.Create(context.Background(), users.CreateUserInput{
		FullName:   "Bound User",
		TelegramID: &telegramID,
		Role:       users.RoleStreetLeader,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	handler := newTelegramMiddlewareHandler(t, userService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current, err := requestcontext.CurrentUser(r.Context())
		if err != nil {
			t.Fatalf("current user: %v", err)
		}
		if current.Role != users.RoleStreetLeader {
			t.Fatalf("expected role STREET_LEADER, got %s", current.Role)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/miniapp/protected", nil)
	request.Header.Set(TelegramInitDataHeader, signedMiniAppInitData(t, "token", time.Now(), `{"id":123}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d: %s", response.Code, response.Body.String())
	}
}

func newTelegramMiddlewareHandler(t *testing.T, userService *users.Service, next http.Handler) http.Handler {
	t.Helper()
	authService, err := auth.NewService(config.Config{
		AppEnv:                   "development",
		JWTSecret:                "dev-secret-change-me",
		JWTAccessTokenTTLMinutes: 60,
		BotToken:                 "token",
	}, userService)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	return RequireTelegramInitData(authService, nil, false)(next)
}

func signedMiniAppInitData(t *testing.T, botToken string, authDate time.Time, user string) string {
	t.Helper()
	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(authDate.Unix(), 10))
	values.Set("user", user)
	pairs := []string{"auth_date=" + values.Get("auth_date"), "user=" + values.Get("user")}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretHMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretHMAC.Write([]byte(botToken))
	secret := secretHMAC.Sum(nil)

	dataHMAC := hmac.New(sha256.New, secret)
	dataHMAC.Write([]byte(dataCheckString))
	values.Set("hash", hex.EncodeToString(dataHMAC.Sum(nil)))
	return values.Encode()
}

func mustUUID(t *testing.T, raw string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return id
}
