package streets

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/http/requestcontext"
	"github.com/my-tashabbus/api/internal/modules/mfys"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func TestCreateHandlerValidationError(t *testing.T) {
	userService := users.NewService(users.NewMemoryStore())
	mfyService := mfys.NewService(mfys.NewMemoryStore(), userService)
	handler := NewHandler(NewService(NewMemoryStore(), mfyService, userService))
	request := httptest.NewRequest(http.MethodPost, "/mfys/invalid/streets", strings.NewReader(`{"name":""}`))
	request = request.WithContext(requestcontext.WithCurrentUser(request.Context(), users.User{ID: uuid.New(), Role: users.RoleSuperAdmin}))
	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}
