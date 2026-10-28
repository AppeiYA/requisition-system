ALTER TABLE requisitions
    ALTER COLUMN needed_by TYPE DATE
    USING needed_by::DATE;
