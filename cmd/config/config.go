package config

import (
	"database/sql"
	"flag"
	"io"
	"os"
	"sync"
	"time"

	"github.com/3WDeveloper-GM/json-endpoints/internal/data"
	"github.com/3WDeveloper-GM/json-endpoints/internal/jsonlog"
	"github.com/3WDeveloper-GM/json-endpoints/internal/mailer"
	"github.com/go-mail/mail/v2"
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
	Mailer *AppSMTP
	sync.WaitGroup
}

type AppConfig struct {
	Port     int
	Version  string
	Mode     string
	Database struct {
		Dsn          string
		MaxIdleConns int
		MaxOpenConns int
		MaxIdleTime  string
	}
	Limiter struct {
		Rps     float64
		Burst   int
		Enabled bool
	}
	SMTP struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}
}

type AppLoggers struct {
	jsonlog.Logger
}

type AppSMTP struct {
	mailer.Mailer
}

type AppModels struct {
	data.Models
}

// SetStructConfig interface for the main application struct
func (appcfg *AppConfig) SetStructConfig(version string) {
	//port, environment, and version
	appcfg.Version = version
	flag.IntVar(&appcfg.Port, "port", 4000, "API server port")
	flag.StringVar(&appcfg.Mode, "env", "development", "Environment (development|staging|production)")

	//database configurations
	flag.StringVar(&appcfg.Database.Dsn, "db-dsn", os.Getenv("TESTING_DSN"), "PostgreSQL DSN")

	flag.IntVar(&appcfg.Database.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.IntVar(&appcfg.Database.MaxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.StringVar(&appcfg.Database.MaxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	//rate limiter configurations
	flag.Float64Var(&appcfg.Limiter.Rps, "rps", 2, "rate limiter maximum requests per second")
	flag.IntVar(&appcfg.Limiter.Burst, "burst", 4, "rate limiter maximum burst")
	flag.BoolVar(&appcfg.Limiter.Enabled, "limited-enabled", true, "rate limiter enabler")

	//SMTP configuration flags
	flag.StringVar(&appcfg.SMTP.Host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&appcfg.SMTP.Port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&appcfg.SMTP.Username, "smtp-username", "0b6416a1d5a92d", "SMTP username")
	flag.StringVar(&appcfg.SMTP.Password, "smtp-password", "756080f5f1c2c3", "SMTP password")
	flag.StringVar(&appcfg.SMTP.Sender, "smtp-sender", "Greenlight <no-reply@greenlight.3wdevel.net>", "SMTP sender address")

}

func (appsmtp *AppSMTP) SetStructConfig(appcfg *AppConfig) {
	appsmtp.Dialer = mail.NewDialer(appcfg.SMTP.Host, appcfg.SMTP.Port, appcfg.SMTP.Username, appcfg.SMTP.Password)
	appsmtp.Dialer.Timeout = 5 * time.Second

	appsmtp.Sender = appcfg.SMTP.Sender
}

func (applog *AppLoggers) SetStructConfig(out io.Writer, min jsonlog.Level) {
	applog.Out = out
	applog.Minlevel = min
}

// Interface for configuring the model struct, for the CRUD operations
func (appModel *AppModels) SetStructConfig(db *sql.DB) {
	appModel.Movies = data.MovieModel{DB: db}
	appModel.Users = data.UserModel{DB: db}
	appModel.Tokens = data.TokenModel{DB: db}
}

// Interface for getting the configuration of the main application struct
func (app *Application) SetStructConfig(appcfg *AppConfig, applog *AppLoggers, appModel *AppModels, appsmtp *AppSMTP) {
	app.Config = appcfg
	app.Logger = applog
	app.Models = appModel
	app.Mailer = appsmtp
}
