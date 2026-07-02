// Package app
package app

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/turner-ps/forge-fitness/internal/store"
	"github.com/turner-ps/forge-fitness/migrations"
	"github.com/turner-ps/forge-fitness/utils"
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
	err := utils.WriteJSON(w, http.StatusOK, utils.Envelope{"success": "status is available"})
	if err != nil {
		panic(err)
	}
}
