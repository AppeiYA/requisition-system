CREATE TABLE IF NOT EXISTS item_categories (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(200) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_item_categories_status ON item_categories(status);

-- Category names are unique regardless of casing so the same category cannot
-- be seeded or created twice with different capitalization.
CREATE UNIQUE INDEX IF NOT EXISTS ux_item_categories_lower_name ON item_categories (LOWER(name));
