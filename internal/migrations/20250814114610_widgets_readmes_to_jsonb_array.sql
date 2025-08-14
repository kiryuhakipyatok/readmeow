-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS readmes
DROP COLUMN widgets,
ADD COLUMN widgets JSONB[] NOT NULL

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS readmes
DROP COLUMN widgets,
ADD COLUMN widgets TEXT[] NOT NULL
-- +goose StatementEnd
