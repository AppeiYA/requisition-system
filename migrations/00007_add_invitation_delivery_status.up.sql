-- Delivery outcome of the invitation email, recorded by the worker process:
-- 'pending' until the mail provider accepts the mail, 'delivered' once it
-- does, and 'failed' after the task queue has exhausted its retries.
ALTER TABLE invitations
    ADD COLUMN IF NOT EXISTS status VARCHAR(16) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS delivery_failed_at TIMESTAMPTZ NULL;

-- Backfill the lifecycle states that are already knowable from the timestamps.
UPDATE invitations SET status = 'accepted' WHERE accepted_at IS NOT NULL;
UPDATE invitations SET status = 'revoked'  WHERE accepted_at IS NULL AND revoked_at IS NOT NULL;
UPDATE invitations SET status = 'expired'  WHERE accepted_at IS NULL AND revoked_at IS NULL AND expires_at <= NOW();

-- Rebuild the pending-email uniqueness on the status column. Time-based
-- expiry cannot appear in an index predicate (NOW() is not immutable), so the
-- old index kept expired — and now also permanently undeliverable —
-- invitations blocking a re-invite forever. The application also checks
-- expiry on top of this index, as the comment on the original index requires.
DROP INDEX IF EXISTS ux_invitations_pending_email;
CREATE UNIQUE INDEX IF NOT EXISTS ux_invitations_pending_email
    ON invitations (LOWER(email))
    WHERE status IN ('pending', 'delivered');

CREATE INDEX IF NOT EXISTS idx_invitations_status ON invitations(status);
