-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS templates_create_time_idx ON templates(create_time)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS templates_create_time_idx
-- +goose StatementEnd
