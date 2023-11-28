package main

import (
	"errors"
	"net/http"

	"github.com/3WDeveloper-GM/json-endpoints/cmd/config"
	"github.com/3WDeveloper-GM/json-endpoints/internal/data"
	"github.com/3WDeveloper-GM/json-endpoints/internal/validator"
)

func usercreatePost(app *config.Application) http.HandlerFunc {
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

		err = app.JsonWriter(w, http.StatusCreated, config.Envelope{"user": user}, nil)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
		}
	}
}
