-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS widgets_likes_idx ON widgets(likes)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS widgets_likes_idx
-- +goose StatementEnd
