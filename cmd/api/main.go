package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	"github.com/3WDeveloper-GM/json-endpoints/cmd/config"
	"github.com/3WDeveloper-GM/json-endpoints/internal/jsonlog"
)

const version = "1.0.0"

var appcfg, applog, appmodel, appsmtp, db = initConfig()

// initializes the configuration struct at runtime.
func initConfig() (*config.AppConfig, *config.AppLoggers, *config.AppModels, *config.AppSMTP, *sql.DB) {

	//loggers first
	applog := &config.AppLoggers{}
	applog.SetStructConfig(os.Stdout, jsonlog.LevelInfo)
	applog.PrintInfo("logger object initialized", nil)

	//configuration flags second
	appcfg := &config.AppConfig{}
	appcfg.SetStructConfig(version)
	flag.Parse()
	applog.PrintInfo("config object correctly configured", nil)

	//smtp third
	appsmtp := &config.AppSMTP{}
	appsmtp.SetStructConfig(appcfg)
	applog.PrintInfo("SMTP mailer object initialized", nil)

	//finally DB and models
	db, err := openDB(appcfg)
	if err != nil {
		applog.PrintFatal(err, nil)
	}

	appmodel := &config.AppModels{}
	appmodel.SetStructConfig(db)
	applog.PrintInfo("database connection pool established", nil)

	return appcfg, applog, appmodel, appsmtp, db
}

func main() {

	defer db.Close()             //deferring the database shutdown when the program terminates
	app := &config.Application{} //Getting an application struct

	app.SetStructConfig(appcfg, applog, appmodel, appsmtp)   //configuring the app struct in a single data structure
	applog.PrintInfo("Application object initialized.", nil) //confirmation message

	err := serve(app)
	if err != nil {
		applog.PrintFatal(err, nil)
	}
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
