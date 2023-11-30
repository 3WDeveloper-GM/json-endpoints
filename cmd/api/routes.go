package main

import (
	"github.com/3WDeveloper-GM/json-endpoints/cmd/config"
	"github.com/go-chi/chi/v5"
)

func getRoutes(app *config.Application) *chi.Mux {
	r := chi.NewMux()

	r.Use(app.RecoverPanic)
	r.Use(app.RateLimiter)
	r.Use(app.Authenticate)
	r.NotFound(app.NotFoundResponse)

	r.MethodNotAllowed(app.MethodNAResponse)

	// GET routes
	r.Get("/v1/healthcheck", healthcheckhandler(app))                            //Display application information in JSON
	r.Get("/v1/movies", app.RequireActivatedUsr(listMoviesHandlerGet(app)))      //Display a list of movies in the DB
	r.Get("/v1/movies/{id}", app.RequireActivatedUsr(showMoviesHandlerGet(app))) //Display a particular movie in the DB

	r.Post("/v1/movies", app.RequireActivatedUsr(createMovieHandlerPost(app))) //Add some movie to the DB using a JSON request body
	r.Post("/v1/users", userRegisterPost(app))                                 //Add user to the DB using a JSON request body
	r.Post("/v1/users/authentication", createAuthenticationTokenPost(app))

	r.Patch("/v1/movies/{id}", app.RequireActivatedUsr(movielistHandlerPatch(app))) //Patching some of the resources in the DB

	r.Delete("/v1/movies/{id}", app.RequireActivatedUsr(movieupdateHandlerDelete(app))) //Deleting an entry in the DB

	r.Put("/v1/users/activated", activateUserPut(app))

	return r
}
