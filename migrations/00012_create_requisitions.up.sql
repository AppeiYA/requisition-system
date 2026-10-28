CREATE TABLE IF NOT EXISTS requisitions (
    id VARCHAR(36) PRIMARY KEY,
    requested_by VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    department_id VARCHAR(36) NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
    item_id VARCHAR(36) NULL REFERENCES items(id) ON DELETE RESTRICT,
    category_id VARCHAR(36) NOT NULL REFERENCES item_categories(id) ON DELETE RESTRICT,
    requested_item_name VARCHAR(100) NOT NULL,
    requested_item_description VARCHAR(200) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0 AND quantity <= 1000),
    estimated_cost BIGINT NOT NULL CHECK (estimated_cost >= 0),
    justification VARCHAR(500) NOT NULL,
    priority VARCHAR(20) NOT NULL CHECK (priority IN ('low', 'normal', 'medium', 'urgent')),
    needed_by DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected', 'completed', 'cancelled', 'in_fulfillment')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_requisitions_requested_by ON requisitions(requested_by);
CREATE INDEX IF NOT EXISTS idx_requisitions_department_id ON requisitions(department_id);
CREATE INDEX IF NOT EXISTS idx_requisitions_item_id ON requisitions(item_id);
CREATE INDEX IF NOT EXISTS idx_requisitions_category_id ON requisitions(category_id);
CREATE INDEX IF NOT EXISTS idx_requisitions_status ON requisitions(status);
