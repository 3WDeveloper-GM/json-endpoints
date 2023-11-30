package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/3WDeveloper-GM/json-endpoints/cmd/config"
	"github.com/3WDeveloper-GM/json-endpoints/internal/data"
	"github.com/3WDeveloper-GM/json-endpoints/internal/validator"
)

func createAuthenticationTokenPost(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		err := app.JsonReader(w, r, &input)
		if err != nil {
			app.BadRequestResponse(w, r, err)
			return
		}

		v := validator.NewValidator()

		data.ValidateEmail(v, input.Email)
		data.ValidatePasswordPlaintext(v, input.Password)

		if !v.Valid() {
			app.FailedValidationResponse(w, r, v.Errors)
			return
		}

		user, err := app.Models.Users.GetByEmail(input.Email)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.InvalidCredentialsResponse(w, r)
			default:
				app.InternalSErrorResponse(w, r, err)
			}
			return
		}

		match, err := user.Password.Matches(input.Password)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
			return
		}

		if !match {
			app.InvalidCredentialsResponse(w, r)
			return
		}

		token, err := app.Models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
			return
		}

		err = app.JsonWriter(w, http.StatusCreated, config.Envelope{"auth_token": token}, nil)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
		}
	}
}
