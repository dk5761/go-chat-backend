

ALTER TABLE users ADD COLUMN username VARCHAR(255) UNIQUE NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
