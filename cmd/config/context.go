package config

import (
	"context"
	"net/http"

	"github.com/3WDeveloper-GM/json-endpoints/internal/data"
)

type contextkey string

const userCtxKey = contextkey("user")

func (app *Application) ContextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userCtxKey, user)
	return r.WithContext(ctx)
}

func (app *Application) ContextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userCtxKey).(*data.User)
	if !ok {
		panic("missing value in request context")
	}
	return user
}
