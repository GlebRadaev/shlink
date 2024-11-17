-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN user_id VARCHAR(255) NOT NULL;
CREATE INDEX idx_user_id ON urls (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_id ON urls;
ALTER TABLE urls DROP COLUMN user_id;
-- +goose StatementEnd
