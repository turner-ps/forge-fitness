// Package routes
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/app"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	registerFrontendRoutes(r, app)
	registerMeRoutes(r, app)
	registerHealthRoutes(r, app)
	registerWorkoutsRoutes(r, app)
	registerExercisesRoutes(r, app)

	return r
}
