-- name: CreateStreet :one
INSERT INTO streets (
    mfy_id,
    name,
    planned_households_count,
    notes,
    created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetStreetByID :one
SELECT * FROM streets
WHERE id = $1;

-- name: ListStreetsByMFYID :many
SELECT * FROM streets
WHERE mfy_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateStreet :one
UPDATE streets
SET
    name = $2,
    planned_households_count = $3,
    notes = $4,
    is_active = $5,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetStreetActive :one
UPDATE streets
SET is_active = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListStreetsForStreetLeader :many
SELECT s.*
FROM streets s
JOIN street_leader_assignments sla ON sla.street_id = s.id
WHERE sla.user_id = $1 AND sla.is_active = true AND s.is_active = true
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;
