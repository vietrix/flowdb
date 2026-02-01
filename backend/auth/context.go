package auth

import (
	"context"

	"flowdb/backend/store"
)

type contextKey string

const (
	userKey    contextKey = "user"
	sessionKey contextKey = "session"
)

func WithUser(ctx context.Context, user store.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) (store.User, bool) {
	val := ctx.Value(userKey)
	if val == nil {
		return store.User{}, false
	}
	user, ok := val.(store.User)
	return user, ok
}

func WithSession(ctx context.Context, session store.Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

func SessionFromContext(ctx context.Context) (store.Session, bool) {
	val := ctx.Value(sessionKey)
	if val == nil {
		return store.Session{}, false
	}
	session, ok := val.(store.Session)
	return session, ok
}
