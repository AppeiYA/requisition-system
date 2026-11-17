DROP INDEX IF EXISTS idx_requested_items_reviewed_by;

ALTER TABLE requested_items
    DROP CONSTRAINT IF EXISTS requested_items_review_metadata_consistency,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS reviewed_by;
