-- name: CreateUser :one
INSERT INTO users (
    full_name,
    phone,
    telegram_id,
    telegram_username,
    role,
    mfy_id
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByTelegramID :one
SELECT * FROM users
WHERE telegram_id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET
    full_name = $2,
    phone = $3,
    role = $4,
    mfy_id = $5,
    is_active = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetUserTelegramIdentity :one
UPDATE users
SET
    telegram_id = $2,
    telegram_username = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeactivateUser :one
UPDATE users
SET is_active = false, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ActivateUser :one
UPDATE users
SET is_active = true, updated_at = now()
WHERE id = $1
RETURNING *;
