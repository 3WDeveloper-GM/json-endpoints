package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/3WDeveloper-GM/json-endpoints/cmd/config"
	"github.com/3WDeveloper-GM/json-endpoints/internal/data"
	"github.com/3WDeveloper-GM/json-endpoints/internal/validator"
)

func healthcheckhandler(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		envelope := &config.Envelope{
			"status": "available",
			"sys_info": map[string]string{
				"environment": app.Config.Mode,
				"version":     app.Config.Version,
			},
		}

		err := app.JsonWriter(w, http.StatusOK, *envelope, nil)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
		}
	}
}

func listMoviesHandlerGet(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var input struct {
			Title  string
			Genres []string
			data.Filters
		}

		v := validator.NewValidator()

		qs := r.URL.Query()

		input.Title = app.ReadStrings(qs, "title", "")
		input.Genres = app.ReadCSV(qs, "genres", []string{})

		input.Filters.Page = app.ReadInt(qs, "page", 1, v)
		input.Filters.PageSize = app.ReadInt(qs, "page_size", 20, v)

		input.Filters.Sort = app.ReadStrings(qs, "sort", "id")

		input.Filters.SortSafeList = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

		if data.ValidateFilters(v, input.Filters); !v.Valid() {
			app.FailedValidationResponse(w, r, v.Errors)
			return
		}

		movies, err := app.Models.Movies.GetAll(input.Title, input.Genres, input.Filters)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
			return
		}

		err = app.JsonWriter(w, http.StatusOK, config.Envelope{"movies": movies}, nil)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
		}

	}
}

func showMoviesHandlerGet(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := app.ReadIDparameter(w, r)
		if err != nil {
			app.NotFoundResponse(w, r)
			return
		}

		movie, err := app.Models.Movies.Get(id)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				app.NotFoundResponse(w, r)
			default:
				app.InternalSErrorResponse(w, r, err)
			}
			return
		}

		err = app.JsonWriter(w, http.StatusOK, config.Envelope{"movie": movie}, nil)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
		}

	}
}

func createMovieHandlerPost(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Title   string       `json:"title"`
			Year    int32        `json:"year"`
			Runtime data.Runtime `json:"runtime"`
			Genres  []string     `json:"genres"`
		}

		err := app.JsonReader(w, r, &input)
		if err != nil {
			app.BadRequestResponse(w, r, err)
			return
		}

		movie := &data.Movie{
			Title:   input.Title,
			Year:    input.Year,
			Runtime: input.Runtime,
			Genres:  input.Genres,
		}

		//input validation

		v := validator.NewValidator()

		// must be provided errors

		// var message = "must be provided"
		// v.Check(input.Title != "", "title", message)
		// v.Check(input.Year != 0, "year", message)
		// v.Check(input.Genres != nil, "genres", message)
		// v.Check(input.Runtime != 0, "runtime", message)

		// // title errors and year errorsa

		// // Generic text messages
		// var lessThanmessage = "must be less than"
		// var moreThanmessage = "must be more than"

		// // Dates for cleaner code.
		// var titleMaxCharacters = 500
		// var firstFilmTapeYear = 1888

		// var presentYear = time.Now().Year()

		// var minimumGenres = 1
		// var maxGenres = 5

		// v.Check(len(input.Title) < titleMaxCharacters, "title", fmt.Sprintf(lessThanmessage+" %d characters long", titleMaxCharacters))
		// v.Check(input.Year >= int32(firstFilmTapeYear), "year", fmt.Sprintf(moreThanmessage+" %d. the first film is dated in %d", firstFilmTapeYear, firstFilmTapeYear))
		// v.Check(input.Year <= int32(presentYear), "year", fmt.Sprintf(moreThanmessage+" %v , film cannot be dated in the future.", int32(presentYear)))
		// v.Check(input.Runtime > 0, "runtime", fmt.Sprint(moreThanmessage+" 0, it must be a positive integer."))
		// v.Check(len(input.Genres) >= minimumGenres, "genres", fmt.Sprintf(moreThanmessage+" %d genre, please add genres", minimumGenres))
		// v.Check(len(input.Genres) <= maxGenres, "genres", fmt.Sprintf(lessThanmessage+" %d genres, please remove genres", maxGenres))

		// // Unique characters

		// var messageUnique = "must have unique values."
		// v.Check(validator.Unique(input.Genres), "genres", messageUnique)

		if data.ValidateMovie(v, movie); !v.Valid() {
			app.FailedValidationResponse(w, r, v.Errors)
			return
		}

		err = app.Models.Movies.Insert(movie)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
			return
		}

		headers := make(http.Header)
		headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

		err = app.JsonWriter(w, http.StatusOK, config.Envelope{"movie": movie}, headers)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
			return
		}

	}
}

func movieupdateHandlerDelete(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := app.ReadIDparameter(w, r)
		if err != nil {
			app.NotFoundResponse(w, r)
			return
		}

		err = app.Models.Movies.Delete(id)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				app.NotFoundResponse(w, r)
			default:
				app.InternalSErrorResponse(w, r, err)
			}
			return
		}

		err = app.JsonWriter(w, http.StatusOK, config.Envelope{"movie": fmt.Sprintf("movie at id %v deleted succesfully", id)}, nil)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
		}
	}
}

func movielistHandlerPatch(app *config.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Title   *string       `json:"title"`
			Year    *int32        `json:"year"`
			Runtime *data.Runtime `json:"runtime"`
			Genres  []string      `json:"genres"`
		}

		id, err := app.ReadIDparameter(w, r)
		if err != nil {
			app.NotFoundResponse(w, r)
			return
		}

		err = app.JsonReader(w, r, &input)
		if err != nil {
			app.BadRequestResponse(w, r, err)
			return
		}

		movie, err := app.Models.Movies.Get(id)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				app.NotFoundResponse(w, r)
			default:
				app.InternalSErrorResponse(w, r, err)
			}
			return
		}

		if input.Title != nil {
			movie.Title = *input.Title
		}

		if input.Year != nil {
			movie.Year = *input.Year
		}

		if input.Runtime != nil {
			movie.Runtime = *input.Runtime
		}

		if input.Genres != nil {
			movie.Genres = input.Genres
		}

		v := validator.NewValidator()

		if data.ValidateMovie(v, movie); !v.Valid() {
			app.FailedValidationResponse(w, r, v.Errors)
		}

		err = app.Models.Movies.Update(movie)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrEditConflict):
				app.EditConflictResponse(w, r)
			default:
				app.InternalSErrorResponse(w, r, err)
			}
			return
		}

		headers := make(http.Header)
		headers.Set("location", fmt.Sprintf("/v1/movies/%d", id))

		err = app.JsonWriter(w, http.StatusOK, config.Envelope{"movie": movie}, headers)
		if err != nil {
			app.InternalSErrorResponse(w, r, err)
			return
		}
	}
}
