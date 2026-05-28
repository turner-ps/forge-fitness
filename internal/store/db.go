// Package store
package store

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const defaultDSN = "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"

func Open() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDSN
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("db:open %w", err)
	}

	if err := pingWithRetry(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("db:open %w", err)
	}

	fmt.Println("connected to database")
	return db, nil
}

func pingWithRetry(db *sql.DB) error {
	var err error

	for attempt := 1; attempt <= 10; attempt++ {
		err = db.Ping()
		if err == nil {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return err
}

func MigrateFS(db *sql.DB, migrationFS fs.FS, dir string) error {
	goose.SetBaseFS(migrationFS)
	defer func() {
		goose.SetBaseFS(nil)
	}()

	return Migrate(db, dir)
}

func Migrate(db *sql.DB, dir string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("migrate %w", err)
	}

	err = goose.Up(db, dir)
	if err != nil {
		return fmt.Errorf("goose up %w", err)
	}

	return nil
}
