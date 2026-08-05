-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_user (
  id BIGSERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO app_user (email, name)
VALUES ('default@forge-fitness.local', 'Default User')
ON CONFLICT (email) DO NOTHING;

ALTER TABLE workout
  ADD COLUMN user_id BIGINT REFERENCES app_user(id) ON DELETE CASCADE;

UPDATE workout
SET user_id = (
  SELECT id
  FROM app_user
  WHERE email = 'default@forge-fitness.local'
)
WHERE user_id IS NULL;

ALTER TABLE workout
  ALTER COLUMN user_id SET NOT NULL;

CREATE INDEX idx_workout_user_id ON workout(user_id);

CREATE TABLE IF NOT EXISTS workout_exercise (
  id BIGSERIAL PRIMARY KEY,
  workout_id BIGINT NOT NULL REFERENCES workout(id) ON DELETE CASCADE,
  exercise_id BIGINT NOT NULL REFERENCES exercise(id),
  position INTEGER NOT NULL DEFAULT 0,
  sets INTEGER CHECK (sets IS NULL OR sets > 0),
  reps INTEGER CHECK (reps IS NULL OR reps > 0),
  weight NUMERIC(10, 2) CHECK (weight IS NULL OR weight >= 0),
  duration_seconds INTEGER CHECK (duration_seconds IS NULL OR duration_seconds > 0),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_workout_exercise_workout_id ON workout_exercise(workout_id);
CREATE INDEX idx_workout_exercise_exercise_id ON workout_exercise(exercise_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workout_exercise;
DROP INDEX IF EXISTS idx_workout_user_id;
ALTER TABLE workout DROP CONSTRAINT IF EXISTS workout_user_id_fkey;
ALTER TABLE workout DROP COLUMN IF EXISTS user_id;
DROP TABLE IF EXISTS app_user;
-- +goose StatementEnd
