package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/turner-ps/forge-fitness/internal/store"
	"github.com/turner-ps/forge-fitness/migrations"
)

const insertExercise = `
INSERT INTO exercise (
  name,
  level,
  force,
  mechanic,
  equipment,
  primary_muscle_group,
  secondary_muscle_group,
  instructions,
  category,
  images
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  ARRAY(SELECT jsonb_array_elements_text($6::jsonb)),
  ARRAY(SELECT jsonb_array_elements_text($7::jsonb)),
  ARRAY(SELECT jsonb_array_elements_text($8::jsonb)),
  $9,
  ARRAY(SELECT jsonb_array_elements_text($10::jsonb))
)`

type exerciseData struct {
	SourceID             string   `json:"id"`
	Name                 string   `json:"name"`
	Force                *string  `json:"force"`
	Level                string   `json:"level"`
	Mechanic             *string  `json:"mechanic"`
	Equipment            *string  `json:"equipment"`
	PrimaryMuscleGroup   []string `json:"primaryMuscleGroup"`
	SecondaryMuscleGroup []string `json:"secondaryMuscleGroup"`
	Instructions         []string `json:"instructions"`
	Category             string   `json:"category"`
	Images               []string `json:"images"`
}

var errAlreadyImported = errors.New("exercise table already contains rows")

func main() {
	var dataPath string
	flag.StringVar(&dataPath, "file", "data/exercise_data/exercises.nd.json", "path to exercise ndjson file")
	flag.Parse()

	ctx := context.Background()

	db, err := store.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close db: %v", err)
		}
	}()

	if err := store.MigrateFS(db, migrations.FS, "."); err != nil {
		log.Fatal(err)
	}

	count, err := importExercises(ctx, db, dataPath)
	if errors.Is(err, errAlreadyImported) {
		log.Print(err)
		return
	}
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("imported %d exercises", count)
}

func importExercises(ctx context.Context, db *sql.DB, dataPath string) (int, error) {
	exercises, err := readExercises(dataPath)
	if err != nil {
		return 0, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, "LOCK TABLE exercise IN EXCLUSIVE MODE"); err != nil {
		return 0, err
	}

	var existing int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM exercise").Scan(&existing); err != nil {
		return 0, err
	}
	if existing > 0 {
		return 0, fmt.Errorf("%w: found %d rows; skipping import", errAlreadyImported, existing)
	}

	stmt, err := tx.PrepareContext(ctx, insertExercise)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = stmt.Close()
	}()

	for _, exercise := range exercises {
		primaryMuscleGroup, err := jsonStringArray(exercise.PrimaryMuscleGroup)
		if err != nil {
			return 0, fmt.Errorf("marshal primary muscle group %q: %w", exercise.SourceID, err)
		}

		secondaryMuscleGroup, err := jsonStringArray(exercise.SecondaryMuscleGroup)
		if err != nil {
			return 0, fmt.Errorf("marshal secondary muscle group %q: %w", exercise.SourceID, err)
		}

		instructions, err := jsonStringArray(exercise.Instructions)
		if err != nil {
			return 0, fmt.Errorf("marshal instructions %q: %w", exercise.SourceID, err)
		}

		images, err := jsonStringArray(exercise.Images)
		if err != nil {
			return 0, fmt.Errorf("marshal images %q: %w", exercise.SourceID, err)
		}

		_, err = stmt.ExecContext(ctx,
			exercise.Name,
			exercise.Level,
			stringValue(exercise.Force),
			stringValue(exercise.Mechanic),
			stringValue(exercise.Equipment),
			primaryMuscleGroup,
			secondaryMuscleGroup,
			instructions,
			exercise.Category,
			images,
		)
		if err != nil {
			return 0, fmt.Errorf("insert exercise %q: %w", exercise.SourceID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return len(exercises), nil
}

func jsonStringArray(values []string) (string, error) {
	if values == nil {
		values = []string{}
	}

	data, err := json.Marshal(values)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func readExercises(dataPath string) ([]exerciseData, error) {
	file, err := os.Open(dataPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	decoder := json.NewDecoder(file)
	seen := make(map[string]struct{})
	var exercises []exerciseData

	for {
		var exercise exerciseData
		if err := decoder.Decode(&exercise); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}

		if exercise.SourceID == "" {
			return nil, fmt.Errorf("exercise at import index %d has empty source id", len(exercises))
		}
		if _, ok := seen[exercise.SourceID]; ok {
			return nil, fmt.Errorf("duplicate source id %q in import file", exercise.SourceID)
		}
		seen[exercise.SourceID] = struct{}{}

		if exercise.Name == "" {
			return nil, fmt.Errorf("exercise %q has empty name", exercise.SourceID)
		}
		if exercise.Level == "" {
			return nil, fmt.Errorf("exercise %q has empty level", exercise.SourceID)
		}
		if exercise.Category == "" {
			return nil, fmt.Errorf("exercise %q has empty category", exercise.SourceID)
		}

		exercises = append(exercises, exercise)
	}

	return exercises, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
