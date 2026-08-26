package app

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/auth"
	"github.com/turner-ps/forge-fitness/internal/store"
	"github.com/turner-ps/forge-fitness/internal/testutil"
)

// newTestApp wires an Application to a fresh truncated test database and
// returns the application plus a router for the authenticated endpoints.
func newTestApp(t *testing.T, verifier auth.TokenVerifier) (*Application, http.Handler) {
	t.Helper()

	db := testutil.OpenDB(t)
	testutil.Truncate(t, db)

	application := &Application{
		Logger:        log.New(os.Stdout, "", 0),
		Store:         &store.Store{DB: db},
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
		r.Patch("/workouts/{id}", application.UpdateWorkout)
		r.Delete("/workouts/{id}", application.DeleteWorkout)
		r.Get("/workout-sessions", application.GetUserWorkoutSessions)
		r.Get("/workout-sessions/{sessionID}", application.GetWorkoutSessionByID)
	})
	return r
}
