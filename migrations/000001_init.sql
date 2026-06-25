-- +goose Up
DROP TABLE IF EXISTS certificate_applications CASCADE;
DROP TABLE IF EXISTS attachments CASCADE;
DROP TABLE IF EXISTS documents CASCADE;
DROP TABLE IF EXISTS certificate_documents CASCADE;

CREATE TABLE IF NOT EXISTS certificate_applications (
    id SERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL,
    application_status VARCHAR(20) NOT NULL DEFAULT 'Pending',
    certificate_type VARCHAR(50) NOT NULL,
    obtain_method VARCHAR(20) NOT NULL,
    comment TEXT,
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    form_data JSONB
);

CREATE TABLE IF NOT EXISTS certificate_attachments (
    id SERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES certificate_applications(id) ON DELETE CASCADE,
    file_id TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_type TEXT,
    mime_type TEXT,
    size BIGINT,
    uploaded_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES certificate_applications(id) ON DELETE CASCADE,
    file_id TEXT NOT NULL,
    file_name TEXT NOT NULL,
    storage_url TEXT,
    uploaded_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS certificate_documents (
    id SERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES certificate_applications(id) ON DELETE CASCADE,
    file_id TEXT NOT NULL,
    file_name TEXT NOT NULL,
    storage_url TEXT,
    uploaded_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_cert_applications_student_id ON certificate_applications(student_id);
CREATE INDEX idx_cert_applications_status ON certificate_applications(application_status);
CREATE INDEX idx_cert_applications_type ON certificate_applications(certificate_type);
CREATE INDEX idx_cert_applications_created_at ON certificate_applications(created_at);

CREATE INDEX idx_attachments_order_id ON certificate_attachments(order_id);
CREATE INDEX idx_attachments_file_id ON certificate_attachments(file_id);

CREATE INDEX idx_documents_order_id ON documents(order_id);
CREATE INDEX idx_certificate_documents_order_id ON certificate_documents(order_id);

-- +goose Down
DROP INDEX IF EXISTS idx_certificate_documents_order_id;
DROP INDEX IF EXISTS idx_documents_order_id;
DROP INDEX IF EXISTS idx_certificate_attachments_file_id;
DROP INDEX IF EXISTS idx_certificate_attachments_order_id;
DROP INDEX IF EXISTS idx_cert_applications_created_at;
DROP INDEX IF EXISTS idx_cert_applications_type;
DROP INDEX IF EXISTS idx_cert_applications_status;
DROP INDEX IF EXISTS idx_cert_applications_student_id;

DROP TABLE IF EXISTS certificate_documents;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS certificate_attachments;
DROP TABLE IF EXISTS certificate_applications;