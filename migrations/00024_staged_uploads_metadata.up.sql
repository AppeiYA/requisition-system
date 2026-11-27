ALTER TABLE staged_uploads
    ADD COLUMN IF NOT EXISTS claimed_at TIMESTAMPTZ;

UPDATE staged_uploads
SET claimed_at = created_at
WHERE status = 'claimed' AND claimed_at IS NULL;
