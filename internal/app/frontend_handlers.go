package app

import (
	"database/sql"
	"embed"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/store"
)

//go:embed templates/*.html
var frontendTemplateFS embed.FS

var frontendTemplates = template.Must(template.ParseFS(frontendTemplateFS, "templates/*.html"))

type frontendData struct {
	Title     string
	Active    string
	Search    string
	Workouts  []store.Workout
	Workout   *store.Workout
	Exercises []store.Exercise
	Exercise  *store.Exercise
	Error     string
}

func (a *Application) FrontendHome(w http.ResponseWriter, r *http.Request) {
	workouts, err := a.Store.GetWorkouts(r.Context())
	if err != nil {
		a.frontendServerError(w, r, err)
		return
	}

	exercises, err := a.Store.GetExercises(r.Context(), "", 6)
	if err != nil {
		a.frontendServerError(w, r, err)
		return
	}

	a.renderFrontend(w, r, "home-page", "home-content", frontendData{
		Title:     "Forge Fitness",
		Active:    "home",
		Workouts:  workouts,
		Exercises: exercises,
	})
}

func (a *Application) FrontendWorkouts(w http.ResponseWriter, r *http.Request) {
	workouts, err := a.Store.GetWorkouts(r.Context())
	if err != nil {
		a.frontendServerError(w, r, err)
		return
	}

	a.renderFrontend(w, r, "workouts-page", "workouts-content", frontendData{
		Title:    "Workouts",
		Active:   "workouts",
		Workouts: workouts,
	})
}

func (a *Application) FrontendWorkoutByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		a.frontendError(w, r, http.StatusBadRequest, "invalid workout id")
		return
	}

	workout, err := a.Store.GetWorkoutByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		a.frontendError(w, r, http.StatusNotFound, "workout not found")
		return
	}
	if err != nil {
		a.frontendServerError(w, r, err)
		return
	}

	a.renderFrontend(w, r, "workout-page", "workout-content", frontendData{
		Title:   workout.Name,
		Active:  "workouts",
		Workout: workout,
	})
}

func (a *Application) FrontendExercises(w http.ResponseWriter, r *http.Request) {
	a.renderFrontendExercises(w, r, "exercises-page", "exercises-content")
}

func (a *Application) FrontendExerciseResults(w http.ResponseWriter, r *http.Request) {
	a.renderFrontendExercises(w, r, "exercise-results", "exercise-results")
}

func (a *Application) FrontendExerciseByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		a.frontendError(w, r, http.StatusBadRequest, "invalid exercise id")
		return
	}

	exercise, err := a.Store.GetExerciseByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		a.frontendError(w, r, http.StatusNotFound, "exercise not found")
		return
	}
	if err != nil {
		a.frontendServerError(w, r, err)
		return
	}

	a.renderFrontend(w, r, "exercise-page", "exercise-content", frontendData{
		Title:    exercise.Name,
		Active:   "exercises",
		Exercise: exercise,
	})
}

func (a *Application) renderFrontendExercises(w http.ResponseWriter, r *http.Request, pageTemplate string, partialTemplate string) {
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	exercises, err := a.Store.GetExercises(r.Context(), search, 24)
	if err != nil {
		a.frontendServerError(w, r, err)
		return
	}

	a.renderFrontend(w, r, pageTemplate, partialTemplate, frontendData{
		Title:     "Exercises",
		Active:    "exercises",
		Search:    search,
		Exercises: exercises,
	})
}

func (a *Application) renderFrontend(w http.ResponseWriter, r *http.Request, pageTemplate string, partialTemplate string, data frontendData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	templateName := pageTemplate
	if r.Header.Get("HX-Request") == "true" {
		templateName = partialTemplate
	}

	if err := frontendTemplates.ExecuteTemplate(w, templateName, data); err != nil {
		a.Logger.Printf("render frontend template %s: %v", templateName, err)
	}
}

func (a *Application) frontendServerError(w http.ResponseWriter, r *http.Request, err error) {
	a.Logger.Printf("server error: %v", err)
	a.frontendError(w, r, http.StatusInternalServerError, "something tore a gasket")
}

func (a *Application) frontendError(w http.ResponseWriter, r *http.Request, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	a.renderFrontend(w, r, "error-page", "error-content", frontendData{
		Title:  "Forge Fitness",
		Active: "home",
		Error:  message,
	})
}
