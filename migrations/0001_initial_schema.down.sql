-- 0001_initial_schema.down.sql
-- Down migration: drop initial schema for inventory reservation system

DROP TRIGGER IF EXISTS trg_update_reservations_updated_at ON reservations;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_reservations_quote_id;
DROP INDEX IF EXISTS idx_reservations_order_id;
DROP INDEX IF EXISTS idx_reservations_status;
DROP INDEX IF EXISTS idx_reservation_items_sku;

DROP TABLE IF EXISTS reservation_items;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS stock;