-- Reverting this migration drops the admin audit log and its history. That is
-- acceptable only for development rollbacks: in any environment where audit
-- records matter, export them first.

DROP TRIGGER IF EXISTS admin_audit_log_no_update ON admin_audit_log;
DROP TRIGGER IF EXISTS admin_audit_log_no_delete ON admin_audit_log;
DROP FUNCTION IF EXISTS reject_admin_audit_log_mutation();
DROP TABLE IF EXISTS admin_audit_log;

DROP INDEX IF EXISTS idx_requisition_status_history_to_status_created_at;
DROP INDEX IF EXISTS idx_requested_items_created_at;
DROP INDEX IF EXISTS idx_requisition_items_item_id;
DROP INDEX IF EXISTS idx_requisitions_priority;
DROP INDEX IF EXISTS idx_requisitions_department_created_at;
DROP INDEX IF EXISTS idx_requisitions_status_created_at;
DROP INDEX IF EXISTS idx_requisitions_created_at;
DROP INDEX IF EXISTS idx_invitations_department_id;
DROP INDEX IF EXISTS idx_invitations_status_created_at;
DROP INDEX IF EXISTS idx_invitations_created_at;
DROP INDEX IF EXISTS idx_users_status_created_at;
DROP INDEX IF EXISTS idx_users_role_status;
DROP INDEX IF EXISTS idx_users_created_at;
