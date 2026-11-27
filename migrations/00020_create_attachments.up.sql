CREATE TABLE IF NOT EXISTS attachments (
    id           VARCHAR(36) PRIMARY KEY,
    ref_name     VARCHAR(32) NOT NULL
        CHECK (ref_name IN ('requisition', 'requisition_item')),
    ref_id       VARCHAR(36) NOT NULL,
    storage_key  TEXT NOT NULL,
    file_name    TEXT NOT NULL,
    content_type VARCHAR(127) NOT NULL,
    file_size    BIGINT NOT NULL CHECK (file_size > 0),
    checksum     CHAR(64) NOT NULL,
    uploaded_by  VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT attachments_storage_key_unique UNIQUE (storage_key)
);

CREATE INDEX IF NOT EXISTS idx_attachments_ref
    ON attachments (ref_name, ref_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_attachments_uploaded_by
    ON attachments (uploaded_by);
