-- +goose Up 
-- +goose StatementBegin 
ALTER TABLE urls ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE; 
CREATE INDEX idx_is_deleted ON urls (is_deleted); 
-- +goose StatementEnd

-- +goose Down 
-- +goose StatementBegin 
DROP INDEX IF EXISTS idx_is_deleted ON urls; 
ALTER TABLE urls DROP COLUMN is_deleted; 
-- +goose StatementEnd