-- +goose Up
-- +goose StatementBegin
INSERT INTO widgets (id, title, image, description, type, link) 
VALUES ('00000000-0000-0000-0000-000000000000','Default widget', 'https://res.cloudinary.com/dt02alvlt/image/upload/v1761942803/templates/21247ea9-44cb-497b-b0b0-f4c5974568fd-1761942803.png', 'default widget', 'default', 'https://www.google.com/')
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM widgets WHERE id = '00000000-0000-0000-0000-000000000000'
-- +goose StatementEnd
