-- Phase 2: split requisitions into headers and lines, add requested items,
-- and introduce append-only status history.

ALTER TABLE requisitions
    ADD COLUMN reference_code VARCHAR(30);

CREATE TABLE requested_items (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category_id VARCHAR(36) NOT NULL
        REFERENCES item_categories(id) ON DELETE RESTRICT,
    description TEXT NOT NULL,
    created_by VARCHAR(36) NOT NULL
        REFERENCES users(id) ON DELETE RESTRICT,
    review_status VARCHAR(30) NOT NULL DEFAULT 'unreviewed'
        CHECK (review_status IN ('unreviewed', 'promoted', 'dismissed')),
    promoted_item_id VARCHAR(36) NULL
        REFERENCES items(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT requested_items_promotion_consistency CHECK (
        (review_status = 'promoted' AND promoted_item_id IS NOT NULL)
        OR (review_status <> 'promoted' AND promoted_item_id IS NULL)
    )
);

CREATE INDEX idx_requested_items_category_id ON requested_items(category_id);
CREATE INDEX idx_requested_items_review_status ON requested_items(review_status);

CREATE TABLE requisition_items (
    id VARCHAR(36) PRIMARY KEY,
    requisition_id VARCHAR(36) NOT NULL
        REFERENCES requisitions(id) ON DELETE RESTRICT,
    line_number SMALLINT NOT NULL CHECK (line_number > 0),
    item_id VARCHAR(36) NULL
        REFERENCES items(id) ON DELETE RESTRICT,
    requested_item_id VARCHAR(36) NULL
        REFERENCES requested_items(id) ON DELETE RESTRICT,
    item_name_snapshot VARCHAR(100) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    estimated_unit_amount BIGINT NOT NULL
        CHECK (estimated_unit_amount >= 0),
    notes TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT requisition_items_one_source CHECK (
        (item_id IS NOT NULL) <> (requested_item_id IS NOT NULL)
    ),
    CONSTRAINT requisition_items_line_identity UNIQUE (id, requisition_id),
    CONSTRAINT requisition_items_line_number_unique UNIQUE (requisition_id, line_number)
);

CREATE UNIQUE INDEX ux_requisition_items_catalogue_item
    ON requisition_items (requisition_id, item_id)
    WHERE item_id IS NOT NULL;

CREATE UNIQUE INDEX ux_requisition_items_requested_item
    ON requisition_items (requested_item_id)
    WHERE requested_item_id IS NOT NULL;

CREATE INDEX idx_requisition_items_requisition_id
    ON requisition_items(requisition_id);

CREATE TABLE requisition_status_history (
    id VARCHAR(36) PRIMARY KEY,
    requisition_id VARCHAR(36) NOT NULL
        REFERENCES requisitions(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    from_status VARCHAR(30) NULL,
    to_status VARCHAR(30) NOT NULL
        CHECK (to_status IN (
            'pending', 'approved', 'rejected',
            'in_fulfillment', 'completed', 'cancelled'
        )),
    acted_by VARCHAR(36) NOT NULL
        REFERENCES users(id) ON DELETE RESTRICT,
    actor_role VARCHAR(20) NOT NULL
        CHECK (actor_role IN ('admin', 'staff')),
    reason TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT requisition_status_history_sequence_unique
        UNIQUE (requisition_id, sequence),
    CONSTRAINT requisition_status_history_status_change CHECK (
        from_status IS DISTINCT FROM to_status
    ),
    CONSTRAINT requisition_status_history_from_status_valid CHECK (
        from_status IS NULL OR from_status IN (
            'pending', 'approved', 'rejected',
            'in_fulfillment', 'completed', 'cancelled'
        )
    ),
    CONSTRAINT requisition_status_history_initial_sequence CHECK (
        (sequence = 1 AND from_status IS NULL)
        OR (sequence > 1 AND from_status IS NOT NULL)
    )
);

CREATE INDEX idx_requisition_status_history_requisition_id
    ON requisition_status_history(requisition_id, sequence);

-- Generate stable legacy identifiers before moving the old single-line data.
INSERT INTO requested_items (
    id, name, category_id, description, created_by, created_at
)
SELECT
    'ritem_' || id,
    requested_item_name,
    category_id,
    requested_item_description,
    requested_by,
    created_at
FROM requisitions
WHERE item_id IS NULL;

INSERT INTO requisition_items (
    id, requisition_id, line_number, item_id, requested_item_id,
    item_name_snapshot, quantity, estimated_unit_amount, created_at
)
SELECT
    'reqitem_' || r.id,
    r.id,
    1,
    r.item_id,
    CASE WHEN r.item_id IS NULL THEN 'ritem_' || r.id ELSE NULL END,
    r.requested_item_name,
    r.quantity,
    r.estimated_cost,
    r.created_at
FROM requisitions r;

ALTER TABLE requisitions
    DROP COLUMN item_id,
    DROP COLUMN category_id,
    DROP COLUMN requested_item_name,
    DROP COLUMN requested_item_description,
    DROP COLUMN quantity,
    DROP COLUMN estimated_cost;

UPDATE requisitions
SET priority = 'medium'
WHERE priority = 'normal';

ALTER TABLE requisitions
    DROP CONSTRAINT IF EXISTS requisitions_priority_check,
    ADD CONSTRAINT requisitions_priority_check
        CHECK (priority IN ('low', 'medium', 'high', 'urgent'));

WITH numbered_requisitions AS (
    SELECT
        id,
        EXTRACT(YEAR FROM created_at)::INTEGER AS created_year,
        10000 + row_number() OVER (ORDER BY created_at, id) AS sequence_number
    FROM requisitions
)
UPDATE requisitions r
SET reference_code = 'REQ-' || nr.created_year || '-' || nr.sequence_number::INTEGER
FROM numbered_requisitions nr
WHERE nr.id = r.id;

ALTER TABLE requisitions
    ALTER COLUMN reference_code SET NOT NULL;

ALTER TABLE requisitions
    ADD CONSTRAINT requisitions_reference_code_unique UNIQUE (reference_code);

-- Existing rows are represented by their current status as a legacy initial
-- snapshot. New submissions will always write NULL -> pending.
INSERT INTO requisition_status_history (
    id, requisition_id, sequence, from_status, to_status, acted_by, actor_role, created_at
)
SELECT
    'rsh_' || id,
    id,
    1,
    NULL,
    status,
    requested_by,
    'staff',
    created_at
FROM requisitions;

DROP INDEX IF EXISTS idx_requisitions_item_id;
DROP INDEX IF EXISTS idx_requisitions_category_id;

-- Catalogue names are globally unique regardless of case in the new model.
DROP INDEX IF EXISTS ux_items_category_lower_name;
CREATE UNIQUE INDEX ux_items_lower_name ON items (LOWER(name));

UPDATE items
SET updated_at = created_at
WHERE updated_at IS NULL;

ALTER TABLE items
    ALTER COLUMN description TYPE TEXT,
    ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP,
    ALTER COLUMN updated_at SET NOT NULL;

CREATE OR REPLACE FUNCTION reject_requisition_status_history_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'requisition status history is append-only';
END;
$$;

CREATE TRIGGER requisition_status_history_no_update
BEFORE UPDATE ON requisition_status_history
FOR EACH ROW EXECUTE FUNCTION reject_requisition_status_history_mutation();

CREATE TRIGGER requisition_status_history_no_delete
BEFORE DELETE ON requisition_status_history
FOR EACH ROW EXECUTE FUNCTION reject_requisition_status_history_mutation();

CREATE OR REPLACE FUNCTION validate_requisition_status_sync()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    requisition_key VARCHAR(36);
    latest_status VARCHAR(30);
BEGIN
    IF TG_TABLE_NAME = 'requisitions' THEN
        requisition_key := NEW.id;
    ELSE
        requisition_key := NEW.requisition_id;
    END IF;

    SELECT h.to_status
    INTO latest_status
    FROM requisition_status_history h
    WHERE h.requisition_id = requisition_key
    ORDER BY h.sequence DESC
    LIMIT 1;

    IF latest_status IS NULL THEN
        RAISE EXCEPTION 'requisition must have status history';
    END IF;

    IF latest_status <> (
        SELECT r.status
        FROM requisitions r
        WHERE r.id = requisition_key
    ) THEN
        RAISE EXCEPTION 'requisition status must match latest status history';
    END IF;

    RETURN NEW;
END;
$$;

CREATE CONSTRAINT TRIGGER requisition_status_sync_after_update
AFTER INSERT OR UPDATE OF status ON requisitions
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION validate_requisition_status_sync();

CREATE CONSTRAINT TRIGGER requisition_status_sync_after_history
AFTER INSERT ON requisition_status_history
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION validate_requisition_status_sync();
