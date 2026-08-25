package app

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/auth"
	"github.com/turner-ps/forge-fitness/internal/store"
	"github.com/turner-ps/forge-fitness/migrations"
)

var migrateOnce sync.Once

type fakeVerifier struct {
	identities map[string]*auth.Identity
	reject     error
}

func (f *fakeVerifier) VerifyIDToken(_ context.Context, token string) (*auth.Identity, error) {
	if f.reject != nil {
		return nil, f.reject
	}
	identity, ok := f.identities[token]
	if !ok {
		return nil, errors.New("invalid token")
	}
	return identity, nil
}

func testDSN(t *testing.T) string {
	t.Helper()
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}
	return "host=localhost user=postgres password=postgres dbname=postgres port=5433 sslmode=disable"
}

func newTestApp(t *testing.T, verifier auth.TokenVerifier) (*Application, http.Handler) {
	t.Helper()

	db, err := sql.Open("pgx", testDSN(t))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	migrateOnce.Do(func() {
		if err := store.MigrateFS(db, migrations.FS, "."); err != nil {
			t.Fatalf("migrate test db: %v", err)
		}
	})

	truncateTables(t, db)

	dataStore := &store.Store{DB: db}
	application := &Application{
		Logger:        log.New(os.Stdout, "", 0),
		Store:         dataStore,
		TokenVerifier: verifier,
	}

	return application, testRouter(application)
}

func testRouter(application *Application) http.Handler {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(application.RequireAuth)
		r.Get("/me", application.GetMe)
		r.Get("/workouts", application.GetWorkouts)
		r.Post("/workouts", application.CreateWorkout)
		r.Get("/workouts/{id}", application.GetWorkoutByID)
		r.Get("/workout-sessions", application.GetUserWorkoutSessions)
		r.Get("/workout-sessions/{sessionID}", application.GetWorkoutSessionByID)
	})
	return r
}

func truncateTables(t *testing.T, db *sql.DB) {
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

func doRequest(t *testing.T, handler http.Handler, method, target, token string, body string) *httptest.ResponseRecorder {
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

func firebaseIdentity(uid, email string) *auth.Identity {
	return &auth.Identity{
		Provider: auth.ProviderFirebase,
		Subject:  uid,
		Email:    email,
	}
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
