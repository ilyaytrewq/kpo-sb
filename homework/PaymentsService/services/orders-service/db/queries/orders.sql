-- name: CreateOrder :one
INSERT INTO orders (user_id, amount, description, status)
VALUES ($1, $2, $3, 'NEW')
    RETURNING order_id, user_id, amount, description, status, created_at;

-- name: CreateOrderIdempotent :one
INSERT INTO orders (user_id, amount, description, status, idempotency_key)
VALUES ($1, $2, $3, 'NEW', $4)
    ON CONFLICT (user_id, idempotency_key) DO NOTHING
RETURNING order_id, user_id, amount, description, status, created_at, idempotency_key;

-- name: GetOrderByIdempotency :one
SELECT order_id, user_id, amount, description, status, created_at, idempotency_key
FROM orders
WHERE user_id = $1 AND idempotency_key = $2;

-- name: GetOrder :one
SELECT order_id, user_id, amount, description, status, created_at
FROM orders
WHERE order_id = $1 AND user_id = $2;

-- name: ListOrders :many
SELECT order_id, user_id, amount, description, status, created_at
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC, order_id DESC
    LIMIT $2 OFFSET $3;

-- Важно для consumer: обновляем статус только если он ещё NEW (идемпотентно)
-- name: UpdateOrderStatusIfNew :exec
UPDATE orders
SET status = $2
WHERE order_id = $1 AND status = 'NEW';
