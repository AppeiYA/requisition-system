CREATE TABLE IF NOT EXISTS items (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(200) NOT NULL DEFAULT '',
    category_id VARCHAR(36) NOT NULL,
    unit_cost BIGINT NOT NULL DEFAULT 0 CHECK (unit_cost >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_items_category FOREIGN KEY (category_id)
        REFERENCES item_categories(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_items_category_id ON items(category_id);
CREATE INDEX IF NOT EXISTS idx_items_status ON items(status);

-- Item names must be unique within their category, regardless of casing.
CREATE UNIQUE INDEX IF NOT EXISTS ux_items_category_lower_name
    ON items (category_id, LOWER(name));
