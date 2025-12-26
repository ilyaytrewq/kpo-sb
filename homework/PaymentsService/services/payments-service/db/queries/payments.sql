-- name: TryDeductOnce :one
WITH upd AS (
UPDATE accounts
SET balance = accounts.balance - $3
WHERE accounts.user_id = $2
  AND accounts.balance >= $3
  AND NOT EXISTS (SELECT 1 FROM account_ops ao WHERE ao.order_id = $1)
    RETURNING balance
),
ins AS (
INSERT INTO account_ops (order_id, user_id, delta)
SELECT $1, $2, -$3
WHERE EXISTS (SELECT 1 FROM upd)
ON CONFLICT (order_id) DO NOTHING
    RETURNING 1 AS inserted
    )
SELECT
    COALESCE((SELECT balance FROM upd), 0)::bigint AS new_balance,
    COALESCE((SELECT inserted FROM ins), 0)::bigint AS op_inserted;
