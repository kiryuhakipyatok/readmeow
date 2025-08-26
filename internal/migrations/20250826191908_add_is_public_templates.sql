-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS templates
ADD COLUMN is_public BOOLEAN NOT NULL DEFAULT TRUE
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS templates
DROP COLUMN IF EXISTS is_public
-- +goose StatementEnd
