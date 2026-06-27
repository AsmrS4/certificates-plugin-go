-- +goose Up
DROP TABLE IF EXISTS user_positions CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP TABLE IF EXISTS cert_user_positions CASCADE;
DROP TABLE IF EXISTS cert_users CASCADE;

CREATE TABLE IF NOT EXISTS cert_users (
    id BIGINT PRIMARY KEY,
    full_name TEXT,
    external_id TEXT,
    tsu_accounts_id TEXT,
    tsu_linked BOOLEAN DEFAULT FALSE,
    is_teacher BOOLEAN DEFAULT FALSE,
    is_student BOOLEAN DEFAULT FALSE,
    is_dean_office BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cert_user_positions (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES cert_users(id) ON DELETE CASCADE,
    position_type TEXT,
    status TEXT,
    nationality_type TEXT,
    funding_type TEXT,
    education_form TEXT,
    faculty_name TEXT,
    department_name TEXT,
    program_name TEXT,
    stream_name TEXT,
    group_code TEXT,
    group_name TEXT,
    UNIQUE(user_id, position_type)
);

CREATE INDEX idx_cert_users_tsu_accounts_id ON cert_users(tsu_accounts_id);
CREATE INDEX idx_cert_user_positions_user_id ON cert_user_positions(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_cert_users_tsu_accounts_id;
DROP INDEX IF EXISTS idx_cert_user_positions_user_id;
DROP TABLE IF EXISTS cert_user_positions;
DROP TABLE IF EXISTS cert_users;