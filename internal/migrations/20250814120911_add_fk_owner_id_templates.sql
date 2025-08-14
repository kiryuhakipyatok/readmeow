-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS templates
ADD CONSTRAINT fk_owner_id
FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS templates
DROP CONSTRAINT IF EXISTS fk_owner_id 
-- +goose StatementEnd
