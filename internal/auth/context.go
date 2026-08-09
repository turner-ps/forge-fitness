package auth

import (
	"context"
	"errors"

	"github.com/turner-ps/forge-fitness/internal/store"
)

type userContextKey struct{}

var ErrMissingUser = errors.New("missing authenticated user")

func ContextWithUser(ctx context.Context, user *store.User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func UserFromContext(ctx context.Context) (*store.User, bool) {
	user, ok := ctx.Value(userContextKey{}).(*store.User)
	return user, ok && user != nil
}

func RequireUser(ctx context.Context) (*store.User, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, ErrMissingUser
	}

	return user, nil
}
