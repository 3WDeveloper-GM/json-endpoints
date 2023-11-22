package loggers

import (
	"log"
	"os"
)

func InfoLog() *log.Logger {
	return log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

}

func ErrorLog() *log.Logger {
	return log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
}
