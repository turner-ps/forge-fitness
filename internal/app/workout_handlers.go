package app

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/store"
	"github.com/turner-ps/forge-fitness/utils"
)

func (a *Application) GetWorkouts(w http.ResponseWriter, r *http.Request) {
	workouts, err := store.GetWorkouts(r.Context(), a.DB)
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"workouts": workouts})
	if err != nil {
		a.Logger.Printf("write workouts response: %v", err)
	}
}

func (a *Application) GetWorkout(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		a.badRequest(w, "invalid workout id")
		return
	}

	workout, err := store.GetWorkoutByID(r.Context(), a.DB, id)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "workout not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"workout": workout})
	if err != nil {
		a.Logger.Printf("write workout response: %v", err)
	}
}
