-- +goose Up
CREATE TABLE categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    slug text NOT NULL,
    is_visible boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT categories_slug_unique UNIQUE (slug),
    CONSTRAINT categories_name_check
        CHECK (char_length(btrim(name)) BETWEEN 1 AND 200),
    CONSTRAINT categories_slug_check
        CHECK (slug = btrim(slug) AND char_length(slug) BETWEEN 1 AND 100)
);

CREATE TABLE products (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id uuid REFERENCES categories (id) ON DELETE RESTRICT,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    image_url text,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT products_name_check
        CHECK (char_length(btrim(name)) BETWEEN 1 AND 200),
    CONSTRAINT products_description_check
        CHECK (char_length(description) <= 5000),
    CONSTRAINT products_status_check
        CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX products_public_keyset
    ON products (status, created_at DESC, id DESC);
CREATE INDEX products_by_category_keyset
    ON products (category_id, status, created_at DESC, id DESC);

CREATE TABLE skus (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    code text NOT NULL,
    attributes jsonb NOT NULL DEFAULT '{}'::jsonb,
    regular_price bigint NOT NULL,
    max_quantity_per_order integer NOT NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT skus_code_unique UNIQUE (code),
    CONSTRAINT skus_code_check
        CHECK (code = btrim(code) AND char_length(code) BETWEEN 1 AND 100),
    CONSTRAINT skus_attributes_object_check
        CHECK (jsonb_typeof(attributes) = 'object'),
    CONSTRAINT skus_regular_price_check CHECK (regular_price > 0),
    CONSTRAINT skus_max_quantity_per_order_check
        CHECK (max_quantity_per_order > 0),
    CONSTRAINT skus_status_check
        CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX skus_by_product_status
    ON skus (product_id, status, id);

-- +goose Down
DROP TABLE skus;
DROP TABLE products;
DROP TABLE categories;
