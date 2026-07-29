// Package routes
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/app"
)

func RegisterWorkoutsRoutes(r chi.Router, app *app.Application) {
	r.Route("/workouts", func(r chi.Router) {
		r.Get("/", app.GetWorkouts)
		r.Get("/{id}", app.GetWorkoutByID)
	})
}
