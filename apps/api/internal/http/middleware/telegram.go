package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/my-tashabbus/api/internal/http/requestcontext"
	"github.com/my-tashabbus/api/internal/modules/auth"
	"github.com/my-tashabbus/api/internal/modules/users"
)

const TelegramInitDataHeader = "X-Telegram-Init-Data"

func RequireTelegramInitData(authService *auth.Service, log *slog.Logger, enabled bool) func(http.Handler) http.Handler {
	if log == nil {
		log = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			initData := r.Header.Get(TelegramInitDataHeader)
			if initData == "" {
				logTelegramMiniAppDiagnostics(log, enabled, "missing", auth.TelegramValidationDiagnostics{
					InitDataPresent: false,
					InitDataLength:  0,
					Result:          "missing",
				}, false, "")
				writeTelegramMiniAppError(w, http.StatusUnauthorized, "TELEGRAM_INIT_DATA_MISSING", "Telegram initData is required.")
				return
			}

			user, diagnostics, err := authService.UserFromTelegramInitData(r.Context(), initData)
			if err != nil {
				localUserFound := false
				role := ""
				status := http.StatusUnauthorized
				code := "TELEGRAM_INIT_DATA_INVALID"
				message := "Invalid Telegram authentication data"
				switch {
				case errors.Is(err, auth.ErrOldInitData):
					code = "TELEGRAM_INIT_DATA_EXPIRED"
					message = "Telegram initData is expired"
				case errors.Is(err, auth.ErrTelegramTokenNeeded), errors.Is(err, auth.ErrInvalidInitData):
					code = "TELEGRAM_INIT_DATA_INVALID"
					message = "Invalid Telegram authentication data"
				case errors.Is(err, auth.ErrUserNotAssigned), errors.Is(err, auth.ErrUserNotRegistered):
					code = "USER_NOT_ASSIGNED"
					message = "Telegram user is not assigned to this MFY."
					if diagnostics.TelegramUserID != nil {
						message = "Telegram user is not assigned to this MFY. Telegram ID: " + strconv.FormatInt(*diagnostics.TelegramUserID, 10)
					}
				case errors.Is(err, users.ErrUserInactive):
					status = http.StatusForbidden
					code = "FORBIDDEN"
					message = "User is inactive"
				default:
					log.Error("telegram mini app auth failed", "error", err)
					code = "INTERNAL_ERROR"
					message = "Internal server error"
					status = http.StatusInternalServerError
				}
				logTelegramMiniAppDiagnostics(log, enabled, "rejected", diagnostics, localUserFound, role)
				writeTelegramMiniAppError(w, status, code, message)
				return
			}

			logTelegramMiniAppDiagnostics(log, enabled, "accepted", diagnostics, true, string(user.Role))
			ctx := requestcontext.WithCurrentUser(r.Context(), user)
			if mfy, ok := authService.DeploymentMFY(); ok {
				ctx = requestcontext.WithCurrentMFY(ctx, requestcontext.MFYContext{
					ID:   mfy.ID,
					Name: mfy.Name,
					Slug: mfy.Slug,
				})
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeTelegramMiniAppError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}{
		Error:   code,
		Message: message,
	})
}

func logTelegramMiniAppDiagnostics(
	log *slog.Logger,
	enabled bool,
	result string,
	diagnostics auth.TelegramValidationDiagnostics,
	localUserFound bool,
	role string,
) {
	if !enabled {
		return
	}
	attrs := []any{
		"operation", "telegram_miniapp_auth",
		"result", result,
		"init_data_present", diagnostics.InitDataPresent,
		"init_data_length", diagnostics.InitDataLength,
		"validation_result", diagnostics.Result,
		"local_user_found", localUserFound,
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
	if diagnostics.MaxAgeSeconds > 0 {
		attrs = append(attrs, "max_allowed_age_seconds", diagnostics.MaxAgeSeconds)
	}
	if !diagnostics.ServerTime.IsZero() {
		attrs = append(attrs, "server_time", diagnostics.ServerTime)
	}
	if role != "" {
		attrs = append(attrs, "local_user_role", role)
	}
	log.Info("telegram mini app auth diagnostics", attrs...)
}
