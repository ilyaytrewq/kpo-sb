-- name: CreateWork :one
INSERT INTO works (work_id, name, description)
VALUES ($1, $2, $3)
RETURNING work_id, name, description, created_at;

-- name: GetWork :one
SELECT work_id, name, description, created_at
FROM works
WHERE work_id = $1;
