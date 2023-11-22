package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/3WDeveloper-GM/json-app/cmd/config"
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
	applog.SetStructConfig()                          //setting the loggers

	// database initialization
	db, err := openDB(appcfg)
	if err != nil {
		applog.Error.Fatal(err)
	}

	defer db.Close()

	appmodel.SetStructConfig(db)                  //setting the model struct
	app.SetStructConfig(appcfg, applog, appmodel) //configuring the app struct in a single data structure

	applog.Info.Print("database connection pool established")

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.Port),
		Handler:      getRoutes(app),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	applog.Info.Printf("Starting server on port :%v", app.Config.Port)
	err = srv.ListenAndServe()
	applog.Error.Fatal(err)

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
