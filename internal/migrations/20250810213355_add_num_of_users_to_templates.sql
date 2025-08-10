-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS templates
ADD COLUMN num_of_users INTEGER NOT NULL DEFAULT 0

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS templates
DROP COLUMN num_of_users
-- +goose StatementEnd
