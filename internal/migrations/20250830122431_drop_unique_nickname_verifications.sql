-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS verifications
DROP CONSTRAINT IF EXISTS verifications_nickname_key;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS verifications
ADD CONSTRAINT verifications_nickname_key UNIQUE (nickname);
-- +goose StatementEnd

