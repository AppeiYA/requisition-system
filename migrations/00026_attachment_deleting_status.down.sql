ALTER TABLE attachments
    DROP CONSTRAINT IF EXISTS attachments_status_check;

ALTER TABLE attachments
    ADD CONSTRAINT attachments_status_check
    CHECK (status IN ('pending', 'processing', 'ready', 'failed'));
