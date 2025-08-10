-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS templates(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    title VARCHAR(80) NOT NULL UNIQUE,
    image TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL UNIQUE,
    text TEXT[] NOT NULL,
    links TEXT[] NOT NULL,
    widgets TEXT[] NOT NULL,
    likes NUMERIC NOT NULL DEFAULT 0,
    render_order TEXT[] NOT NULL,
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS templates
-- +goose StatementEnd
