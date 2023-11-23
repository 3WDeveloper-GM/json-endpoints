package config

import (
	"fmt"
	"net/http"
)

func (app *Application) ServerError(w http.ResponseWriter, err error) {
	// trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// app.Logger.Error.Output(2, trace)

	app.Logger.PrintError(err, nil)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *Application) ClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *Application) NotFound(w http.ResponseWriter) {
	app.ClientError(w, http.StatusNotFound)
}

func (app *Application) ErrLog(r *http.Request, err error) {
	app.Logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

func (app *Application) ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	envelope := Envelope{"error": message}

	err := app.JsonWriter(w, status, envelope, nil)
	if err != nil {
		app.ErrLog(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *Application) InternalSErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.ErrLog(r, err)

	message := "the server encountered a problem and could not process the request."
	app.ErrorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *Application) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the resource could not be found."
	app.ErrorResponse(w, r, http.StatusNotFound, message)
}

func (app *Application) MethodNAResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %v method is not allowed.", r.Method)
	app.ErrorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *Application) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *Application) FailedValidationResponse(w http.ResponseWriter, r *http.Request, errorMap map[string]string) {
	app.ErrorResponse(w, r, http.StatusUnprocessableEntity, errorMap)
}

func (app *Application) EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.ErrorResponse(w, r, http.StatusConflict, message)
}

func (app *Application) RateLimitExceedsResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.ErrorResponse(w, r, http.StatusTooManyRequests, message)
}
