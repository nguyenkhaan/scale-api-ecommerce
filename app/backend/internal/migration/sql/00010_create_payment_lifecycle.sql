-- +goose Up
ALTER TABLE orders
    ADD COLUMN updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN terminal_at timestamptz;

-- Orders that already ended before this migration get their creation time as a lower bound.
UPDATE orders SET terminal_at = created_at WHERE status <> 'pending_payment';

ALTER TABLE orders
    ADD CONSTRAINT orders_terminal_at_check CHECK ((status = 'pending_payment') = (terminal_at IS NULL));

-- The expiry worker scans pending orders by deadline.
CREATE INDEX orders_due ON orders (expires_at, id) WHERE status = 'pending_payment';

-- Checkout records which cart line (and its version) each order line came from, so a payment success
-- removes only cart lines the customer has not changed since (BR-13). NULL keeps the cart untouched.
ALTER TABLE order_items
    ADD COLUMN cart_item_id uuid REFERENCES cart_items (id) ON DELETE SET NULL,
    ADD COLUMN cart_item_version bigint;

CREATE INDEX order_items_by_cart_item ON order_items (cart_item_id) WHERE cart_item_id IS NOT NULL;

CREATE TABLE payment_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id uuid NOT NULL UNIQUE REFERENCES orders (id) ON DELETE RESTRICT,
    status text NOT NULL DEFAULT 'pending',
    amount bigint NOT NULL,
    currency text NOT NULL DEFAULT 'VND',
    simulator_reference text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT payment_sessions_status_check CHECK (status IN ('pending', 'unknown', 'succeeded', 'failed')),
    CONSTRAINT payment_sessions_amount_check CHECK (amount > 0),
    CONSTRAINT payment_sessions_currency_check CHECK (currency = 'VND')
);

-- Every received delivery is stored. Verified events are deduplicated by the simulator's event_id;
-- unverified ones (bad signature) are kept for audit but can never block a later genuine event.
CREATE TABLE payment_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL,
    payment_session_id uuid REFERENCES payment_sessions (id) ON DELETE RESTRICT,
    event_type text NOT NULL,
    amount bigint NOT NULL,
    currency text NOT NULL,
    payload_hash bytea NOT NULL,
    source_verified boolean NOT NULL,
    result text NOT NULL,
    reason_code text,
    occurred_at timestamptz,
    received_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT payment_events_type_check CHECK (event_type IN ('succeeded', 'failed')),
    CONSTRAINT payment_events_result_check CHECK (result IN ('applied', 'duplicate', 'rejected', 'late', 'conflict')),
    CONSTRAINT payment_events_payload_hash_check CHECK (octet_length(payload_hash) = 32),
    CONSTRAINT payment_events_unverified_rejected_check CHECK (source_verified OR result = 'rejected')
);

CREATE UNIQUE INDEX payment_events_verified_event_id ON payment_events (event_id) WHERE source_verified;
CREATE INDEX payment_events_by_session ON payment_events (payment_session_id, received_at);

CREATE TABLE payment_reconciliations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_event_id uuid NOT NULL UNIQUE REFERENCES payment_events (id) ON DELETE RESTRICT,
    order_id uuid NOT NULL REFERENCES orders (id) ON DELETE RESTRICT,
    reason text NOT NULL,
    status text NOT NULL DEFAULT 'open',
    reviewed_by uuid REFERENCES users (id) ON DELETE RESTRICT,
    review_note text,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reviewed_at timestamptz,
    CONSTRAINT payment_reconciliations_status_check CHECK (status IN ('open', 'reviewed')),
    CONSTRAINT payment_reconciliations_review_check CHECK (
        (status = 'open' AND reviewed_at IS NULL) OR (status = 'reviewed' AND reviewed_at IS NOT NULL)
    )
);

CREATE INDEX payment_reconciliations_by_order ON payment_reconciliations (order_id);
CREATE INDEX payment_reconciliations_open ON payment_reconciliations (created_at, id) WHERE status = 'open';

-- +goose Down
DROP TABLE payment_reconciliations;
DROP TABLE payment_events;
DROP TABLE payment_sessions;
DROP INDEX order_items_by_cart_item;
ALTER TABLE order_items DROP COLUMN cart_item_version, DROP COLUMN cart_item_id;
DROP INDEX orders_due;
ALTER TABLE orders DROP CONSTRAINT orders_terminal_at_check, DROP COLUMN terminal_at, DROP COLUMN updated_at;
