ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role VARCHAR(32) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user'));

CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

