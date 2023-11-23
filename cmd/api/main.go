package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/3WDeveloper-GM/json-endpoints/cmd/config"
	"github.com/3WDeveloper-GM/json-endpoints/internal/jsonlog"
)

var env = flag.String("env", "development", "Environment (development|production|staging)")
var port = flag.Int("port", 4000, "Defines the port for the web server.")
var db = flag.String("database-dsn", os.Getenv("TESTING_DSN"), "PostgreSQL DSN")

const version = "1.0.0"

func init() {
	flag.Parse()
}

func main() {

	app := &config.Application{}    //Getting an application struct
	appcfg := &config.AppConfig{}   //Getting a configuration struct
	applog := &config.AppLoggers{}  //Getting a loggers struct
	appmodel := &config.AppModels{} //Getting a model struct for CRUD operations

	// Setting the configuration of the main app components
	appcfg.SetStructConfig(*port, version, *db, *env) //setting the configuration struct
	applog.SetStructConfig(os.Stdout, jsonlog.LevelInfo)
	applog.PrintInfo("configuration object correctly configured", nil)
	applog.PrintInfo("logger configuration correctly established", nil)
	// database initialization
	db, err := openDB(appcfg)
	if err != nil {
		applog.Logger.PrintFatal(err, nil)
	}

	defer db.Close()

	appmodel.SetStructConfig(db)                  //setting the model struct
	app.SetStructConfig(appcfg, applog, appmodel) //configuring the app struct in a single data structure
	applog.PrintInfo("connection pool established", nil)

	applog.PrintInfo("starting server with configuration:", map[string]string{
		"addr":    fmt.Sprint(appcfg.Port),
		"env":     appcfg.Mode,
		"version": appcfg.Version,
	})
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.Port),
		Handler:      getRoutes(app),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	err = srv.ListenAndServe()
	applog.PrintFatal(err, nil)

}

func openDB(appcfg *config.AppConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", appcfg.Database.Dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
