DROP INDEX IF EXISTS idx_attachments_status_updated_at;

ALTER TABLE attachments
    DROP COLUMN IF EXISTS updated_at;
