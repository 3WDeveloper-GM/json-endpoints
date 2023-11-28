package main

import (
	"github.com/3WDeveloper-GM/json-endpoints/cmd/config"
	"github.com/go-chi/chi/v5"
)

func getRoutes(app *config.Application) *chi.Mux {
	r := chi.NewMux()

	r.Use(app.RecoverPanic)
	r.Use(app.RateLimiter)
	r.NotFound(app.NotFoundResponse)

	r.MethodNotAllowed(app.MethodNAResponse)

	// GET routes
	r.Get("/v1/healthcheck", healthcheckhandler(app))   //Display application information in JSON
	r.Get("/v1/movies", listMoviesHandlerGet(app))      //Display a list of movies in the DB
	r.Get("/v1/movies/{id}", showMoviesHandlerGet(app)) //Display a particular movie in the DB

	r.Post("/v1/movies", createMovieHandlerPost(app)) //Add some movie to the DB using a JSON request body
	r.Post("/v1/users", usercreatePost(app))          //Add user to the DB using a JSON request body

	r.Patch("/v1/movies/{id}", movielistHandlerPatch(app)) //Patching some of the resources in the DB

	r.Delete("/v1/movies/{id}", movieupdateHandlerDelete(app)) //Deleting an entry in the DB

	return r
}
