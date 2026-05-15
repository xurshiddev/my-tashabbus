-- name: CreateMFY :one
INSERT INTO mfys (
    name,
    region,
    district,
    target_votes,
    season,
    created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetMFYByID :one
SELECT * FROM mfys
WHERE id = $1;

-- name: ListMFYs :many
SELECT * FROM mfys
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateMFY :one
UPDATE mfys
SET
    name = $2,
    region = $3,
    district = $4,
    target_votes = $5,
    season = $6,
    is_active = $7,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetMFYActive :one
UPDATE mfys
SET is_active = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: AssignUserToMFY :one
UPDATE users
SET mfy_id = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetUsersByMFYID :many
SELECT * FROM users
WHERE mfy_id = $1
ORDER BY created_at DESC;
