-- +goose Up
-- Restore payment lifecycle tables that were removed while 00010 remained marked as applied.
CREATE TABLE IF NOT EXISTS payment_events (
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

CREATE UNIQUE INDEX IF NOT EXISTS payment_events_verified_event_id
    ON payment_events (event_id) WHERE source_verified;
CREATE INDEX IF NOT EXISTS payment_events_by_session
    ON payment_events (payment_session_id, received_at);

CREATE TABLE IF NOT EXISTS payment_reconciliations (
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

CREATE INDEX IF NOT EXISTS payment_reconciliations_by_order
    ON payment_reconciliations (order_id);
CREATE INDEX IF NOT EXISTS payment_reconciliations_open
    ON payment_reconciliations (created_at, id) WHERE status = 'open';

-- +goose Down
-- Intentionally keep repaired data and tables; this migration may have been a no-op on healthy databases.
