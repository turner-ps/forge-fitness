-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_user
  ADD COLUMN firebase_uid TEXT;

CREATE UNIQUE INDEX idx_app_user_firebase_uid ON app_user(firebase_uid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_app_user_firebase_uid;

ALTER TABLE app_user
  DROP COLUMN IF EXISTS firebase_uid;
-- +goose StatementEnd
