ALTER TABLE invitations
    ADD COLUMN IF NOT EXISTS status VARCHAR(16) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS delivery_failed_at TIMESTAMPTZ NULL;

-- Backfill the lifecycle states that are already knowable from the timestamps.
UPDATE invitations SET status = 'accepted' WHERE accepted_at IS NOT NULL;
UPDATE invitations SET status = 'revoked'  WHERE accepted_at IS NULL AND revoked_at IS NOT NULL;
UPDATE invitations SET status = 'expired'  WHERE accepted_at IS NULL AND revoked_at IS NULL AND expires_at <= NOW();


DROP INDEX IF EXISTS ux_invitations_pending_email;
CREATE UNIQUE INDEX IF NOT EXISTS ux_invitations_pending_email
    ON invitations (LOWER(email))
    WHERE status IN ('pending', 'delivered');

CREATE INDEX IF NOT EXISTS idx_invitations_status ON invitations(status);
