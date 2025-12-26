CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS accounts (
                                        user_id text PRIMARY KEY,
                                        balance bigint NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at timestamptz NOT NULL DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS inbox (
                                     message_id uuid PRIMARY KEY,
                                     order_id uuid NOT NULL UNIQUE,
                                     processed_at timestamptz NOT NULL DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS account_ops (
                                           order_id uuid PRIMARY KEY,
                                           user_id text NOT NULL,
                                           delta bigint NOT NULL,
                                           created_at timestamptz NOT NULL DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS account_ops_user_idx
    ON account_ops (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS outbox (
                                      id bigserial PRIMARY KEY,
                                      topic text NOT NULL,
                                      kafka_key text NOT NULL,
                                      payload bytea NOT NULL,
                                      status text NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SENT', 'FAILED')),
    attempts int NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    sent_at timestamptz NULL,
    last_error text NULL
    );

CREATE INDEX IF NOT EXISTS outbox_pending_idx
    ON outbox (id)
    WHERE status = 'PENDING';
