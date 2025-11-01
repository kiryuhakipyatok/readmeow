-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, nickname, email, avatar) 
VALUES ('00000000-0000-0000-0000-000000000000', 'readmeow', 'readmeow@gmail.com', 'https://res.cloudinary.com/dt02alvlt/image/upload/v1755885015/avatars/2a1fba6f-0a4f-4b01-b970-c4bddddd0307-1755885014.jpg')
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000000'
-- +goose StatementEnd
