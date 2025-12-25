CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS orders (
                                      order_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id text NOT NULL,
    amount bigint NOT NULL CHECK (amount > 0),
    description text NOT NULL,
    status text NOT NULL CHECK (status IN ('NEW', 'FINISHED', 'CANCELLED')),
    created_at timestamptz NOT NULL DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS orders_user_created_idx
    ON orders (user_id, created_at DESC, order_id DESC);

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

CREATE TABLE IF NOT EXISTS inbox (
                                     message_id uuid PRIMARY KEY,
                                     processed_at timestamptz NOT NULL DEFAULT now()
    );
