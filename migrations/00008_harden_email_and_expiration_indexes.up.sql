-- Email addresses are normalized by the application before persistence. The
-- expression indexes keep older rows and direct SQL callers safe as well.
UPDATE users SET email = LOWER(email) WHERE email <> LOWER(email);
UPDATE invitations SET email = LOWER(email) WHERE email <> LOWER(email);

CREATE UNIQUE INDEX IF NOT EXISTS ux_users_lower_email ON users (LOWER(email));
CREATE INDEX IF NOT EXISTS idx_invitations_active_email_created
    ON invitations (LOWER(email), created_at DESC)
    WHERE status IN ('pending', 'delivered');
CREATE INDEX IF NOT EXISTS idx_invitations_active_expiration
    ON invitations (status, expires_at)
    WHERE status IN ('pending', 'delivered');
