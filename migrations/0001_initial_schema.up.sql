-- 0001_initial_schema.up.sql
-- Up migration: create initial schema for inventory reservation system

CREATE TABLE stock (
    sku TEXT PRIMARY KEY,
    on_hand INT NOT NULL CHECK (on_hand >= 0),
    reserved INT NOT NULL CHECK (reserved >= 0),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE reservations (
    reservation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status TEXT NOT NULL CHECK (status IN ('PENDING','COMMITTED','RELEASED','EXPIRED','RESERVED','REVERTED')),
    quote_id TEXT UNIQUE NOT NULL,
    order_id TEXT UNIQUE,
    expires_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    items_hash TEXT,
    version INT NOT NULL DEFAULT 1
);

CREATE TABLE reservation_items (
    reservation_id UUID NOT NULL REFERENCES reservations(reservation_id) ON DELETE CASCADE,
    sku TEXT NOT NULL,
    qty INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (reservation_id, sku)
);


CREATE UNIQUE INDEX idx_reservations_quote_id ON reservations (quote_id) WHERE quote_id IS NOT NULL;
CREATE UNIQUE INDEX idx_reservations_order_id ON reservations (order_id) WHERE order_id IS NOT NULL;

CREATE INDEX idx_reservations_status ON reservations (status);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_reservations_updated_at
    BEFORE UPDATE ON reservations
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();
