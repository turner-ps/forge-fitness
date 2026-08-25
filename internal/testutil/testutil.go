// Package testutil provides reusable helpers for integration tests that
// exercise the API against a real Postgres test database via a fake token
// verifier. It is intentionally independent of internal/app so it can be
// shared across handler, store, and template tests.
package testutil

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/turner-ps/forge-fitness/internal/auth"
	"github.com/turner-ps/forge-fitness/internal/store"
	"github.com/turner-ps/forge-fitness/migrations"
)

var migrateOnce sync.Once

// DSN returns the test database connection string. Override with
// TEST_DATABASE_URL; defaults to the docker-compose test_db on port 5433.
func DSN(t *testing.T) string {
	t.Helper()
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}
	return "host=localhost user=postgres password=postgres dbname=postgres port=5433 sslmode=disable"
}

// OpenDB opens the test database and applies migrations a single time across
// the test run. The connection is closed via t.Cleanup.
func OpenDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", DSN(t))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	migrateOnce.Do(func() {
		if err := store.MigrateFS(db, migrations.FS, "."); err != nil {
			t.Fatalf("migrate test db: %v", err)
		}
	})

	return db
}

// Truncate clears all user-owned tables between tests, resetting identity
// sequences so each test starts from a clean slate.
func Truncate(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		TRUNCATE TABLE
			workout_session_set,
			workout_session_exercise,
			workout_session,
			workout_exercise,
			workout,
			app_user
		RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("truncate test tables: %v", err)
	}
}

// FakeVerifier implements auth.TokenVerifier without network access. If
// Reject is non-nil it always fails; otherwise tokens missing from Identities
// are treated as invalid.
type FakeVerifier struct {
	Identities map[string]*auth.Identity
	Reject     error
}

// VerifyIDToken resolves token to a predefined identity.
func (f *FakeVerifier) VerifyIDToken(_ context.Context, token string) (*auth.Identity, error) {
	if f.Reject != nil {
		return nil, f.Reject
	}
	identity, ok := f.Identities[token]
	if !ok {
		return nil, errors.New("invalid token")
	}
	return identity, nil
}

// FirebaseIdentity builds a Firebase-backed identity for tests.
func FirebaseIdentity(uid, email string) *auth.Identity {
	return &auth.Identity{
		Provider: auth.ProviderFirebase,
		Subject:  uid,
		Email:    email,
	}
}

// DoRequest issues an HTTP request against handler, optionally attaching an
// Authorization Bearer token and a JSON body, and returns the recorder.
func DoRequest(t *testing.T, handler http.Handler, method, target, token string, body string) *httptest.ResponseRecorder {
	t.Helper()

	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// Itoa formats an int64 for use in URL path segments.
func Itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
