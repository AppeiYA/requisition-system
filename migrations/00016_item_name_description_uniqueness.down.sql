DROP INDEX IF EXISTS ux_items_lower_name_description;

CREATE UNIQUE INDEX ux_items_lower_name
    ON items (LOWER(name));
