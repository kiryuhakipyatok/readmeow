-- +goose Up
-- +goose StatementBegin
INSERT INTO templates (id, owner_id, title, image, description, render_order, is_public)
VALUES ('00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000', 'Default template', 'https://res.cloudinary.com/dt02alvlt/image/upload/v1761942803/templates/21247ea9-44cb-497b-b0b0-f4c5974568fd-1761942803.png', 'default template', '{}'::TEXT[], true)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM templates WHERE id = '00000000-0000-0000-0000-000000000000'
-- +goose StatementEnd
