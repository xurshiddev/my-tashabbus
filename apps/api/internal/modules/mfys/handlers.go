package mfys

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/http/requestcontext"
	"github.com/my-tashabbus/api/internal/http/response"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) Handler {
	return Handler{service: service}
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	current, ok := currentUser(w, r)
	if !ok {
		return
	}
	limit := clamp(parseIntDefault(r.URL.Query().Get("limit"), 20), 1, 100)
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
	items, err := h.service.List(r.Context(), current, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Paginated(w, http.StatusOK, items, limit, offset)
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	current, ok := currentUser(w, r)
	if !ok {
		return
	}
	var input CreateMFYInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	mfy, err := h.service.Create(r.Context(), current, input)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Data(w, http.StatusCreated, mfy)
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	current, id, ok := currentUserAndID(w, r, "id")
	if !ok {
		return
	}
	mfy, err := h.service.Get(r.Context(), current, id)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Data(w, http.StatusOK, mfy)
}

func (h Handler) Update(w http.ResponseWriter, r *http.Request) {
	current, id, ok := currentUserAndID(w, r, "id")
	if !ok {
		return
	}
	var input UpdateMFYInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	mfy, err := h.service.Update(r.Context(), current, id, input)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Data(w, http.StatusOK, mfy)
}

func (h Handler) AssignChairman(w http.ResponseWriter, r *http.Request) {
	current, id, ok := currentUserAndID(w, r, "id")
	if !ok {
		return
	}
	var input AssignChairmanInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	user, err := h.service.AssignChairman(r.Context(), current, id, input.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Data(w, http.StatusOK, user)
}

func currentUser(w http.ResponseWriter, r *http.Request) (users.User, bool) {
	user, err := requestcontext.CurrentUser(r.Context())
	if err != nil {
		response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing current user")
		return users.User{}, false
	}
	return user, true
}

func currentUserAndID(w http.ResponseWriter, r *http.Request, param string) (users.User, uuid.UUID, bool) {
	user, ok := currentUser(w, r)
	if !ok {
		return users.User{}, uuid.UUID{}, false
	}
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid id")
		return users.User{}, uuid.UUID{}, false
	}
	return user, id, true
}

func parseIntDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrForbidden):
		response.ErrorCode(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource")
	case errors.Is(err, ErrMFYNotFound), errors.Is(err, users.ErrUserNotFound):
		response.ErrorCode(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, ErrNameRequired), errors.Is(err, ErrTargetVotesNegative), errors.Is(err, ErrInvalidPagination), errors.Is(err, ErrChairmanRoleRequired):
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		response.ErrorCode(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
