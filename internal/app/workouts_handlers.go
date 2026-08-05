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
	"github.com/turner-ps/forge-fitness/internal/httpjson"
	"github.com/turner-ps/forge-fitness/internal/store"
)

type createWorkoutRequest struct {
	Name string `json:"name"`
}

type addWorkoutExerciseRequest struct {
	ExerciseID      int64    `json:"exercise_id"`
	Position        int      `json:"position"`
	Sets            *int     `json:"sets"`
	Reps            *int     `json:"reps"`
	Weight          *float64 `json:"weight"`
	DurationSeconds *int     `json:"duration_seconds"`
}

type addWorkoutExercisesRequest struct {
	Exercises []addWorkoutExerciseRequest `json:"exercises"`
}

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

func (a *Application) GetWorkouts(w http.ResponseWriter, r *http.Request) {
	userID, ok := a.userIDParam(w, r)
	if !ok {
		return
	}

	_, err := a.Store.GetUserByID(r.Context(), userID)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "user not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	workouts, err := a.Store.GetWorkoutsByUserID(r.Context(), userID)
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
	userID, ok := a.userIDParam(w, r)
	if !ok {
		return
	}

	_, err := a.Store.GetUserByID(r.Context(), userID)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "user not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
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
	userID, ok := a.userIDParam(w, r)
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

func (a *Application) GetWorkoutExercises(w http.ResponseWriter, r *http.Request) {
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

	workoutExercises, err := a.Store.GetWorkoutExercisesForUser(r.Context(), userID, workoutID)
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusOK, httpjson.Envelope{"workout_exercises": workoutExercises})
	if err != nil {
		a.Logger.Printf("write workout exercises response: %v", err)
	}
}

func (a *Application) AddExerciseToWorkout(w http.ResponseWriter, r *http.Request) {
	userID, workoutID, ok := a.userWorkoutParams(w, r)
	if !ok {
		return
	}

	var request addWorkoutExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		a.badRequest(w, "invalid workout exercise request")
		return
	}

	if !a.validWorkoutExerciseRequest(w, r, request) {
		return
	}

	workoutExercise, err := a.Store.AddExerciseToWorkout(r.Context(), store.AddExerciseToWorkoutInput{
		UserID:          userID,
		WorkoutID:       workoutID,
		ExerciseID:      request.ExerciseID,
		Position:        request.Position,
		Sets:            request.Sets,
		Reps:            request.Reps,
		Weight:          request.Weight,
		DurationSeconds: request.DurationSeconds,
	})
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "workout not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusCreated, httpjson.Envelope{"workout_exercise": workoutExercise})
	if err != nil {
		a.Logger.Printf("write workout exercise response: %v", err)
	}
}

func (a *Application) AddExercisesToWorkout(w http.ResponseWriter, r *http.Request) {
	userID, workoutID, ok := a.userWorkoutParams(w, r)
	if !ok {
		return
	}

	var request addWorkoutExercisesRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		a.badRequest(w, "invalid workout exercises request")
		return
	}

	if len(request.Exercises) == 0 {
		a.badRequest(w, "at least one exercise is required")
		return
	}

	inputs := make([]store.AddExerciseToWorkoutInput, 0, len(request.Exercises))
	for _, exercise := range request.Exercises {
		if !a.validWorkoutExerciseRequest(w, r, exercise) {
			return
		}

		inputs = append(inputs, store.AddExerciseToWorkoutInput{
			UserID:          userID,
			WorkoutID:       workoutID,
			ExerciseID:      exercise.ExerciseID,
			Position:        exercise.Position,
			Sets:            exercise.Sets,
			Reps:            exercise.Reps,
			Weight:          exercise.Weight,
			DurationSeconds: exercise.DurationSeconds,
		})
	}

	workoutExercises, err := a.Store.AddExercisesToWorkout(r.Context(), inputs)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "workout not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	err = httpjson.WriteJSON(w, http.StatusCreated, httpjson.Envelope{"workout_exercises": workoutExercises})
	if err != nil {
		a.Logger.Printf("write workout exercises response: %v", err)
	}
}

func (a *Application) GetUserWorkoutSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := a.userIDParam(w, r)
	if !ok {
		return
	}

	_, err := a.Store.GetUserByID(r.Context(), userID)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "user not found")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}

	sessions, err := a.Store.GetWorkoutSessionsByUserID(r.Context(), userID)
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

func (a *Application) userIDParam(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil || userID < 1 {
		a.badRequest(w, "invalid user id")
		return 0, false
	}

	return userID, true
}

func (a *Application) userWorkoutParams(w http.ResponseWriter, r *http.Request) (int64, int64, bool) {
	userID, ok := a.userIDParam(w, r)
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

func (a *Application) userSessionParams(w http.ResponseWriter, r *http.Request) (int64, int64, bool) {
	userID, ok := a.userIDParam(w, r)
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

func invalidPositiveInt(value *int) bool {
	return value != nil && *value < 1
}

func (a *Application) validWorkoutExerciseRequest(w http.ResponseWriter, r *http.Request, request addWorkoutExerciseRequest) bool {
	if request.ExerciseID < 1 {
		a.badRequest(w, "exercise id is required")
		return false
	}
	if request.Position < 0 {
		a.badRequest(w, "position cannot be negative")
		return false
	}
	if invalidPositiveInt(request.Sets) || invalidPositiveInt(request.Reps) || invalidPositiveInt(request.DurationSeconds) {
		a.badRequest(w, "sets, reps, and duration_seconds must be positive when provided")
		return false
	}
	if request.Weight != nil && *request.Weight < 0 {
		a.badRequest(w, "weight cannot be negative")
		return false
	}

	_, err := a.Store.GetExerciseByID(r.Context(), request.ExerciseID)
	if errors.Is(err, sql.ErrNoRows) {
		a.notFound(w, "exercise not found")
		return false
	}
	if err != nil {
		a.serverError(w, err)
		return false
	}

	return true
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
