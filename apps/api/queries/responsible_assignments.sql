-- name: CreateResponsibleAssignment :one
INSERT INTO responsible_assignments (
    street_id,
    responsible_user_id,
    assigned_by_user_id,
    from_house_number,
    to_house_number
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetResponsibleAssignmentByID :one
SELECT * FROM responsible_assignments
WHERE id = $1;

-- name: ListResponsibleAssignmentsByStreetID :many
SELECT * FROM responsible_assignments
WHERE street_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveResponsibleAssignmentsByStreetID :many
SELECT * FROM responsible_assignments
WHERE street_id = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListResponsibleAssignmentsByResponsibleUserID :many
SELECT * FROM responsible_assignments
WHERE responsible_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeactivateResponsibleAssignment :one
UPDATE responsible_assignments
SET is_active = false, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: IsResponsibleAssignedToStreet :one
SELECT EXISTS (
    SELECT 1 FROM responsible_assignments
    WHERE street_id = $1 AND responsible_user_id = $2 AND is_active = true
);
