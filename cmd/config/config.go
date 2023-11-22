package config

import (
	"database/sql"
	"log"

	"github.com/3WDeveloper-GM/json-app/cmd/pkg/loggers"
	"github.com/3WDeveloper-GM/json-app/internal/data"
	_ "github.com/lib/pq"
)

type SetStructConfig interface {
}

// type set interface {
// }

// returns an empty instance of the struct for later configuration
type NewStruct interface {
}

type Application struct {
	Config *AppConfig
	Logger *AppLoggers
	Models *AppModels
}

type AppConfig struct {
	Port     int
	Version  string
	Mode     string
	Database struct {
		Dsn string
	}
}

type AppLoggers struct {
	Info  *log.Logger
	Error *log.Logger
}

type AppModels struct {
	data.Models
}

// func (appcfg *AppConfig) set(trait, value interface{}) {
// 	trait = value
// }

func (appcfg *AppConfig) setPort(port int) {
	appcfg.Port = port
}

func (appcfg *AppConfig) setVersion(version string) {
	appcfg.Version = version
}

func (appcfg *AppConfig) setEnvironment(environment string) {
	appcfg.Mode = environment
}

func (appcfg *AppConfig) setDatabase(database string) {
	appcfg.Database.Dsn = database
}

// SetStructConfig interface for the main application struct
func (appConfig *AppConfig) SetStructConfig(port int, version, db, environment string) {
	//string flags

	// appConfig.set(appConfig.Database, db)      //set database
	// appConfig.set(appConfig.Mode, environment) //set environment
	// appConfig.set(appConfig.Version, version)  //set version

	appConfig.setDatabase(db)             //set database
	appConfig.setEnvironment(environment) //set environment
	appConfig.setVersion(version)         //set version

	//integer flags
	appConfig.setPort(port) //set port

	// appConfig.set(appConfig.Port, port) //set port
}

// Interface for configuring the loggers, it just needs to be called
func (appLog *AppLoggers) SetStructConfig() {
	appLog.Error = loggers.ErrorLog()
	appLog.Info = loggers.InfoLog()
}

// Interface for configuring the model struct, for the CRUD operations
func (appModel *AppModels) SetStructConfig(db *sql.DB) {
	appModel.Movies = data.MovieModel{DB: db}
}

// Interface for getting the configuration of the main application struct
func (app *Application) SetStructConfig(appcfg *AppConfig, applog *AppLoggers, appModel *AppModels) {
	app.Config = appcfg
	app.Logger = applog
	app.Models = appModel
}
