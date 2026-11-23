-- Admin operations foundations: indexes for the admin list endpoints and
-- analytics, plus an append-only audit log for administrative mutations.
-- Additive only.

-- User list, filters and analytics.
CREATE INDEX IF NOT EXISTS idx_users_created_at
    ON users (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_role_status
    ON users (role, status);
CREATE INDEX IF NOT EXISTS idx_users_status_created_at
    ON users (status, created_at DESC);

-- Invitation list, filters and analytics.
CREATE INDEX IF NOT EXISTS idx_invitations_created_at
    ON invitations (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_invitations_status_created_at
    ON invitations (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_invitations_department_id
    ON invitations (department_id);

-- Requisition analytics.
CREATE INDEX IF NOT EXISTS idx_requisitions_created_at
    ON requisitions (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requisitions_status_created_at
    ON requisitions (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requisitions_department_created_at
    ON requisitions (department_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requisitions_priority
    ON requisitions (priority);
CREATE INDEX IF NOT EXISTS idx_requisition_items_item_id
    ON requisition_items (item_id)
    WHERE item_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_requested_items_created_at
    ON requested_items (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requisition_status_history_to_status_created_at
    ON requisition_status_history (to_status, created_at);

-- Audit trail. Written by the application in the same transaction as the
-- mutation, so the actor, reason and before/after state are captured together.
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id            VARCHAR(36) PRIMARY KEY,
    actor_id      VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action        VARCHAR(64) NOT NULL,
    target_type   VARCHAR(32) NOT NULL,
    target_id     VARCHAR(36) NOT NULL,
    before_state  JSONB NULL,
    after_state   JSONB NULL,
    reason        TEXT NULL,
    request_id    VARCHAR(64) NULL,
    ip_address    INET NULL,
    user_agent    TEXT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_audit_log_created_at
    ON admin_audit_log (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_target
    ON admin_audit_log (target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_actor
    ON admin_audit_log (actor_id);

-- The audit log is append-only: even a bug or a careless operator must not be
-- able to rewrite history. Corrections are new entries, not edits.
CREATE OR REPLACE FUNCTION reject_admin_audit_log_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'admin audit log is append-only';
END;
$$;

CREATE TRIGGER admin_audit_log_no_update
BEFORE UPDATE ON admin_audit_log
FOR EACH ROW EXECUTE FUNCTION reject_admin_audit_log_mutation();

CREATE TRIGGER admin_audit_log_no_delete
BEFORE DELETE ON admin_audit_log
FOR EACH ROW EXECUTE FUNCTION reject_admin_audit_log_mutation();
