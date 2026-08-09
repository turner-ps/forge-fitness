package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/app"
)

func registerMeRoutes(r chi.Router, app *app.Application) {
	r.Route("/me", func(r chi.Router) {
		r.Use(app.RequireAuth)
		r.Get("/", app.GetMe)
	})
}
