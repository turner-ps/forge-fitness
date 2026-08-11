// Package app
package app

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/turner-ps/forge-fitness/internal/auth"
	"github.com/turner-ps/forge-fitness/internal/store"
	"github.com/turner-ps/forge-fitness/migrations"
)

type Application struct {
	Logger            *log.Logger
	Store             *store.Store
	TokenVerifier     auth.TokenVerifier
	FirebaseWebConfig auth.FirebaseWebConfig
}

func NewApplication() (*Application, error) {
	pgDB, err := store.Open()
	if err != nil {
		return nil, fmt.Errorf("db:init %w", err)
	}

	defer func() {
		if err != nil {
			_ = pgDB.Close()
		}
	}()

	err = store.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("db:migrate %w", err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	dataStore := &store.Store{DB: pgDB}

	var tokenVerifier auth.TokenVerifier
	var firebaseWebConfig auth.FirebaseWebConfig
	authEnabled, err := auth.EnabledFromEnv()
	if err != nil {
		return nil, fmt.Errorf("auth:config %w", err)
	}

	if authEnabled {
		tokenVerifier, err = auth.NewFirebaseVerifierFromEnv(context.Background())
		if err != nil {
			return nil, fmt.Errorf("auth:init firebase %w", err)
		}

		firebaseWebConfig, err = auth.FirebaseWebConfigFromEnv()
		if err != nil {
			return nil, fmt.Errorf("auth:web config %w", err)
		}
	}

	app := &Application{
		Logger:            logger,
		Store:             dataStore,
		TokenVerifier:     tokenVerifier,
		FirebaseWebConfig: firebaseWebConfig,
	}

	return app, nil
}

func (a *Application) Close() error {
	return a.Store.Close()
}
