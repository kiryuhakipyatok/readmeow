-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS users
DROP CONSTRAINT IF EXISTS users_nickname_key;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS users
ADD CONSTRAINT users_nickname_key UNIQUE (nickname);
-- +goose StatementEnd
