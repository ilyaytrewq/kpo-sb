-- name: InsertAccountOp :one
INSERT INTO account_ops (order_id, user_id, delta)
VALUES ($1, $2, $3)
    ON CONFLICT (order_id) DO NOTHING
RETURNING order_id;
