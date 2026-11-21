ALTER TABLE requested_items
    ADD COLUMN reviewed_by VARCHAR(36)
        REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN reviewed_at TIMESTAMPTZ NULL;

ALTER TABLE requested_items
    ADD CONSTRAINT requested_items_review_metadata_consistency CHECK (
        (review_status = 'unreviewed' AND reviewed_by IS NULL AND reviewed_at IS NULL)
        OR
        (review_status IN ('promoted', 'dismissed')
            AND reviewed_by IS NOT NULL AND reviewed_at IS NOT NULL)
    );

CREATE INDEX idx_requested_items_reviewed_by
    ON requested_items(reviewed_by);
