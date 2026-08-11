package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/app"
)

func registerFrontendRoutes(r chi.Router, app *app.Application) {
	r.Get("/", app.FrontendHome)
	r.Get("/ui", app.FrontendHome)
	r.Route("/ui", func(r chi.Router) {
		r.Get("/", app.FrontendHome)
		r.Get("/login", app.FrontendLogin)
		r.Get("/auth-config", app.FrontendAuthConfig)
		r.Get("/workouts", app.FrontendWorkouts)
		r.Get("/workouts/{id}", app.FrontendWorkoutByID)
		r.Get("/exercises", app.FrontendExercises)
		r.Get("/exercises/results", app.FrontendExerciseResults)
		r.Get("/exercises/{id}", app.FrontendExerciseByID)
	})
}
