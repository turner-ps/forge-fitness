-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workout_session (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  workout_id BIGINT REFERENCES workout(id) ON DELETE SET NULL,
  performed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  notes TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_workout_session_user_id_performed_at ON workout_session(user_id, performed_at DESC);
CREATE INDEX idx_workout_session_workout_id ON workout_session(workout_id);

CREATE TABLE IF NOT EXISTS workout_session_exercise (
  id BIGSERIAL PRIMARY KEY,
  workout_session_id BIGINT NOT NULL REFERENCES workout_session(id) ON DELETE CASCADE,
  exercise_id BIGINT NOT NULL REFERENCES exercise(id),
  position INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_workout_session_exercise_session_id_position ON workout_session_exercise(workout_session_id, position, id);
CREATE INDEX idx_workout_session_exercise_exercise_id ON workout_session_exercise(exercise_id);

CREATE TABLE IF NOT EXISTS workout_session_set (
  id BIGSERIAL PRIMARY KEY,
  workout_session_exercise_id BIGINT NOT NULL REFERENCES workout_session_exercise(id) ON DELETE CASCADE,
  set_number INTEGER NOT NULL CHECK (set_number > 0),
  reps INTEGER CHECK (reps IS NULL OR reps > 0),
  weight NUMERIC(10, 2) CHECK (weight IS NULL OR weight >= 0),
  duration_seconds INTEGER CHECK (duration_seconds IS NULL OR duration_seconds > 0),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_workout_session_set_exercise_id_set_number ON workout_session_set(workout_session_exercise_id, set_number, id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workout_session_set;
DROP TABLE IF EXISTS workout_session_exercise;
DROP TABLE IF EXISTS workout_session;
-- +goose StatementEnd
