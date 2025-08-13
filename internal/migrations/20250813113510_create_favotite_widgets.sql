-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS widgets_likes(
    widget_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (widget_id, user_id),
    FOREIGN KEY (widget_id) REFERENCES widgets(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS widgets_likes
-- +goose StatementEnd
