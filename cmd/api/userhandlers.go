package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/3WDeveloper-GM/json-endpoints/cmd/config"
	"github.com/3WDeveloper-GM/json-endpoints/internal/data"
	"github.com/3WDeveloper-GM/json-endpoints/internal/validator"
)

func userRegisterPost(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var input struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		err := app.JsonReader(w, r, &input)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
			return
		}

		user := &data.User{
			Name:      input.Name,
			Email:     input.Email,
			Activated: false,
		}

		err = user.Password.Set(input.Password)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
			return
		}

		v := validator.NewValidator()

		if data.ValidateUser(v, user); !v.Valid() {
			app.FailedValidationResponse(w, r, v.Errors)
			return
		}

		err = app.Models.Users.Insert(user)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrDuplicateEmail):
				v.AddError("email", "a user with this email already exists")
				app.FailedValidationResponse(w, r, v.Errors)
			default:
				app.InternalSErrorResponse(w, r, err)
			}
			return
		}

		token, err := app.Models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
			return
		}

		app.Background(func() {

			data := map[string]interface{}{
				"activationToken": token.Plaintext,
				"userID":          user.ID,
			}

			err = app.Mailer.Send(user.Email, "usr_welcome.tmpl", data)
			if err != nil {
				app.Logger.PrintError(err, nil)
			}
		})

		err = app.JsonWriter(w, http.StatusAccepted, config.Envelope{"user": user}, nil)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
		}
	}
}

func activateUserPut(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			TokenPlaintext string `json:"token"`
		}

		err := app.JsonReader(w, r, &input)
		if err != nil {
			app.BadRequestResponse(w, r, err)
			return
		}

		v := validator.NewValidator()

		if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
			app.FailedValidationResponse(w, r, v.Errors)
			return
		}

		user, err := app.Models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				v.AddError("token", "invalid or expired activation token")
				app.FailedValidationResponse(w, r, v.Errors)
			default:
				app.InternalSErrorResponse(w, r, err)
			}
			return
		}

		user.Activated = true

		err = app.Models.Users.Update(user)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrEditConflict):
				app.EditConflictResponse(w, r)
			default:
				app.InternalSErrorResponse(w, r, err)
			}
			return
		}

		err = app.Models.Tokens.DeleteForAllUser(data.ScopeActivation, user.ID)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
			return
		}

		err = app.JsonWriter(w, http.StatusOK, config.Envelope{"user": user}, nil)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
		}
	}
}
