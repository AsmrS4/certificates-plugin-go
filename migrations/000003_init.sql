-- +goose Up
ALTER TABLE certificate_attachments 
    ADD COLUMN IF NOT EXISTS file_name TEXT,
    ADD COLUMN IF NOT EXISTS file_type TEXT,
    ADD COLUMN IF NOT EXISTS mime_type TEXT;

-- +goose Down
ALTER TABLE certificate_attachments 
    DROP COLUMN IF EXISTS file_name,
    DROP COLUMN IF EXISTS file_type,
    DROP COLUMN IF EXISTS mime_type;