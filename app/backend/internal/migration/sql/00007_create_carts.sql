-- +goose Up
CREATE TABLE carts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT carts_status_check CHECK (status IN ('active', 'archived'))
);

CREATE UNIQUE INDEX carts_one_active_per_user
    ON carts (user_id) WHERE status = 'active';

CREATE TABLE cart_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id uuid NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    sku_id uuid NOT NULL REFERENCES skus (id) ON DELETE RESTRICT,
    quantity integer NOT NULL,
    intended_price_source text NOT NULL,
    flash_sale_item_id uuid REFERENCES flash_sale_items (id) ON DELETE RESTRICT,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT cart_items_cart_sku_unique UNIQUE (cart_id, sku_id),
    CONSTRAINT cart_items_quantity_check CHECK (quantity > 0),
    CONSTRAINT cart_items_price_source_check CHECK (
        (intended_price_source = 'regular' AND flash_sale_item_id IS NULL) OR
        (intended_price_source = 'flash_sale' AND flash_sale_item_id IS NOT NULL)
    ),
    CONSTRAINT cart_items_version_check CHECK (version >= 1)
);

CREATE INDEX cart_items_by_cart ON cart_items (cart_id);

ALTER TABLE flash_sale_items
    ADD CONSTRAINT flash_sale_items_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES skus (id) ON DELETE RESTRICT;

-- +goose Down
ALTER TABLE flash_sale_items
    DROP CONSTRAINT IF EXISTS flash_sale_items_sku_id_fkey;

DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;
