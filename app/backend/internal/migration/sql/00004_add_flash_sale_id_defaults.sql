-- +goose Up
ALTER TABLE flash_sale_campaigns
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

ALTER TABLE flash_sale_items
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- +goose Down
ALTER TABLE flash_sale_items
    ALTER COLUMN id DROP DEFAULT;

ALTER TABLE flash_sale_campaigns
    ALTER COLUMN id DROP DEFAULT;
