// Package routes
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/app"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/heartbeat", app.Heartbeat)
	r.Get("/workout", app.GetWorkout)

	return r
}
