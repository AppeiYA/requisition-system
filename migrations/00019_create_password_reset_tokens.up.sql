CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_password_reset_tokens_hash
    ON password_reset_tokens (token_hash);

-- At most one redeemable token per user. Expired rows are invalidated by the
-- request flow before a new row is inserted; cleanup removes them later.
CREATE UNIQUE INDEX IF NOT EXISTS ux_password_reset_tokens_active_user
    ON password_reset_tokens (user_id)
    WHERE used_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_created
    ON password_reset_tokens (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_created
    ON password_reset_tokens (created_at);
