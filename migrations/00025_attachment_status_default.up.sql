ALTER TABLE attachments
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;

UPDATE attachments
SET status = 'ready'
WHERE status IS NULL;

CREATE INDEX IF NOT EXISTS idx_attachments_status_updated_at
    ON attachments (status, updated_at);
