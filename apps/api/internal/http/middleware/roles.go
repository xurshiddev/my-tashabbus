package middleware

import (
	"net/http"

	"github.com/my-tashabbus/api/internal/http/requestcontext"
	"github.com/my-tashabbus/api/internal/http/response"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func RequireAnyRole(allowed ...users.Role) func(http.Handler) http.Handler {
	allowedSet := make(map[users.Role]struct{}, len(allowed))
	for _, role := range allowed {
		allowedSet[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := requestcontext.CurrentUser(r.Context())
			if err != nil {
				response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing current user")
				return
			}
			if _, ok := allowedSet[user.Role]; !ok {
				response.ErrorCode(w, http.StatusForbidden, "FORBIDDEN", "Missing required role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireRole(role users.Role) func(http.Handler) http.Handler {
	return RequireAnyRole(role)
}
