// Package routes
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/app"
)

func registerWorkoutsRoutes(r chi.Router, app *app.Application) {
	r.Group(func(r chi.Router) {
		r.Use(app.RequireAuth)

		r.Get("/workout-sessions", app.GetUserWorkoutSessions)
		r.Get("/workout-sessions/{sessionID}", app.GetWorkoutSessionByID)

		r.Get("/workouts", app.GetWorkouts)
		r.Post("/workouts", app.CreateWorkout)
		r.Get("/workouts/{id}", app.GetWorkoutByID)
		r.Post("/workouts/{id}/exercises", app.AddExerciseToWorkout)
		r.Get("/workouts/{id}/sessions", app.GetWorkoutSessions)
		r.Post("/workouts/{id}/sessions", app.CreateWorkoutSession)
	})
}
