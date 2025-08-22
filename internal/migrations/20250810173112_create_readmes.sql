-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS readmes(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    template_id UUID NOT NULL,
    image TEXT NOT NULL UNIQUE,
    title VARCHAR(80) NOT NULL UNIQUE,
    text TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    links TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    widgets JSONB[] NOT NULL DEFAULT ARRAY[]::JSONB[],
    render_order TEXT[] NOT NULL,
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id),
    FOREIGN KEY (template_id) REFERENCES templates(id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS readmes
-- +goose StatementEnd
