package requestcontext

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type contextKey string

const currentUserKey contextKey = "current_user"
const currentMFYKey contextKey = "current_mfy"

type MFYContext struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

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

func WithCurrentMFY(ctx context.Context, mfy MFYContext) context.Context {
	return context.WithValue(ctx, currentMFYKey, mfy)
}

func CurrentMFY(ctx context.Context) (MFYContext, error) {
	mfy, ok := ctx.Value(currentMFYKey).(MFYContext)
	if !ok {
		return MFYContext{}, errors.New("current mfy missing")
	}
	return mfy, nil
}
