-- +goose Up
-- +goose StatementBegin
INSERT INTO templates (id, owner_id, title, image, description, render_order, is_public)
VALUES ('00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000', 'Default template', 'https://res.cloudinary.com/dt02alvlt/image/upload/v1755885015/avatars/2a1fba6f-0a4f-4b01-b970-c4bddddd0307-1755885014.jpg', 'default template', '{}'::TEXT[], true)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM templates WHERE id = '00000000-0000-0000-0000-000000000000'
-- +goose StatementEnd
