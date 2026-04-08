-- 001_users.sql: Users table with auth fields
-- Requirements: 28.1, 28.2
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email VARCHAR(255),
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMPTZ,
    account_locked_until TIMESTAMPTZ,
    last_login TIMESTAMPTZ,
    disclaimer_acknowledged BOOLEAN DEFAULT FALSE,
    theme_preference VARCHAR(10) DEFAULT 'light',
    language_preference VARCHAR(10) DEFAULT 'vi-VN',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
ALTER TABLE users
ADD COLUMN IF NOT EXISTS last_login TIMESTAMPTZ;
ALTER TABLE users
ADD COLUMN IF NOT EXISTS account_locked_until TIMESTAMPTZ;
ALTER TABLE users
ADD COLUMN IF NOT EXISTS theme_preference VARCHAR(10) DEFAULT 'light';
ALTER TABLE users
ADD COLUMN IF NOT EXISTS language_preference VARCHAR(10) DEFAULT 'vi-VN';
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);