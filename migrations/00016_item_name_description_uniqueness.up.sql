DROP INDEX IF EXISTS ux_items_lower_name;

CREATE UNIQUE INDEX ux_items_lower_name_description
    ON items (LOWER(name), LOWER(description));
