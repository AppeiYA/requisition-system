CREATE TABLE IF NOT EXISTS staged_uploads (
    id                VARCHAR(36) PRIMARY KEY CHECK (id LIKE 'upl-%'),
    uploaded_by       VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    storage_path      TEXT NOT NULL UNIQUE CHECK (storage_path LIKE 'staging/upl-%'),
    original_filename TEXT NOT NULL,
    content_type      VARCHAR(127) NOT NULL,
    file_size         BIGINT NOT NULL CHECK (file_size > 0),
    checksum          CHAR(64) NOT NULL,
    status            VARCHAR(16) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'claimed', 'expired', 'deleted')),
    expires_at        TIMESTAMPTZ NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_staged_uploads_owner_status
    ON staged_uploads (uploaded_by, status);

CREATE INDEX IF NOT EXISTS idx_staged_uploads_pending_expiry
    ON staged_uploads (status, expires_at)
    WHERE status = 'pending';
