-- name: Add :one
INSERT INTO users (
    first_name,
    last_name,
    nickname,
    password,
    email,
    country,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
RETURNING *;

-- name: Update :one
UPDATE users SET
    first_name = $2,
    last_name = $3,
    nickname = $4,
    password = $5,
    email = $6,
    country = $7,
    updated_at = NOW()
WHERE
    id = $1
RETURNING *;

-- name: Remove :one
DELETE FROM users
WHERE
    id = $1
RETURNING *;

-- name: ListPaginated :many
SELECT
    *,
    count(*) OVER() AS total_elements
FROM users
WHERE $3
OFFSET $1
LIMIT $2;
