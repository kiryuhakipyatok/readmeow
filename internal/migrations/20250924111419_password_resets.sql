-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS password_resets(
    user_id UUID PRIMARY KEY,
    token BYTEA UNIQUE NOT NULL,
    expired_time TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) 
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS password_resets
-- +goose StatementEnd
