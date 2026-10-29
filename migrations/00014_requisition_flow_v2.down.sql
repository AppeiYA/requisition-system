DROP TRIGGER IF EXISTS requisition_status_sync_after_history ON requisition_status_history;
DROP TRIGGER IF EXISTS requisition_status_sync_after_update ON requisitions;
DROP FUNCTION IF EXISTS validate_requisition_status_sync();

DROP TRIGGER IF EXISTS requisition_status_history_no_update ON requisition_status_history;
DROP TRIGGER IF EXISTS requisition_status_history_no_delete ON requisition_status_history;
DROP FUNCTION IF EXISTS reject_requisition_status_history_mutation();

DROP INDEX IF EXISTS ux_items_lower_name;
CREATE UNIQUE INDEX IF NOT EXISTS ux_items_category_lower_name
    ON items (category_id, LOWER(name));

ALTER TABLE items
    ALTER COLUMN description TYPE VARCHAR(200),
    ALTER COLUMN updated_at DROP NOT NULL;

ALTER TABLE requisitions
    ADD COLUMN item_id VARCHAR(36) NULL REFERENCES items(id) ON DELETE RESTRICT,
    ADD COLUMN category_id VARCHAR(36),
    ADD COLUMN requested_item_name VARCHAR(100),
    ADD COLUMN requested_item_description VARCHAR(200),
    ADD COLUMN quantity INTEGER,
    ADD COLUMN estimated_cost BIGINT;

UPDATE requisitions r
SET
    item_id = ri.item_id,
    category_id = COALESCE(i.category_id, req.category_id),
    requested_item_name = ri.item_name_snapshot,
    requested_item_description = COALESCE(req.description, ''),
    quantity = ri.quantity,
    estimated_cost = ri.estimated_unit_amount
FROM requisition_items ri
LEFT JOIN items i ON i.id = ri.item_id
LEFT JOIN requested_items req ON req.id = ri.requested_item_id
WHERE ri.requisition_id = r.id
  AND ri.line_number = 1;

ALTER TABLE requisitions
    ALTER COLUMN category_id SET NOT NULL,
    ALTER COLUMN requested_item_name SET NOT NULL,
    ALTER COLUMN requested_item_description SET NOT NULL,
    ALTER COLUMN quantity SET NOT NULL,
    ALTER COLUMN estimated_cost SET NOT NULL;

DROP TABLE requisition_status_history;
DROP TABLE requisition_items;
DROP TABLE requested_items;

ALTER TABLE requisitions
    DROP CONSTRAINT requisitions_reference_code_unique,
    DROP CONSTRAINT requisitions_priority_check,
    DROP COLUMN reference_code;

ALTER TABLE requisitions
    ADD CONSTRAINT requisitions_priority_check
        CHECK (priority IN ('low', 'normal', 'medium', 'urgent'));
