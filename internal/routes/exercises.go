// Package routes
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/app"
)

func registerExercisesRoutes(r chi.Router, app *app.Application) {
	r.Route("/exercises", func(r chi.Router) {
		r.Get("/", app.GetExercises)
		r.Get("/{id}", app.GetExerciseByID)
	})
}
