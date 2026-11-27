DROP INDEX IF EXISTS idx_attachments_staged_upload_unique;
ALTER TABLE attachments
    DROP COLUMN IF EXISTS staged_upload_id,
    DROP COLUMN IF EXISTS status;
