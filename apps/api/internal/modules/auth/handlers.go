package auth

import (
	"encoding/json"
	"errors"
	"net/http"

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

func (h Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := requestcontext.CurrentUser(r.Context())
	if err != nil {
		response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing current user")
		return
	}
	response.Data(w, http.StatusOK, user)
}

func (h Handler) DevLogin(w http.ResponseWriter, r *http.Request) {
	var input DevLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	result, err := h.service.DevLogin(r.Context(), input)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	response.Data(w, http.StatusOK, result)
}

func (h Handler) TelegramLogin(w http.ResponseWriter, r *http.Request) {
	var input TelegramAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	result, err := h.service.TelegramLogin(r.Context(), input)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	response.Data(w, http.StatusOK, result)
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrDevLoginDisabled):
		response.ErrorCode(w, http.StatusForbidden, "FORBIDDEN", "Development login is disabled")
	case errors.Is(err, ErrTelegramTokenNeeded), errors.Is(err, ErrInvalidInitData), errors.Is(err, ErrOldInitData):
		response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid Telegram authentication data")
	case errors.Is(err, ErrUserNotRegistered):
		response.ErrorCode(w, http.StatusForbidden, "USER_NOT_REGISTERED", "Telegram user is not registered")
	case errors.Is(err, users.ErrFullNameRequired), errors.Is(err, users.ErrInvalidRole):
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, users.ErrTelegramIDConflict):
		response.ErrorCode(w, http.StatusConflict, "CONFLICT", "Telegram identity is already assigned")
	default:
		response.ErrorCode(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
