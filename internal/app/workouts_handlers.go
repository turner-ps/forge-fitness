package app

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/httpjson"
)

func (a *Application) GetWorkouts(w http.ResponseWriter, r *http.Request) {
	workouts, err := a.Store.GetWorkouts(r.Context())
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{"workouts": workouts})
	if err != nil {
		a.Logger.Printf("write workouts response: %v", err)
	}
}

func (a *Application) GetWorkoutByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		a.badRequest(w, "invalid workout id")
		return
	}

	workout, err := a.Store.GetWorkoutByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "workout not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{"workout": workout})
	if err != nil {
		a.Logger.Printf("write workout response: %v", err)
	}
}
