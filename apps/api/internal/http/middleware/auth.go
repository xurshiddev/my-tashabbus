package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/my-tashabbus/api/internal/http/requestcontext"
	"github.com/my-tashabbus/api/internal/http/response"
	"github.com/my-tashabbus/api/internal/modules/auth"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func RequireAuth(tokens *auth.TokenManager, usersService *users.Service) func(http.Handler) http.Handler {
	return RequireAuthWithDiagnostics(tokens, usersService, nil, false)
}

func RequireAuthWithDiagnostics(tokens *auth.TokenManager, usersService *users.Service, log *slog.Logger, enabled bool) func(http.Handler) http.Handler {
	if log == nil {
		log = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			headerPresent := header != ""
			tokenValue, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || strings.TrimSpace(tokenValue) == "" {
				logAuthDiagnostics(log, enabled, r, "rejected", http.StatusUnauthorized, headerPresent, ok, false, "", "", false)
				response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid access token")
				return
			}
			claims, err := tokens.ParseAccessToken(strings.TrimSpace(tokenValue))
			if err != nil {
				logAuthDiagnostics(log, enabled, r, "rejected", http.StatusUnauthorized, headerPresent, ok, false, "", "", false)
				response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid access token")
				return
			}
			user, err := usersService.GetByID(r.Context(), claims.UserID)
			if err != nil {
				logAuthDiagnostics(log, enabled, r, "rejected", http.StatusUnauthorized, headerPresent, ok, true, claims.UserID.String(), string(claims.Role), false)
				response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid access token")
				return
			}
			if !user.IsActive {
				logAuthDiagnostics(log, enabled, r, "rejected", http.StatusForbidden, headerPresent, ok, true, claims.UserID.String(), string(claims.Role), true)
				response.ErrorCode(w, http.StatusForbidden, "FORBIDDEN", "User is inactive")
				return
			}
			logAuthDiagnostics(log, enabled, r, "accepted", http.StatusOK, headerPresent, ok, true, claims.UserID.String(), string(claims.Role), true)
			ctx := requestcontext.WithCurrentUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func logAuthDiagnostics(
	log *slog.Logger,
	enabled bool,
	r *http.Request,
	result string,
	status int,
	headerPresent bool,
	bearerPrefixValid bool,
	tokenParsed bool,
	userID string,
	role string,
	dbUserFound bool,
) {
	if !enabled {
		return
	}
	attrs := []any{
		"path", r.URL.Path,
		"method", r.Method,
		"result", result,
		"response_status", status,
		"authorization_header_present", headerPresent,
		"bearer_prefix_valid", bearerPrefixValid,
		"token_parsed", tokenParsed,
		"db_user_found", dbUserFound,
	}
	if userID != "" {
		attrs = append(attrs, "user_id", userID)
	}
	if role != "" {
		attrs = append(attrs, "role", role)
	}
	log.Info("auth middleware diagnostics", attrs...)
}
