-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS templates(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    title VARCHAR(80) NOT NULL UNIQUE,
    image TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL UNIQUE,
    text TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    links TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    widgets JSONB[] NOT NULL DEFAULT ARRAY[]::JSONB[],
    likes INTEGER NOT NULL CHECK(likes>=0) DEFAULT 0,
    render_order TEXT[] NOT NULL,
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    num_of_users INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (owner_id) REFERENCES users(id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS templates
-- +goose StatementEnd
