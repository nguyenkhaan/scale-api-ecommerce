-- +goose Up
CREATE TABLE sessions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id),
    token_hash bytea NOT NULL UNIQUE,
    csrf_secret_hash bytea NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    last_seen_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT sessions_token_hash_length_check CHECK (octet_length(token_hash) = 32),
    CONSTRAINT sessions_csrf_secret_hash_length_check CHECK (octet_length(csrf_secret_hash) = 32)
);

CREATE INDEX sessions_active_by_expiry ON sessions (expires_at) WHERE revoked_at IS NULL;
CREATE INDEX sessions_by_user ON sessions (user_id);

CREATE TABLE login_throttles (
    email_normalized text PRIMARY KEY,
    failure_count integer NOT NULL,
    window_started_at timestamptz NOT NULL,
    CONSTRAINT login_throttles_failure_count_check CHECK (failure_count > 0)
);

-- +goose Down
DROP TABLE login_throttles;
DROP TABLE sessions;
