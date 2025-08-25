-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS readmes
DROP CONSTRAINT readmes_template_id_fkey;

ALTER TABLE IF EXISTS readmes
ALTER COLUMN template_id SET DEFAULT '00000000-0000-0000-0000-000000000000';

ALTER TABLE IF EXISTS readmes
ADD CONSTRAINT readmes_template_id_fkey 
FOREIGN KEY (template_id) REFERENCES templates(id)
ON DELETE SET DEFAULT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS readmes
DROP CONSTRAINT readmes_template_id_fkey;

ALTER TABLE IF EXISTS readmes
ADD CONSTRAINT readmes_template_id_fkey 
FOREIGN KEY (template_id) REFERENCES templates(id);
-- +goose StatementEnd
