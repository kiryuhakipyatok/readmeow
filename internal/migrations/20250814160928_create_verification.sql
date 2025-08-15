-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS verifications(
    email VARCHAR(256) PRIMARY KEY NOT NULL,
    nickname VARCHAR(80) NOT NULL UNIQUE,
    login VARCHAR(80) NOT NULL UNIQUE,
    password BYTEA NOT NULL,
    code BYTEA NOT NULL UNIQUE,
    attempts NUMERIC NOT NULL CHECK (attempts >= 0) DEFAULT 5,
    expired_time TIMESTAMP NOT NULL 
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS verifications
-- +goose StatementEnd
