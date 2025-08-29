-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS users
ALTER COLUMN password DROP NOT NULL;

ALTER TABLE IF EXISTS users
ALTER COLUMN login DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS users
ALTER COLUMN password SET NOT NULL;

ALTER TABLE IF EXISTS users
ALTER COLUMN login SET NOT NULL;
-- +goose StatementEnd
