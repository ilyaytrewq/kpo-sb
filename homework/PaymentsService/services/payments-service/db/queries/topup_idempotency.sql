-- name: InsertTopupIdempotency :one
WITH ins AS (
INSERT INTO topup_idempotency (user_id, idempotency_key, amount, balance_after)
VALUES ($1, $2, $3, 0)
ON CONFLICT (user_id, idempotency_key) DO NOTHING
    RETURNING 1 AS inserted
    )
SELECT COALESCE((SELECT inserted FROM ins), 0)::bigint AS inserted;

-- name: GetTopupIdempotency :one
SELECT user_id, idempotency_key, amount, balance_after
FROM topup_idempotency
WHERE user_id = $1 AND idempotency_key = $2;

-- name: SetTopupIdempotencyBalance :one
UPDATE topup_idempotency
SET balance_after = $3
WHERE user_id = $1 AND idempotency_key = $2
RETURNING balance_after;

-- name: DeleteTopupIdempotency :exec
DELETE FROM topup_idempotency
WHERE user_id = $1 AND idempotency_key = $2;
