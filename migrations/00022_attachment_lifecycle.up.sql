ALTER TABLE attachments
    ADD COLUMN IF NOT EXISTS status VARCHAR(16) NOT NULL DEFAULT 'ready'
        CHECK (status IN ('pending', 'processing', 'ready', 'failed')),
    ADD COLUMN IF NOT EXISTS staged_upload_id VARCHAR(36)
        REFERENCES staged_uploads(id) ON DELETE RESTRICT;

-- A staged upload may produce at most one permanent attachment. The unique
-- index is also the concurrent-claim guard: a second requisition creation
-- referencing the same token fails and rolls back.
CREATE UNIQUE INDEX IF NOT EXISTS idx_attachments_staged_upload_unique
    ON attachments (staged_upload_id)
    WHERE staged_upload_id IS NOT NULL;
