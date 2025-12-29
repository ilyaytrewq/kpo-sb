-- name: InsertOutbox :one
INSERT INTO outbox (topic, kafka_key, payload)
VALUES ($1, $2, $3)
    RETURNING id;

-- name: LockUnsentOutbox :many
SELECT id, topic, kafka_key, payload, attempts
FROM outbox
WHERE sent_at IS NULL
ORDER BY id
    LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: MarkOutboxSent :exec
UPDATE outbox
SET sent_at = now(), status = 'SENT'
WHERE id = $1;

-- name: MarkOutboxAttemptFailed :exec
UPDATE outbox
SET attempts = attempts + 1, last_error = $2, status = 'FAILED'
WHERE id = $1;
