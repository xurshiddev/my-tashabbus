package auth

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/my-tashabbus/api/internal/http/requestcontext"
	"github.com/my-tashabbus/api/internal/http/response"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Handler struct {
	service *Service
	log     *slog.Logger
}

func NewHandler(service *Service, loggers ...*slog.Logger) Handler {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	if len(loggers) > 0 && loggers[0] != nil {
		log = loggers[0]
	}
	return Handler{service: service, log: log}
}

func (h Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := requestcontext.CurrentUser(r.Context())
	if err != nil {
		response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing current user")
		return
	}
	response.Data(w, http.StatusOK, user)
}

func (h Handler) MiniAppMe(w http.ResponseWriter, r *http.Request) {
	user, err := requestcontext.CurrentUser(r.Context())
	if err != nil {
		response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing current user")
		return
	}
	mfy, err := requestcontext.CurrentMFY(r.Context())
	if err != nil {
		if deploymentMFY, ok := h.service.DeploymentMFY(); ok {
			mfy = requestcontext.MFYContext{
				ID:   deploymentMFY.ID,
				Name: deploymentMFY.Name,
				Slug: deploymentMFY.Slug,
			}
		}
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"user": user,
		"mfy":  mfy,
	})
}

func (h Handler) DevLogin(w http.ResponseWriter, r *http.Request) {
	var input DevLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	result, err := h.service.DevLogin(r.Context(), input)
	if err != nil {
		h.writeAuthError(w, "dev_login", err)
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
	if h.service.AppEnv() != "production" {
		h.logTelegramDiagnostics("telegram_login_request", input.InitData)
	}
	result, err := h.service.TelegramLogin(r.Context(), input)
	if err != nil {
		if h.service.AppEnv() != "production" {
			h.logTelegramDiagnostics("telegram_login_failed", input.InitData)
			if errors.Is(err, ErrUserNotRegistered) {
				h.log.Info("telegram local user lookup failed",
					"operation", "telegram_login",
					"local_user_found", false,
				)
			}
		}
		h.writeAuthError(w, "telegram_login", err)
		return
	}
	if h.service.AppEnv() != "production" {
		h.log.Info("telegram auth succeeded",
			"operation", "telegram_login",
			"local_user_id", result.User.ID,
			"local_user_role", result.User.Role,
		)
	}
	response.Data(w, http.StatusOK, result)
}

func (h Handler) logTelegramDiagnostics(message, initData string) {
	diagnostics := h.service.TelegramDiagnostics(initData)
	attrs := []any{
		"operation", "telegram_login",
		"init_data_present", diagnostics.InitDataPresent,
		"init_data_length", diagnostics.InitDataLength,
		"server_time", diagnostics.ServerTime,
		"max_allowed_age_seconds", diagnostics.MaxAgeSeconds,
		"validation_result", diagnostics.Result,
	}
	if diagnostics.TelegramUserID != nil {
		attrs = append(attrs, "telegram_user_id", *diagnostics.TelegramUserID)
	}
	if diagnostics.AuthDate != nil {
		attrs = append(attrs, "auth_date", *diagnostics.AuthDate)
	}
	if diagnostics.AgeSeconds != nil {
		attrs = append(attrs, "age_seconds", *diagnostics.AgeSeconds)
	}
	h.log.Info(message, attrs...)
}

func (h Handler) writeAuthError(w http.ResponseWriter, operation string, err error) {
	switch {
	case errors.Is(err, ErrDevLoginDisabled):
		response.ErrorCode(w, http.StatusForbidden, "FORBIDDEN", "Development login is disabled")
	case errors.Is(err, ErrOldInitData):
		response.ErrorCode(w, http.StatusUnauthorized, "TELEGRAM_INIT_DATA_EXPIRED", "Telegram init data is expired")
	case errors.Is(err, ErrTelegramTokenNeeded), errors.Is(err, ErrInvalidInitData):
		response.ErrorCode(w, http.StatusUnauthorized, "TELEGRAM_INIT_DATA_INVALID", "Invalid Telegram authentication data")
	case errors.Is(err, ErrUserNotRegistered):
		response.ErrorCode(w, http.StatusForbidden, "TELEGRAM_USER_NOT_BOUND", "Telegram user is not bound to a local user")
	case errors.Is(err, users.ErrFullNameRequired), errors.Is(err, users.ErrInvalidRole):
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, users.ErrTelegramIDConflict):
		response.ErrorCode(w, http.StatusConflict, "CONFLICT", "Telegram identity is already assigned")
	default:
		h.log.Error("auth request failed", "operation", operation, "error", err)
		response.ErrorCode(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
