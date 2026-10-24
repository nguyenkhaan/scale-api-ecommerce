-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    email text NOT NULL,
    email_normalized text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    role text NOT NULL DEFAULT 'customer',
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_email_normalized_check
        CHECK (email_normalized = lower(btrim(email_normalized))),
    CONSTRAINT users_role_check CHECK (role IN ('customer', 'admin')),
    CONSTRAINT users_status_check CHECK (status IN ('active', 'disabled'))
);

-- +goose Down
DROP TABLE users;
