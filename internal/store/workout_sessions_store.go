package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type WorkoutSession struct {
	ID          int64             `json:"id"`
	UserID      int64             `json:"user_id"`
	WorkoutID   *int64            `json:"workout_id"`
	PerformedAt time.Time         `json:"performed_at"`
	Notes       *string           `json:"notes"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Exercises   []SessionExercise `json:"exercises,omitempty"`
}

type SessionExercise struct {
	ID               int64        `json:"id"`
	WorkoutSessionID int64        `json:"workout_session_id"`
	ExerciseID       int64        `json:"exercise_id"`
	ExerciseName     string       `json:"exercise_name"`
	Position         int          `json:"position"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
	Sets             []SessionSet `json:"sets,omitempty"`
}

type SessionSet struct {
	ID                       int64     `json:"id"`
	WorkoutSessionExerciseID int64     `json:"workout_session_exercise_id"`
	SetNumber                int       `json:"set_number"`
	Reps                     *int      `json:"reps"`
	Weight                   *float64  `json:"weight"`
	DurationSeconds          *int      `json:"duration_seconds"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

type CreateWorkoutSessionInput struct {
	UserID      int64
	WorkoutID   int64
	PerformedAt *time.Time
	Notes       *string
	Exercises   []CreateSessionExerciseInput
}

type CreateSessionExerciseInput struct {
	ExerciseID int64
	Position   int
	Sets       []CreateSessionSetInput
}

type CreateSessionSetInput struct {
	SetNumber       int
	Reps            *int
	Weight          *float64
	DurationSeconds *int
}

func (s *Store) CreateWorkoutSession(ctx context.Context, input CreateWorkoutSessionInput) (*WorkoutSession, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create workout session: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	session, err := createWorkoutSession(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	for _, exerciseInput := range input.Exercises {
		exercise, err := createSessionExercise(ctx, tx, session.ID, exerciseInput)
		if err != nil {
			return nil, err
		}

		for _, setInput := range exerciseInput.Sets {
			set, err := createSessionSet(ctx, tx, exercise.ID, setInput)
			if err != nil {
				return nil, err
			}

			exercise.Sets = append(exercise.Sets, *set)
		}

		session.Exercises = append(session.Exercises, *exercise)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create workout session: %w", err)
	}

	return session, nil
}

func (s *Store) GetWorkoutSessionsByUserID(ctx context.Context, userID int64) ([]WorkoutSession, error) {
	const query = `
SELECT id, user_id, workout_id, performed_at, notes, created_at, updated_at
FROM workout_session
WHERE user_id = $1
ORDER BY performed_at DESC, id DESC`

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get workout sessions by user id %d: %w", userID, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var sessions []WorkoutSession
	for rows.Next() {
		session, err := scanWorkoutSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan workout session: %w", err)
		}

		sessions = append(sessions, *session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workout sessions by user id %d: %w", userID, err)
	}

	return sessions, nil
}

func (s *Store) GetWorkoutSessionsByWorkoutIDForUser(ctx context.Context, userID int64, workoutID int64) ([]WorkoutSession, error) {
	const query = `
SELECT id, user_id, workout_id, performed_at, notes, created_at, updated_at
FROM workout_session
WHERE user_id = $1 AND workout_id = $2
ORDER BY performed_at DESC, id DESC`

	rows, err := s.DB.QueryContext(ctx, query, userID, workoutID)
	if err != nil {
		return nil, fmt.Errorf("get workout sessions by workout id %d for user id %d: %w", workoutID, userID, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var sessions []WorkoutSession
	for rows.Next() {
		session, err := scanWorkoutSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan workout session: %w", err)
		}

		sessions = append(sessions, *session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workout sessions by workout id %d for user id %d: %w", workoutID, userID, err)
	}

	return sessions, nil
}

func (s *Store) GetWorkoutSessionByIDForUser(ctx context.Context, userID int64, sessionID int64) (*WorkoutSession, error) {
	const query = `
SELECT id, user_id, workout_id, performed_at, notes, created_at, updated_at
FROM workout_session
WHERE user_id = $1 AND id = $2`

	session, err := scanWorkoutSession(s.DB.QueryRowContext(ctx, query, userID, sessionID))
	if err != nil {
		return nil, fmt.Errorf("get workout session by id %d for user id %d: %w", sessionID, userID, err)
	}

	exercises, err := s.GetSessionExercises(ctx, session.ID)
	if err != nil {
		return nil, err
	}
	session.Exercises = exercises

	return session, nil
}

func (s *Store) GetSessionExercises(ctx context.Context, sessionID int64) ([]SessionExercise, error) {
	const query = `
SELECT
  wse.id,
  wse.workout_session_id,
  wse.exercise_id,
  e.name,
  wse.position,
  wse.created_at,
  wse.updated_at
FROM workout_session_exercise wse
INNER JOIN exercise e ON e.id = wse.exercise_id
WHERE wse.workout_session_id = $1
ORDER BY wse.position, wse.id`

	rows, err := s.DB.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session exercises for session id %d: %w", sessionID, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var exercises []SessionExercise
	for rows.Next() {
		exercise, err := scanSessionExercise(rows)
		if err != nil {
			return nil, fmt.Errorf("scan session exercise: %w", err)
		}

		sets, err := s.GetSessionSets(ctx, exercise.ID)
		if err != nil {
			return nil, err
		}
		exercise.Sets = sets

		exercises = append(exercises, *exercise)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate session exercises for session id %d: %w", sessionID, err)
	}

	return exercises, nil
}

func (s *Store) GetSessionSets(ctx context.Context, sessionExerciseID int64) ([]SessionSet, error) {
	const query = `
SELECT id, workout_session_exercise_id, set_number, reps, weight, duration_seconds, created_at, updated_at
FROM workout_session_set
WHERE workout_session_exercise_id = $1
ORDER BY set_number, id`

	rows, err := s.DB.QueryContext(ctx, query, sessionExerciseID)
	if err != nil {
		return nil, fmt.Errorf("get session sets for session exercise id %d: %w", sessionExerciseID, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var sets []SessionSet
	for rows.Next() {
		set, err := scanSessionSet(rows)
		if err != nil {
			return nil, fmt.Errorf("scan session set: %w", err)
		}

		sets = append(sets, *set)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate session sets for session exercise id %d: %w", sessionExerciseID, err)
	}

	return sets, nil
}

func createWorkoutSession(ctx context.Context, db workoutExecer, input CreateWorkoutSessionInput) (*WorkoutSession, error) {
	const query = `
INSERT INTO workout_session (user_id, workout_id, performed_at, notes)
SELECT
  w.user_id,
  w.id,
  COALESCE($3::timestamptz, CURRENT_TIMESTAMP),
  $4
FROM workout w
WHERE w.user_id = $1 AND w.id = $2
RETURNING id, user_id, workout_id, performed_at, notes, created_at, updated_at`

	session, err := scanWorkoutSession(db.QueryRowContext(ctx, query, input.UserID, input.WorkoutID, input.PerformedAt, nullableString(input.Notes)))
	if err != nil {
		return nil, fmt.Errorf("create workout session for workout id %d and user id %d: %w", input.WorkoutID, input.UserID, err)
	}

	return session, nil
}

func createSessionExercise(ctx context.Context, db workoutExecer, sessionID int64, input CreateSessionExerciseInput) (*SessionExercise, error) {
	const query = `
INSERT INTO workout_session_exercise (workout_session_id, exercise_id, position)
VALUES ($1, $2, $3)
RETURNING
  id,
  workout_session_id,
  exercise_id,
  (SELECT e.name FROM exercise e WHERE e.id = workout_session_exercise.exercise_id),
  position,
  created_at,
  updated_at`

	exercise, err := scanSessionExercise(db.QueryRowContext(ctx, query, sessionID, input.ExerciseID, input.Position))
	if err != nil {
		return nil, fmt.Errorf("create session exercise id %d for session id %d: %w", input.ExerciseID, sessionID, err)
	}

	return exercise, nil
}

func createSessionSet(ctx context.Context, db workoutExecer, sessionExerciseID int64, input CreateSessionSetInput) (*SessionSet, error) {
	const query = `
INSERT INTO workout_session_set (workout_session_exercise_id, set_number, reps, weight, duration_seconds)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, workout_session_exercise_id, set_number, reps, weight, duration_seconds, created_at, updated_at`

	set, err := scanSessionSet(db.QueryRowContext(
		ctx,
		query,
		sessionExerciseID,
		input.SetNumber,
		nullableInt(input.Reps),
		nullableFloat(input.Weight),
		nullableInt(input.DurationSeconds),
	))
	if err != nil {
		return nil, fmt.Errorf("create session set %d for session exercise id %d: %w", input.SetNumber, sessionExerciseID, err)
	}

	return set, nil
}

func scanWorkoutSession(row workoutScanner) (*WorkoutSession, error) {
	var session WorkoutSession
	var workoutID sql.NullInt64
	var notes sql.NullString

	err := row.Scan(
		&session.ID,
		&session.UserID,
		&workoutID,
		&session.PerformedAt,
		&notes,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	session.WorkoutID = int64FromNullInt64(workoutID)
	session.Notes = stringFromNullString(notes)

	return &session, nil
}

func scanSessionExercise(row workoutScanner) (*SessionExercise, error) {
	var exercise SessionExercise
	err := row.Scan(
		&exercise.ID,
		&exercise.WorkoutSessionID,
		&exercise.ExerciseID,
		&exercise.ExerciseName,
		&exercise.Position,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &exercise, nil
}

func scanSessionSet(row workoutScanner) (*SessionSet, error) {
	var set SessionSet
	var reps sql.NullInt64
	var weight sql.NullFloat64
	var durationSeconds sql.NullInt64

	err := row.Scan(
		&set.ID,
		&set.WorkoutSessionExerciseID,
		&set.SetNumber,
		&reps,
		&weight,
		&durationSeconds,
		&set.CreatedAt,
		&set.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	set.Reps = intFromNullInt64(reps)
	set.Weight = floatFromNullFloat64(weight)
	set.DurationSeconds = intFromNullInt64(durationSeconds)

	return &set, nil
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}

func int64FromNullInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}

	return &value.Int64
}

func stringFromNullString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}
