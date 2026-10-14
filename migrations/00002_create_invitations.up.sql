CREATE TABLE IF NOT EXISTS invitations (
    id UUID PRIMARY KEY,
    email VARCHAR(254) NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    invited_by VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations(email);
CREATE INDEX IF NOT EXISTS idx_invitations_token_hash ON invitations(token_hash);
CREATE INDEX IF NOT EXISTS idx_invitations_invited_by ON invitations(invited_by);
CREATE INDEX IF NOT EXISTS idx_invitations_expires_accepted_revoked ON invitations(expires_at, accepted_at, revoked_at);

-- Prevent duplicate active (pending) invitations for same normalized email.
-- Cannot use NOW() in partial index predicate; enforce time-based active check in application.
-- This partial index prevents multiple pending non-revoked non-accepted invitations for same email.
CREATE UNIQUE INDEX IF NOT EXISTS ux_invitations_pending_email
    ON invitations (LOWER(email))
    WHERE accepted_at IS NULL AND revoked_at IS NULL;
