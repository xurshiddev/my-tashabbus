-- name: CreateHousehold :one
INSERT INTO households (
    mfy_id,
    street_id,
    house_number,
    total_numbers,
    contacted_numbers,
    voted_numbers,
    status,
    notes,
    assigned_responsible_user_id,
    created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetHouseholdByID :one
SELECT * FROM households
WHERE id = $1;

-- name: ListHouseholdsByStreetID :many
SELECT * FROM households
WHERE street_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListHouseholdsByResponsibleUserID :many
SELECT * FROM households
WHERE assigned_responsible_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListHouseholdsByMFYID :many
SELECT * FROM households
WHERE mfy_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListHouseholdsByStreetIDs :many
SELECT * FROM households
WHERE street_id = ANY($1::uuid[])
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateHousehold :one
UPDATE households
SET
    house_number = $2,
    total_numbers = $3,
    contacted_numbers = $4,
    voted_numbers = $5,
    status = $6,
    notes = $7,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateHouseholdAssignment :one
UPDATE households
SET assigned_responsible_user_id = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetHouseholdStatus :one
UPDATE households
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteHousehold :exec
DELETE FROM households
WHERE id = $1;

-- name: CountHouseholdsByStreetID :one
SELECT count(*) FROM households
WHERE street_id = $1;
