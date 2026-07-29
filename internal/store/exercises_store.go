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

	pgTypes := pgtype.NewMap()
	var exercise Exercise
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
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
		return nil, fmt.Errorf("get exercise by id %d: %w", id, err)
	}

	return &exercise, nil
}
