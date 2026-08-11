package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/auth"
	"github.com/turner-ps/forge-fitness/internal/httpjson"
	"github.com/turner-ps/forge-fitness/internal/store"
)

type createWorkoutRequest struct {
	Name string `json:"name"`
}

func (a *Application) GetWorkouts(w http.ResponseWriter, r *http.Request) {
	user, err := auth.RequireUser(r.Context())
	if err != nil {
		a.unauthorized(w)
		return
	}

	workouts, err := a.Store.GetWorkoutsByUserID(r.Context(), user.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{"workouts": workouts})
	if err != nil {
		a.Logger.Printf("write workouts response: %v", err)
	}
}

func (a *Application) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	userID, ok := a.authenticatedUserID(w, r)
	if !ok {
		return
	}

	var request createWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		a.badRequest(w, "invalid workout request")
		return
	}

	name := strings.TrimSpace(request.Name)
	if name == "" {
		a.badRequest(w, "workout name is required")
		return
	}

	workout, err := a.Store.CreateWorkout(r.Context(), store.CreateWorkoutInput{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusCreated, httpjson.Envelope{"workout": workout})
	if err != nil {
		a.Logger.Printf("write workout response: %v", err)
	}
}

func (a *Application) GetWorkoutByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := a.authenticatedUserID(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		a.badRequest(w, "invalid workout id")
		return
	}

	workout, err := a.Store.GetWorkoutByIDForUser(r.Context(), userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "workout not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	workoutExercises, err := a.Store.GetWorkoutExercisesForUser(r.Context(), userID, id)
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{
		"workout":           workout,
		"workout_exercises": workoutExercises,
	})
	if err != nil {
		a.Logger.Printf("write workout response: %v", err)
	}
}

func (a *Application) authenticatedUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	user, err := auth.RequireUser(r.Context())
	if err != nil {
		a.unauthorized(w)
		return 0, false
	}

	return user.ID, true
}

func (a *Application) userWorkoutParams(w http.ResponseWriter, r *http.Request) (int64, int64, bool) {
	userID, ok := a.authenticatedUserID(w, r)
	if !ok {
		return 0, 0, false
	}

	workoutID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || workoutID < 1 {
		a.badRequest(w, "invalid workout id")
		return 0, 0, false
	}

	return userID, workoutID, true
}
