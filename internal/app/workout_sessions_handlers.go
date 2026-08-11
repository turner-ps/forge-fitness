package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/turner-ps/forge-fitness/internal/auth"
	"github.com/turner-ps/forge-fitness/internal/httpjson"
	"github.com/turner-ps/forge-fitness/internal/store"
)

type createWorkoutSessionRequest struct {
	PerformedAt *time.Time                     `json:"performed_at"`
	Notes       *string                        `json:"notes"`
	Exercises   []createSessionExerciseRequest `json:"exercises"`
}

type createSessionExerciseRequest struct {
	ExerciseID int64                     `json:"exercise_id"`
	Position   int                       `json:"position"`
	Sets       []createSessionSetRequest `json:"sets"`
}

type createSessionSetRequest struct {
	SetNumber       int      `json:"set_number"`
	Reps            *int     `json:"reps"`
	Weight          *float64 `json:"weight"`
	DurationSeconds *int     `json:"duration_seconds"`
}

func (a *Application) GetUserWorkoutSessions(w http.ResponseWriter, r *http.Request) {
	user, err := auth.RequireUser(r.Context())
	if err != nil {
		a.unauthorized(w)
		return
	}

	sessions, err := a.Store.GetWorkoutSessionsByUserID(r.Context(), user.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{"workout_sessions": sessions})
	if err != nil {
		a.Logger.Printf("write workout sessions response: %v", err)
	}
}

func (a *Application) GetWorkoutSessions(w http.ResponseWriter, r *http.Request) {
	userID, workoutID, ok := a.userWorkoutParams(w, r)
	if !ok {
		return
	}

	_, err := a.Store.GetWorkoutByIDForUser(r.Context(), userID, workoutID)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "workout not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	sessions, err := a.Store.GetWorkoutSessionsByWorkoutIDForUser(r.Context(), userID, workoutID)
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{"workout_sessions": sessions})
	if err != nil {
		a.Logger.Printf("write workout sessions response: %v", err)
	}
}

func (a *Application) GetWorkoutSessionByID(w http.ResponseWriter, r *http.Request) {
	userID, sessionID, ok := a.userSessionParams(w, r)
	if !ok {
		return
	}

	session, err := a.Store.GetWorkoutSessionByIDForUser(r.Context(), userID, sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "workout session not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{"workout_session": session})
	if err != nil {
		a.Logger.Printf("write workout session response: %v", err)
	}
}

func (a *Application) CreateWorkoutSession(w http.ResponseWriter, r *http.Request) {
	userID, workoutID, ok := a.userWorkoutParams(w, r)
	if !ok {
		return
	}

	var request createWorkoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		a.badRequest(w, "invalid workout session request")
		return
	}

	if len(request.Exercises) == 0 {
		a.badRequest(w, "at least one session exercise is required")
		return
	}

	input := store.CreateWorkoutSessionInput{
		UserID:      userID,
		WorkoutID:   workoutID,
		PerformedAt: request.PerformedAt,
		Notes:       trimmedOptionalString(request.Notes),
		Exercises:   make([]store.CreateSessionExerciseInput, 0, len(request.Exercises)),
	}

	for _, exercise := range request.Exercises {
		exerciseInput, ok := a.validSessionExerciseRequest(w, r, exercise)
		if !ok {
			return
		}

		input.Exercises = append(input.Exercises, exerciseInput)
	}

	session, err := a.Store.CreateWorkoutSession(r.Context(), input)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "workout not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusCreated, httpjson.Envelope{"workout_session": session})
	if err != nil {
		a.Logger.Printf("write workout session response: %v", err)
	}
}

func (a *Application) userSessionParams(w http.ResponseWriter, r *http.Request) (int64, int64, bool) {
	userID, ok := a.authenticatedUserID(w, r)
	if !ok {
		return 0, 0, false
	}

	sessionID, err := strconv.ParseInt(chi.URLParam(r, "sessionID"), 10, 64)
	if err != nil || sessionID < 1 {
		a.badRequest(w, "invalid workout session id")
		return 0, 0, false
	}

	return userID, sessionID, true
}

func (a *Application) validSessionExerciseRequest(w http.ResponseWriter, r *http.Request, request createSessionExerciseRequest) (store.CreateSessionExerciseInput, bool) {
	if request.ExerciseID < 1 {
		a.badRequest(w, "exercise id is required")
		return store.CreateSessionExerciseInput{}, false
	}
	if request.Position < 0 {
		a.badRequest(w, "position cannot be negative")
		return store.CreateSessionExerciseInput{}, false
	}
	if len(request.Sets) == 0 {
		a.badRequest(w, "at least one set is required for each session exercise")
		return store.CreateSessionExerciseInput{}, false
	}

	_, err := a.Store.GetExerciseByID(r.Context(), request.ExerciseID)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "exercise not found")
		return store.CreateSessionExerciseInput{}, false
	}
	if err != nil {
		a.serverError(w, err)
		return store.CreateSessionExerciseInput{}, false
	}

	input := store.CreateSessionExerciseInput{
		ExerciseID: request.ExerciseID,
		Position:   request.Position,
		Sets:       make([]store.CreateSessionSetInput, 0, len(request.Sets)),
	}

	for _, set := range request.Sets {
		setInput, ok := a.validSessionSetRequest(w, set)
		if !ok {
			return store.CreateSessionExerciseInput{}, false
		}

		input.Sets = append(input.Sets, setInput)
	}

	return input, true
}

func (a *Application) validSessionSetRequest(w http.ResponseWriter, request createSessionSetRequest) (store.CreateSessionSetInput, bool) {
	if request.SetNumber < 1 {
		a.badRequest(w, "set_number must be positive")
		return store.CreateSessionSetInput{}, false
	}
	if invalidPositiveInt(request.Reps) || invalidPositiveInt(request.DurationSeconds) {
		a.badRequest(w, "reps and duration_seconds must be positive when provided")
		return store.CreateSessionSetInput{}, false
	}
	if request.Weight != nil && *request.Weight < 0 {
		a.badRequest(w, "weight cannot be negative")
		return store.CreateSessionSetInput{}, false
	}
	if request.Reps == nil && request.Weight == nil && request.DurationSeconds == nil {
		a.badRequest(w, "at least one of reps, weight, or duration_seconds is required for each set")
		return store.CreateSessionSetInput{}, false
	}

	return store.CreateSessionSetInput{
		SetNumber:       request.SetNumber,
		Reps:            request.Reps,
		Weight:          request.Weight,
		DurationSeconds: request.DurationSeconds,
	}, true
}

func trimmedOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
