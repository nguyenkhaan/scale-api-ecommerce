-- +goose Up
-- Repair databases where 00010 is recorded as applied but its payment session table was removed.
CREATE TABLE IF NOT EXISTS payment_sessions (
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

-- +goose StatementBegin
DO $$
BEGIN
    IF to_regclass('payment_events') IS NOT NULL AND NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = to_regclass('payment_events')
          AND confrelid = to_regclass('payment_sessions')
          AND contype = 'f'
    ) THEN
        ALTER TABLE payment_events
            ADD CONSTRAINT payment_events_payment_session_id_fkey
            FOREIGN KEY (payment_session_id) REFERENCES payment_sessions (id) ON DELETE RESTRICT;
    END IF;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- Intentionally retain the repaired table: this migration may have been a no-op on healthy databases.
