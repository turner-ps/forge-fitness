package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Workout struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateWorkoutInput struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
}

type UpdateWorkoutInput struct {
	UserID int64  `json:"user_id"`
	ID     int64  `json:"id"`
	Name   string `json:"name"`
}

type WorkoutExercise struct {
	ID              int64     `json:"id"`
	WorkoutID       int64     `json:"workout_id"`
	ExerciseID      int64     `json:"exercise_id"`
	ExerciseName    string    `json:"exercise_name"`
	Position        int       `json:"position"`
	Sets            *int      `json:"sets"`
	Reps            *int      `json:"reps"`
	Weight          *float64  `json:"weight"`
	DurationSeconds *int      `json:"duration_seconds"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AddExerciseToWorkoutInput struct {
	UserID          int64    `json:"user_id"`
	WorkoutID       int64    `json:"workout_id"`
	ExerciseID      int64    `json:"exercise_id"`
	Position        int      `json:"position"`
	Sets            *int     `json:"sets"`
	Reps            *int     `json:"reps"`
	Weight          *float64 `json:"weight"`
	DurationSeconds *int     `json:"duration_seconds"`
}

type workoutScanner interface {
	Scan(dest ...any) error
}

type workoutExecer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *Store) GetWorkouts(ctx context.Context) ([]Workout, error) {
	const query = `
SELECT id, user_id, name, created_at, updated_at
FROM workout
ORDER BY id`

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get workouts: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var workouts []Workout
	for rows.Next() {
		workout, err := scanWorkout(rows)
		if err != nil {
			return nil, fmt.Errorf("scan workout: %w", err)
		}

		workouts = append(workouts, *workout)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workouts: %w", err)
	}

	return workouts, nil
}

func (s *Store) GetWorkoutsByUserID(ctx context.Context, userID int64) ([]Workout, error) {
	const query = `
SELECT id, user_id, name, created_at, updated_at
FROM workout
WHERE user_id = $1
ORDER BY id`

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get workouts by user id %d: %w", userID, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var workouts []Workout
	for rows.Next() {
		workout, err := scanWorkout(rows)
		if err != nil {
			return nil, fmt.Errorf("scan workout: %w", err)
		}

		workouts = append(workouts, *workout)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workouts by user id %d: %w", userID, err)
	}

	return workouts, nil
}

func (s *Store) GetWorkoutByID(ctx context.Context, id int64) (*Workout, error) {
	const query = `
SELECT id, user_id, name, created_at, updated_at
FROM workout
WHERE id = $1`

	workout, err := scanWorkout(s.DB.QueryRowContext(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("get workout by id %d: %w", id, err)
	}

	return workout, nil
}

func (s *Store) GetWorkoutByIDForUser(ctx context.Context, userID int64, id int64) (*Workout, error) {
	const query = `
SELECT id, user_id, name, created_at, updated_at
FROM workout
WHERE user_id = $1 AND id = $2`

	workout, err := scanWorkout(s.DB.QueryRowContext(ctx, query, userID, id))
	if err != nil {
		return nil, fmt.Errorf("get workout by id %d for user id %d: %w", id, userID, err)
	}

	return workout, nil
}

func (s *Store) CreateWorkout(ctx context.Context, input CreateWorkoutInput) (*Workout, error) {
	const query = `
INSERT INTO workout (user_id, name)
VALUES ($1, $2)
RETURNING id, user_id, name, created_at, updated_at`

	workout, err := scanWorkout(s.DB.QueryRowContext(ctx, query, input.UserID, input.Name))
	if err != nil {
		return nil, fmt.Errorf("create workout for user id %d: %w", input.UserID, err)
	}

	return workout, nil
}

func (s *Store) UpdateWorkout(ctx context.Context, input UpdateWorkoutInput) (*Workout, error) {
	const query = `
UPDATE workout
SET name = $3, updated_at = NOW()
WHERE user_id = $1 AND id = $2
RETURNING id, user_id, name, created_at, updated_at`

	workout, err := scanWorkout(s.DB.QueryRowContext(ctx, query, input.UserID, input.ID, input.Name))
	if err != nil {
		return nil, fmt.Errorf("update workout id %d for user id %d: %w", input.ID, input.UserID, err)
	}

	return workout, nil
}

func (s *Store) DeleteWorkout(ctx context.Context, userID int64, workoutID int64) error {
	const query = `
DELETE FROM workout
WHERE user_id = $1 AND id = $2`

	result, err := s.DB.ExecContext(ctx, query, userID, workoutID)
	if err != nil {
		return fmt.Errorf("delete workout id %d for user id %d: %w", workoutID, userID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete workout rows affected id %d for user id %d: %w", workoutID, userID, err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Store) AddExerciseToWorkout(ctx context.Context, input AddExerciseToWorkoutInput) (*WorkoutExercise, error) {
	workoutExercise, err := addExerciseToWorkout(ctx, s.DB, input)
	if err != nil {
		return nil, fmt.Errorf("add exercise id %d to workout id %d for user id %d: %w", input.ExerciseID, input.WorkoutID, input.UserID, err)
	}

	return workoutExercise, nil
}

func (s *Store) AddExercisesToWorkout(ctx context.Context, inputs []AddExerciseToWorkoutInput) ([]WorkoutExercise, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin add exercises to workout: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	workoutExercises := make([]WorkoutExercise, 0, len(inputs))
	for _, input := range inputs {
		workoutExercise, err := addExerciseToWorkout(ctx, tx, input)
		if err != nil {
			return nil, fmt.Errorf("add exercise id %d to workout id %d for user id %d: %w", input.ExerciseID, input.WorkoutID, input.UserID, err)
		}

		workoutExercises = append(workoutExercises, *workoutExercise)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit add exercises to workout: %w", err)
	}

	return workoutExercises, nil
}

func addExerciseToWorkout(ctx context.Context, db workoutExecer, input AddExerciseToWorkoutInput) (*WorkoutExercise, error) {
	const query = `
INSERT INTO workout_exercise (
  workout_id,
  exercise_id,
  position,
  sets,
  reps,
  weight,
  duration_seconds
)
SELECT
  w.id,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8
FROM workout w
WHERE w.id = $1 AND w.user_id = $2
RETURNING
  id,
  workout_id,
  exercise_id,
  (SELECT e.name FROM exercise e WHERE e.id = workout_exercise.exercise_id),
  position,
  sets,
  reps,
  weight,
  duration_seconds,
  created_at,
  updated_at`

	return scanWorkoutExercise(db.QueryRowContext(
		ctx,
		query,
		input.WorkoutID,
		input.UserID,
		input.ExerciseID,
		input.Position,
		nullableInt(input.Sets),
		nullableInt(input.Reps),
		nullableFloat(input.Weight),
		nullableInt(input.DurationSeconds),
	))
}

func (s *Store) GetWorkoutExercises(ctx context.Context, workoutID int64) ([]WorkoutExercise, error) {
	const query = `
SELECT
  we.id,
  we.workout_id,
  we.exercise_id,
  e.name,
  we.position,
  we.sets,
  we.reps,
  we.weight,
  we.duration_seconds,
  we.created_at,
  we.updated_at
FROM workout_exercise we
INNER JOIN exercise e ON e.id = we.exercise_id
WHERE we.workout_id = $1
ORDER BY we.position, we.id`

	rows, err := s.DB.QueryContext(ctx, query, workoutID)
	if err != nil {
		return nil, fmt.Errorf("get workout exercises for workout id %d: %w", workoutID, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var workoutExercises []WorkoutExercise
	for rows.Next() {
		workoutExercise, err := scanWorkoutExercise(rows)
		if err != nil {
			return nil, fmt.Errorf("scan workout exercise: %w", err)
		}

		workoutExercises = append(workoutExercises, *workoutExercise)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workout exercises for workout id %d: %w", workoutID, err)
	}

	return workoutExercises, nil
}

func (s *Store) GetWorkoutExercisesForUser(ctx context.Context, userID int64, workoutID int64) ([]WorkoutExercise, error) {
	const query = `
SELECT
  we.id,
  we.workout_id,
  we.exercise_id,
  e.name,
  we.position,
  we.sets,
  we.reps,
  we.weight,
  we.duration_seconds,
  we.created_at,
  we.updated_at
FROM workout_exercise we
INNER JOIN workout w ON w.id = we.workout_id
INNER JOIN exercise e ON e.id = we.exercise_id
WHERE w.user_id = $1 AND we.workout_id = $2
ORDER BY we.position, we.id`

	rows, err := s.DB.QueryContext(ctx, query, userID, workoutID)
	if err != nil {
		return nil, fmt.Errorf("get workout exercises for workout id %d and user id %d: %w", workoutID, userID, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var workoutExercises []WorkoutExercise
	for rows.Next() {
		workoutExercise, err := scanWorkoutExercise(rows)
		if err != nil {
			return nil, fmt.Errorf("scan workout exercise: %w", err)
		}

		workoutExercises = append(workoutExercises, *workoutExercise)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workout exercises for workout id %d and user id %d: %w", workoutID, userID, err)
	}

	return workoutExercises, nil
}

func scanWorkout(row workoutScanner) (*Workout, error) {
	var workout Workout
	err := row.Scan(
		&workout.ID,
		&workout.UserID,
		&workout.Name,
		&workout.CreatedAt,
		&workout.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &workout, nil
}

func scanWorkoutExercise(row workoutScanner) (*WorkoutExercise, error) {
	var workoutExercise WorkoutExercise
	var sets sql.NullInt64
	var reps sql.NullInt64
	var weight sql.NullFloat64
	var durationSeconds sql.NullInt64

	err := row.Scan(
		&workoutExercise.ID,
		&workoutExercise.WorkoutID,
		&workoutExercise.ExerciseID,
		&workoutExercise.ExerciseName,
		&workoutExercise.Position,
		&sets,
		&reps,
		&weight,
		&durationSeconds,
		&workoutExercise.CreatedAt,
		&workoutExercise.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	workoutExercise.Sets = intFromNullInt64(sets)
	workoutExercise.Reps = intFromNullInt64(reps)
	workoutExercise.Weight = floatFromNullFloat64(weight)
	workoutExercise.DurationSeconds = intFromNullInt64(durationSeconds)

	return &workoutExercise, nil
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}

	return *value
}

func nullableFloat(value *float64) any {
	if value == nil {
		return nil
	}

	return *value
}

func intFromNullInt64(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}

	converted := int(value.Int64)
	return &converted
}

func floatFromNullFloat64(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}

	return &value.Float64
}
