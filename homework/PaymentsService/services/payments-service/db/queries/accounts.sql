-- name: CreateAccount :one
INSERT INTO accounts (user_id, balance)
VALUES ($1, 0)
    ON CONFLICT (user_id) DO NOTHING
RETURNING user_id, balance;

-- name: GetBalance :one
SELECT balance FROM accounts WHERE user_id = $1;

-- name: TopUp :one
UPDATE accounts
SET balance = balance + $2
WHERE user_id = $1
    RETURNING user_id, balance;

-- name: AccountExists :one
SELECT EXISTS(SELECT 1 FROM accounts WHERE user_id = $1) AS exists;
