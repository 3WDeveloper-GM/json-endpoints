package config

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/3WDeveloper-GM/json-endpoints/internal/data"
	"github.com/3WDeveloper-GM/json-endpoints/internal/validator"
	"golang.org/x/time/rate"
)

func (app *Application) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("connection", "close")

				app.InternalSErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *Application) RateLimiter(next http.Handler) http.Handler {

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if app.Config.Limiter.Enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.InternalSErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.Config.Limiter.Rps), app.Config.Limiter.Burst),
				}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.RateLimitExceedsResponse(w, r)
				return
			}

			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func (app *Application) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.ContextSetUser(r, data.AnonUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.InvalidCredentialsResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.NewValidator()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.InvalidCredentialsResponse(w, r)
			return
		}

		user, err := app.Models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.InvalidAuthenticationTokenResponse(w, r)
			default:
				app.InternalSErrorResponse(w, r, err)
			}
			return
		}

		r = app.ContextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}

func (app *Application) RequireActivatedUsr(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.ContextGetUser(r)

		if user.IsAnonymous() {
			app.AuthenticationRequiredResponse(w, r)
			return
		}

		if !user.Activated {
			app.InactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
