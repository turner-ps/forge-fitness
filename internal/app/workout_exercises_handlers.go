package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/turner-ps/forge-fitness/internal/httpjson"
	"github.com/turner-ps/forge-fitness/internal/store"
)

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

func (a *Application) AddExerciseToWorkout(w http.ResponseWriter, r *http.Request) {
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

func invalidPositiveInt(value *int) bool {
	return value != nil && *value < 1
}
