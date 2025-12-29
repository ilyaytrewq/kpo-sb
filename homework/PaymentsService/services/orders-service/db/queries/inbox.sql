-- name: InsertInboxCheck :one
WITH ins AS (
INSERT INTO inbox (message_id)
VALUES ($1)
ON CONFLICT (message_id) DO NOTHING
    RETURNING 1 AS inserted
    )
SELECT COALESCE((SELECT inserted FROM ins), 0) AS inserted;
