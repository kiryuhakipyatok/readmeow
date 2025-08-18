-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(80) UNIQUE NOT NULL,
    email VARCHAR(256) UNIQUE NOT NULL,
    avatar TEXT NOT NULL,
    password BYTEA NOT NULL,
    time_of_register TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    num_of_templates INTEGER NOT NULL DEFAULT 0,
    num_of_readmes INTEGER NOT NULL DEFAULT 0   
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users
-- +goose StatementEnd
