-- +goose Up
ALTER TABLE certificate_documents 
    ADD COLUMN IF NOT EXISTS file_type TEXT,
    ADD COLUMN IF NOT EXISTS mime_type TEXT;
    
DROP TABLE IF EXISTS documents CASCADE;

-- +goose Down
ALTER TABLE certificate_documents 
    DROP COLUMN IF NOT EXISTS file_type TEXT,
    DROP COLUMN IF NOT EXISTS mime_type TEXT;

CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES certificate_applications(id) ON DELETE CASCADE,
    file_id TEXT NOT NULL,
    file_name TEXT NOT NULL,
    storage_url TEXT,
    uploaded_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_documents_order_id ON documents(order_id);