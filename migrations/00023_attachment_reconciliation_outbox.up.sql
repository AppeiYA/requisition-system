CREATE TABLE IF NOT EXISTS attachment_reconciliation_outbox (
    id                VARCHAR(36) PRIMARY KEY CHECK (id LIKE 'aevt-%'),
    attachment_id     VARCHAR(36) NOT NULL UNIQUE
        REFERENCES attachments(id) ON DELETE CASCADE,
    staged_upload_id  VARCHAR(36) NOT NULL UNIQUE
        REFERENCES staged_uploads(id) ON DELETE RESTRICT,
    from_key          TEXT NOT NULL,
    to_key            TEXT NOT NULL,
    checksum          CHAR(64) NOT NULL,
    status            VARCHAR(16) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'done', 'failed', 'dead')),
    attempts          INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    next_attempt_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_attachment_outbox_due
    ON attachment_reconciliation_outbox (next_attempt_at)
    WHERE status IN ('pending', 'failed');
