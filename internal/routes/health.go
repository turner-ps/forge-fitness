// Package routes
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/app"
)

func registerHealthRoutes(r chi.Router, app *app.Application) {
	r.Get("/heartbeat", app.Heartbeat)
}
