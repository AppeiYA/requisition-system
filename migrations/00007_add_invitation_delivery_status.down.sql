DROP INDEX IF EXISTS idx_invitations_status;

DROP INDEX IF EXISTS ux_invitations_pending_email;
CREATE UNIQUE INDEX IF NOT EXISTS ux_invitations_pending_email
    ON invitations (LOWER(email))
    WHERE accepted_at IS NULL AND revoked_at IS NULL;

ALTER TABLE invitations
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS delivered_at,
    DROP COLUMN IF EXISTS delivery_failed_at;
