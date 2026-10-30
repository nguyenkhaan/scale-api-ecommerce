-- +goose Up
CREATE TABLE flash_sale_campaigns (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    status text NOT NULL DEFAULT 'draft',
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT flash_sale_campaigns_status_check
        CHECK (status IN ('draft', 'scheduled', 'stopped')),
    CONSTRAINT flash_sale_campaigns_time_check CHECK (ends_at > starts_at)
);

CREATE TABLE flash_sale_items (
    id uuid PRIMARY KEY,
    campaign_id uuid NOT NULL REFERENCES flash_sale_campaigns (id),
    sku_id uuid NOT NULL,
    sale_price bigint NOT NULL,
    quota_total bigint NOT NULL,
    quota_reserved bigint NOT NULL DEFAULT 0,
    quota_sold bigint NOT NULL DEFAULT 0,
    purchase_limit integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT flash_sale_items_campaign_sku_unique UNIQUE (campaign_id, sku_id),
    CONSTRAINT flash_sale_items_sale_price_check CHECK (sale_price > 0),
    CONSTRAINT flash_sale_items_quota_total_check CHECK (quota_total > 0),
    CONSTRAINT flash_sale_items_purchase_limit_check CHECK (purchase_limit > 0),
    CONSTRAINT flash_sale_items_quota_reserved_check CHECK (quota_reserved >= 0),
    CONSTRAINT flash_sale_items_quota_sold_check CHECK (quota_sold >= 0),
    CONSTRAINT flash_sale_items_quota_limit_check
        CHECK (quota_reserved + quota_sold <= quota_total)
);

CREATE INDEX flash_sale_items_by_sku ON flash_sale_items (sku_id, campaign_id);

CREATE TABLE flash_sale_customer_usage (
    flash_sale_item_id uuid NOT NULL REFERENCES flash_sale_items (id),
    user_id uuid NOT NULL REFERENCES users (id),
    reserved_quantity bigint NOT NULL DEFAULT 0,
    sold_quantity bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (flash_sale_item_id, user_id),
    CONSTRAINT flash_sale_customer_usage_reserved_check CHECK (reserved_quantity >= 0),
    CONSTRAINT flash_sale_customer_usage_sold_check CHECK (sold_quantity >= 0)
);

-- +goose Down
DROP TABLE flash_sale_customer_usage;
DROP TABLE flash_sale_items;
DROP TABLE flash_sale_campaigns;
