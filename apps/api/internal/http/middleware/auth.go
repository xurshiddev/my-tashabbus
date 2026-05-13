package middleware

import (
	"net/http"
	"strings"

	"github.com/my-tashabbus/api/internal/http/requestcontext"
	"github.com/my-tashabbus/api/internal/http/response"
	"github.com/my-tashabbus/api/internal/modules/auth"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func RequireAuth(tokens *auth.TokenManager, usersService *users.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			tokenValue, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || strings.TrimSpace(tokenValue) == "" {
				response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid access token")
				return
			}
			claims, err := tokens.ParseAccessToken(strings.TrimSpace(tokenValue))
			if err != nil {
				response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid access token")
				return
			}
			user, err := usersService.GetByID(r.Context(), claims.UserID)
			if err != nil {
				response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid access token")
				return
			}
			if !user.IsActive {
				response.ErrorCode(w, http.StatusForbidden, "FORBIDDEN", "User is inactive")
				return
			}
			ctx := requestcontext.WithCurrentUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
