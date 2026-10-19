DROP INDEX IF EXISTS ux_invitations_pending_email;
DROP INDEX IF EXISTS idx_invitations_expires_accepted_revoked;
DROP INDEX IF EXISTS idx_invitations_invited_by;
DROP INDEX IF EXISTS idx_invitations_token_hash;
DROP INDEX IF EXISTS idx_invitations_email;
DROP TABLE IF EXISTS invitations;
