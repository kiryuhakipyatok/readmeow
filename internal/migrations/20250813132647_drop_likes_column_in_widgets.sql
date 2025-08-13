-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS widgets
DROP COLUMN likes
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS widgets
ADD COLUMN likes INTEGER NOT NULL DEFAULT 0
-- +goose StatementEnd
