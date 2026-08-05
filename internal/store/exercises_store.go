// Package store
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Exercise struct {
	ID                   int       `json:"id"`
	Name                 string    `json:"name"`
	Level                string    `json:"level"`
	Force                string    `json:"force"`
	Mechanic             string    `json:"mechanic"`
	Equipment            string    `json:"equipment"`
	PrimaryMuscleGroup   []string  `json:"primary_muscle_group"`
	SecondaryMuscleGroup []string  `json:"secondary_muscle_group"`
	Instructions         []string  `json:"Instructions"`
	Category             string    `json:"category"`
	Images               []string  `json:"images"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type exerciseScanner interface {
	Scan(dest ...any) error
}

func (s *Store) GetExerciseByID(ctx context.Context, id int64) (*Exercise, error) {
	const query = `
SELECT
  id,
  name,
  COALESCE(level, ''),
  COALESCE(force, ''),
  COALESCE(mechanic, ''),
  COALESCE(equipment, ''),
  COALESCE(primary_muscle_group, ARRAY[]::text[]),
  COALESCE(secondary_muscle_group, ARRAY[]::text[]),
  COALESCE(instructions, ARRAY[]::text[]),
  COALESCE(category, ''),
  COALESCE(images, ARRAY[]::text[]),
  created_at,
  updated_at
FROM exercise
WHERE id = $1`

	exercise, err := scanExercise(s.DB.QueryRowContext(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("get exercise by id %d: %w", id, err)
	}

	return exercise, nil
}

func (s *Store) GetExercises(ctx context.Context, search string, limit int) ([]Exercise, error) {
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	const query = `
SELECT
  id,
  name,
  COALESCE(level, ''),
  COALESCE(force, ''),
  COALESCE(mechanic, ''),
  COALESCE(equipment, ''),
  COALESCE(primary_muscle_group, ARRAY[]::text[]),
  COALESCE(secondary_muscle_group, ARRAY[]::text[]),
  COALESCE(instructions, ARRAY[]::text[]),
  COALESCE(category, ''),
  COALESCE(images, ARRAY[]::text[]),
  created_at,
  updated_at
FROM exercise
WHERE
  $1 = ''
  OR name ILIKE '%' || $1 || '%'
  OR category ILIKE '%' || $1 || '%'
  OR equipment ILIKE '%' || $1 || '%'
  OR EXISTS (
    SELECT 1
    FROM unnest(primary_muscle_group) AS muscle
    WHERE muscle ILIKE '%' || $1 || '%'
  )
ORDER BY name
LIMIT $2`

	rows, err := s.DB.QueryContext(ctx, query, search, limit)
	if err != nil {
		return nil, fmt.Errorf("get exercises: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var exercises []Exercise
	for rows.Next() {
		exercise, err := scanExercise(rows)
		if err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}
		exercises = append(exercises, *exercise)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exercises: %w", err)
	}

	return exercises, nil
}

func scanExercise(row exerciseScanner) (*Exercise, error) {
	pgTypes := pgtype.NewMap()
	var exercise Exercise
	err := row.Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.Level,
		&exercise.Force,
		&exercise.Mechanic,
		&exercise.Equipment,
		pgTypes.SQLScanner(&exercise.PrimaryMuscleGroup),
		pgTypes.SQLScanner(&exercise.SecondaryMuscleGroup),
		pgTypes.SQLScanner(&exercise.Instructions),
		&exercise.Category,
		pgTypes.SQLScanner(&exercise.Images),
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &exercise, nil
}
