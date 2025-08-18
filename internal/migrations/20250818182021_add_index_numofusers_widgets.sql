-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS widgets_num_of_users_idx ON widgets(num_of_users)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS widgets_num_of_users_idx
-- +goose StatementEnd