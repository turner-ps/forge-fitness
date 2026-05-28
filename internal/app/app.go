// Package app
package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Application struct {
	Logger *log.Logger
}

func NewApplication() (*Application, error) {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &Application{
		Logger: logger,
	}

	return app, nil
}

func (a *Application) Heartbeat(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprint(w, "Status is available\n")
	if err != nil {
		panic(err)
	}
}
