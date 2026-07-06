-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS exercise (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  level VARCHAR(255),
  force VARCHAR(255),
  mechanic VARCHAR(255),
  equipment VARCHAR(255),
  primary_muscle_group TEXT[],
  secondary_muscle_group TEXT[],
  instructions TEXT[],
  category VARCHAR(255),
  images TEXT[],
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS exercise;
-- +goose StatementEnd
