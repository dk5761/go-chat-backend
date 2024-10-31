-- up migration: create users table with new schema
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login TIMESTAMP,
    last_login_token TIMESTAMP
);

-- Add index on email to optimize lookups by email
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
