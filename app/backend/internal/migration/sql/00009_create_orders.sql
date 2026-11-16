-- +goose Up
CREATE TABLE orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    order_number text NOT NULL UNIQUE,
    status text NOT NULL DEFAULT 'pending_payment',
    currency text NOT NULL DEFAULT 'VND',
    total_amount bigint NOT NULL,
    recipient_name text NOT NULL,
    recipient_phone text NOT NULL,
    recipient_address text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamptz NOT NULL,
    CONSTRAINT orders_status_check CHECK (status IN ('pending_payment', 'confirmed', 'payment_failed', 'cancelled', 'expired')),
    CONSTRAINT orders_currency_check CHECK (currency = 'VND'),
    CONSTRAINT orders_total_amount_check CHECK (total_amount > 0),
    CONSTRAINT orders_recipient_name_check CHECK (char_length(btrim(recipient_name)) BETWEEN 1 AND 120),
    CONSTRAINT orders_recipient_phone_check CHECK (char_length(btrim(recipient_phone)) BETWEEN 8 AND 20),
    CONSTRAINT orders_recipient_address_check CHECK (char_length(btrim(recipient_address)) BETWEEN 1 AND 500),
    CONSTRAINT orders_expiry_after_creation_check CHECK (expires_at > created_at)
);

CREATE INDEX orders_by_user_created_at ON orders (user_id, created_at DESC, id DESC);

CREATE TABLE order_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id uuid NOT NULL REFERENCES orders (id) ON DELETE RESTRICT,
    sku_id uuid NOT NULL REFERENCES skus (id) ON DELETE RESTRICT,
    sku_code text NOT NULL,
    product_name text NOT NULL,
    sku_attributes jsonb NOT NULL DEFAULT '{}'::jsonb,
    quantity integer NOT NULL,
    unit_price bigint NOT NULL,
    subtotal bigint NOT NULL,
    price_source text NOT NULL,
    flash_sale_item_id uuid REFERENCES flash_sale_items (id) ON DELETE RESTRICT,
    CONSTRAINT order_items_quantity_check CHECK (quantity > 0),
    CONSTRAINT order_items_unit_price_check CHECK (unit_price > 0),
    CONSTRAINT order_items_subtotal_check CHECK (subtotal = quantity * unit_price),
    CONSTRAINT order_items_attributes_object_check CHECK (jsonb_typeof(sku_attributes) = 'object'),
    CONSTRAINT order_items_price_source_check CHECK (
        (price_source = 'regular' AND flash_sale_item_id IS NULL) OR
        (price_source = 'flash_sale' AND flash_sale_item_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX order_items_one_sku_per_order ON order_items (order_id, sku_id);

CREATE TABLE inventory_reservations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_item_id uuid NOT NULL UNIQUE REFERENCES order_items (id) ON DELETE RESTRICT,
    sku_id uuid NOT NULL REFERENCES skus (id) ON DELETE RESTRICT,
    quantity integer NOT NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    released_at timestamptz,
    CONSTRAINT inventory_reservations_quantity_check CHECK (quantity > 0),
    CONSTRAINT inventory_reservations_status_check CHECK (status IN ('active', 'released', 'finalized')),
    CONSTRAINT inventory_reservations_released_check CHECK (
        (status = 'active' AND released_at IS NULL) OR (status <> 'active' AND released_at IS NOT NULL)
    )
);

CREATE INDEX inventory_reservations_active_by_sku ON inventory_reservations (sku_id) WHERE status = 'active';

CREATE TABLE idempotency_records (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    operation text NOT NULL,
    idempotency_key text NOT NULL,
    request_hash bytea NOT NULL,
    response_status integer,
    response_body jsonb,
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idempotency_operation_check CHECK (operation = 'place_order'),
    CONSTRAINT idempotency_key_length_check CHECK (char_length(idempotency_key) BETWEEN 8 AND 255),
    CONSTRAINT idempotency_request_hash_length_check CHECK (octet_length(request_hash) = 32),
    CONSTRAINT idempotency_completion_check CHECK (
        (completed_at IS NULL AND response_status IS NULL AND response_body IS NULL) OR
        (completed_at IS NOT NULL AND response_status BETWEEN 200 AND 599 AND response_body IS NOT NULL)
    ),
    CONSTRAINT idempotency_user_operation_key_unique UNIQUE (user_id, operation, idempotency_key)
);

-- +goose Down
DROP TABLE idempotency_records;
DROP TABLE inventory_reservations;
DROP TABLE order_items;
DROP TABLE orders;
