package store

import (
	"context"
	"fmt"
	"time"
)

type Workout struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Store) GetWorkouts(ctx context.Context) ([]Workout, error) {
	const query = `
SELECT id, name, created_at, updated_at
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
		var workout Workout
		if err := rows.Scan(
			&workout.ID,
			&workout.Name,
			&workout.CreatedAt,
			&workout.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan workout: %w", err)
		}

		workouts = append(workouts, workout)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workouts: %w", err)
	}

	return workouts, nil
}

func (s *Store) GetWorkoutByID(ctx context.Context, id int64) (*Workout, error) {
	const query = `
SELECT id, name, created_at, updated_at
FROM workout
WHERE id = $1`

	var workout Workout
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&workout.ID,
		&workout.Name,
		&workout.CreatedAt,
		&workout.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get workout by id %d: %w", id, err)
	}

	return &workout, nil
}
