-- +goose Up
-- UNIQUE (cart_id, sku_id) already serves every lookup by cart_id.
DROP INDEX IF EXISTS cart_items_by_cart;

-- Supports the ON DELETE RESTRICT check from flash_sale_items and lookups by sale item.
CREATE INDEX cart_items_by_flash_sale_item ON cart_items (flash_sale_item_id);

-- +goose Down
DROP INDEX IF EXISTS cart_items_by_flash_sale_item;

CREATE INDEX cart_items_by_cart ON cart_items (cart_id);
