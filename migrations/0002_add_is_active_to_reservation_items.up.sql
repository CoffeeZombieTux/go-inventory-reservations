ALTER TABLE reservation_items
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE reservation_items
SET is_active = TRUE
WHERE is_active IS DISTINCT FROM TRUE;
