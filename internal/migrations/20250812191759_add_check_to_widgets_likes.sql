-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS widgets
ADD CONSTRAINT likes CHECK (likes >= 0 )
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS widgets
DROP CONSTRAINT likes CHECK (likes >= 0 )
-- +goose StatementEnd
