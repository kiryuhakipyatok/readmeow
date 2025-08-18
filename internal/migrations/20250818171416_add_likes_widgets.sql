-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS widgets
ADD COLUMN IF NOT EXISTS likes INTEGER NOT NULL CHECK(likes>=0) DEFAULT 0
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS widgets
DROP COLUMN IF EXISTS likes
-- +goose StatementEnd
