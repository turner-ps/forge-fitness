// Package app
package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/turner-ps/forge-fitness/internal/store"
	"github.com/turner-ps/forge-fitness/migrations"
)

type Application struct {
	Logger *log.Logger
	DB     *sql.DB
}

func NewApplication() (*Application, error) {
	pgDB, err := store.Open()
	if err != nil {
		return nil, err
	}

	err = store.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		return nil, err
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &Application{
		Logger: logger,
		DB:     pgDB,
	}

	return app, nil
}

func (a *Application) Heartbeat(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprint(w, "Status is available\n")
	if err != nil {
		panic(err)
	}
}
