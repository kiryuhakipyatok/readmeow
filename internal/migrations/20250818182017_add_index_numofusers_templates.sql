-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS templates_num_of_users_idx ON templates(num_of_users)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS templates_num_of_users_idx
-- +goose StatementEnd