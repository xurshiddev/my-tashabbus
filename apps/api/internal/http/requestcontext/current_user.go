package requestcontext

import (
	"context"
	"errors"

	"github.com/my-tashabbus/api/internal/modules/users"
)

type contextKey string

const currentUserKey contextKey = "current_user"

func WithCurrentUser(ctx context.Context, user users.User) context.Context {
	return context.WithValue(ctx, currentUserKey, user)
}

func CurrentUser(ctx context.Context) (users.User, error) {
	user, ok := ctx.Value(currentUserKey).(users.User)
	if !ok {
		return users.User{}, errors.New("current user missing")
	}
	return user, nil
}
