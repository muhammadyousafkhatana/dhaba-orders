-- Dhaba Orders: initial database schema

CREATE TABLE IF NOT EXISTS menu_items (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(120) NOT NULL UNIQUE,
    description TEXT,
    price       NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
    available   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id            BIGSERIAL PRIMARY KEY,
    customer_name VARCHAR(120) NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'NEW'
                  CHECK (status IN ('NEW', 'COOKING', 'READY', 'CANCELLED')),
    total_amount  NUMERIC(10, 2) NOT NULL CHECK (total_amount >= 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id BIGINT NOT NULL REFERENCES menu_items(id),
    quantity     INTEGER NOT NULL CHECK (quantity > 0),
    unit_price   NUMERIC(10, 2) NOT NULL CHECK (unit_price >= 0),
    line_total   NUMERIC(10, 2) NOT NULL CHECK (line_total >= 0)
);

CREATE INDEX IF NOT EXISTS idx_orders_status
    ON orders(status);

CREATE INDEX IF NOT EXISTS idx_orders_created_at
    ON orders(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id
    ON order_items(order_id);

-- Initial menu data. Re-running this file will not duplicate these items.
INSERT INTO menu_items (name, description, price)
VALUES
    ('Chicken Biryani', 'Fragrant basmati rice with spiced chicken', 350.00),
    ('Beef Karahi', 'Traditional beef karahi with tomatoes and spices', 950.00),
    ('Chicken Tikka', 'Grilled chicken marinated in house spices', 420.00),
    ('Daal Fry', 'Yellow lentils tempered with garlic and spices', 220.00),
    ('Fresh Naan', 'Freshly baked tandoor naan', 40.00)
ON CONFLICT (name) DO NOTHING;
