-- +goose Up
CREATE TABLE inventories (
    sku_id uuid PRIMARY KEY REFERENCES skus (id) ON DELETE RESTRICT,
    on_hand bigint NOT NULL DEFAULT 0,
    reserved bigint NOT NULL DEFAULT 0,
    version bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT inventories_on_hand_check CHECK (on_hand >= 0),
    CONSTRAINT inventories_reserved_check
        CHECK (reserved >= 0 AND reserved <= on_hand),
    CONSTRAINT inventories_version_check CHECK (version >= 0)
);

CREATE TABLE inventory_adjustments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    sku_id uuid NOT NULL REFERENCES skus (id) ON DELETE RESTRICT,
    actor_user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    delta bigint NOT NULL,
    on_hand_before bigint NOT NULL,
    on_hand_after bigint NOT NULL,
    reason text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT inventory_adjustments_delta_check CHECK (delta <> 0),
    CONSTRAINT inventory_adjustments_before_check CHECK (on_hand_before >= 0),
    CONSTRAINT inventory_adjustments_after_check CHECK (on_hand_after >= 0),
    CONSTRAINT inventory_adjustments_balance_check
        CHECK (on_hand_after = on_hand_before + delta),
    CONSTRAINT inventory_adjustments_reason_check
        CHECK (char_length(btrim(reason)) BETWEEN 1 AND 500)
);

CREATE INDEX inventory_adjustments_history
    ON inventory_adjustments (sku_id, created_at DESC, id DESC);

-- +goose Down
DROP TABLE inventory_adjustments;
DROP TABLE inventories;
