-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS widgets
ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '{}'
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS widgets
DROP COLUMN IF EXISTS tags 
-- +goose StatementEnd
