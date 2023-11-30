package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/3WDeveloper-GM/json-endpoints/cmd/config"
)

func serve(app *config.Application) error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", app.Config.Port),
		Handler:      getRoutes(app),
		IdleTimeout:  2 * time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		app.Logger.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.Logger.PrintInfo("completing background tasks at port", map[string]string{
			"addr": server.Addr,
		})

		app.Wait()
		shutdownError <- nil

	}()

	app.Logger.PrintInfo("Starting server with the following ", map[string]string{
		"addr":    server.Addr,
		"env":     app.Config.Mode,
		"version": app.Config.Version,
	})

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.Logger.PrintInfo("stopped server", map[string]string{
		"addr": server.Addr,
	})

	return nil
}
