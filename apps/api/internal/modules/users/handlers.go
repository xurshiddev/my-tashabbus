package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/http/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) Handler {
	return Handler{service: service}
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	limit := parseIntDefault(r.URL.Query().Get("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
	users, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		writeUserError(w, err)
		return
	}
	response.Data(w, http.StatusOK, users)
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	user, err := h.service.Create(r.Context(), input)
	if err != nil {
		writeUserError(w, err)
		return
	}
	response.Data(w, http.StatusCreated, user)
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeUserError(w, err)
		return
	}
	response.Data(w, http.StatusOK, user)
}

func (h Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var input UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	user, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		writeUserError(w, err)
		return
	}
	response.Data(w, http.StatusOK, user)
}

func (h Handler) SetTelegramIdentity(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var input SetTelegramIdentityInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	user, err := h.service.SetTelegramIdentity(r.Context(), id, input)
	if err != nil {
		writeUserError(w, err)
		return
	}
	response.Data(w, http.StatusOK, user)
}

func (h Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	user, err := h.service.Deactivate(r.Context(), id)
	if err != nil {
		writeUserError(w, err)
		return
	}
	response.Data(w, http.StatusOK, user)
}

func (h Handler) Activate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	user, err := h.service.Activate(r.Context(), id)
	if err != nil {
		writeUserError(w, err)
		return
	}
	response.Data(w, http.StatusOK, user)
}

func parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid user id")
		return uuid.UUID{}, false
	}
	return id, true
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

func writeUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrFullNameRequired), errors.Is(err, ErrInvalidRole), errors.Is(err, ErrTelegramIDRequired), errors.Is(err, ErrInvalidPagination):
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, ErrUserNotFound):
		response.ErrorCode(w, http.StatusNotFound, "NOT_FOUND", "User not found")
	case errors.Is(err, ErrTelegramIDConflict):
		response.ErrorCode(w, http.StatusConflict, "CONFLICT", "Telegram identity is already assigned")
	case errors.Is(err, ErrStoreNotConfigured):
		response.ErrorCode(w, http.StatusInternalServerError, "INTERNAL_ERROR", "User store is not configured")
	default:
		response.ErrorCode(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
