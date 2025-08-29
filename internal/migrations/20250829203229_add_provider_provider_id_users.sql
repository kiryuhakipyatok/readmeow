-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS users
ADD COLUMN provider VARCHAR(80) NOT NULL DEFAULT 'local';

ALTER TABLE IF EXISTS users
ADD COLUMN provider_id TEXT UNIQUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS users
DROP COLUMN IF EXISTS provider;

ALTER TABLE IF EXISTS users
DROP COLUMN IF EXISTS provider_id;
-- +goose StatementEnd
