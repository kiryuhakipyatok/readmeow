-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS templates_likes(
    template_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (template_id, user_id),
    FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS templates_likes
-- +goose StatementEnd
