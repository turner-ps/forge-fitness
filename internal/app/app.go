// Package app
package app

import (
	"log"
	"os"

	"github.com/turner-ps/forge-fitness/internal/store"
	"github.com/turner-ps/forge-fitness/migrations"
)

type Application struct {
	Logger *log.Logger
	Store  *store.Store
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

	dataStore := &store.Store{DB: pgDB}

	app := &Application{
		Logger: logger,
		Store:  dataStore,
	}

	return app, nil
}

func (a *Application) Close() error {
	return a.Store.Close()
}
