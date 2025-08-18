-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS templates_likes_idx ON templates(likes)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS templates_likes_idx
-- +goose StatementEnd
