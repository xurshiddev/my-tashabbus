package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	localmiddleware "github.com/my-tashabbus/api/internal/http/middleware"
)

func TestCORSPreflightAllowsTelegramInitDataHeader(t *testing.T) {
	handler := New(Config{
		ServiceName:        "test",
		CORSAllowedOrigins: []string{"https://miniapp.example"},
	})
	request := httptest.NewRequest(http.MethodOptions, "/miniapp/me", nil)
	request.Header.Set("Origin", "https://miniapp.example")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", localmiddleware.TelegramInitDataHeader+", Content-Type, ngrok-skip-browser-warning")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected preflight status 200, got %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://miniapp.example" {
		t.Fatalf("expected allow origin header, got %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, localmiddleware.TelegramInitDataHeader) {
		t.Fatalf("expected %s in allow headers, got %q", localmiddleware.TelegramInitDataHeader, got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(strings.ToLower(got), "ngrok-skip-browser-warning") {
		t.Fatalf("expected ngrok-skip-browser-warning in allow headers, got %q", got)
	}
}
