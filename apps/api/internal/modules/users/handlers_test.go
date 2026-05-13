package users

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateHandlerRejectsInvalidRole(t *testing.T) {
	handler := NewHandler(NewService(NewMemoryStore()))
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"full_name":"Ali","role":"ADMIN"}`))
	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}
