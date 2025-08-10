-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS widgets(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(80) NOT NULL UNIQUE,
    image TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    link TEXT NOT NULL UNIQUE,
    likes INTEGER NOT NULL DEFAULT 0,
    num_of_users INTEGER NOT NULL DEFAULT 0
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS widgets
-- +goose StatementEnd
