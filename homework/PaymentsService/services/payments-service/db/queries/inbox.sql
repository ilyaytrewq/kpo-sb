-- name: InsertInboxCheck :one
WITH ins AS (
INSERT INTO inbox (message_id, order_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
    RETURNING 1 AS inserted
    )
SELECT COALESCE((SELECT inserted FROM ins), 0)::bigint AS inserted;
