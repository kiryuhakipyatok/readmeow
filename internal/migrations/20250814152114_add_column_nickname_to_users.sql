-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS users
ADD COLUMN nickname VARCHAR(80) NOT NULL UNIQUE
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS users
DROP COLUMN IF EXISTS nickname
-- +goose StatementEnd
