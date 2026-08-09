// Package routes
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/app"
)

func registerWorkoutsRoutes(r chi.Router, app *app.Application) {
	r.Route("/users/{userID}", func(r chi.Router) {
		r.Get("/workout-sessions", app.GetUserWorkoutSessions)
		r.Get("/workout-sessions/{sessionID}", app.GetWorkoutSessionByID)

		r.Route("/workouts", func(r chi.Router) {
			r.Get("/", app.GetWorkouts)
			r.Post("/", app.CreateWorkout)
			r.Get("/{id}", app.GetWorkoutByID)
			r.Get("/{id}/exercises", app.GetWorkoutExercises)
			r.Post("/{id}/exercises", app.AddExerciseToWorkout)
			r.Post("/{id}/exercises/bulk", app.AddExercisesToWorkout)
			r.Get("/{id}/sessions", app.GetWorkoutSessions)
			r.Post("/{id}/sessions", app.CreateWorkoutSession)
		})
	})
}
