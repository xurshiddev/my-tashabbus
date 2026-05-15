-- name: CreateStreetLeaderAssignment :one
INSERT INTO street_leader_assignments (
    street_id,
    user_id,
    assigned_by_user_id
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: DeactivateActiveStreetLeaderAssignmentForStreet :exec
UPDATE street_leader_assignments
SET is_active = false, updated_at = now()
WHERE street_id = $1 AND is_active = true;

-- name: GetActiveStreetLeaderAssignmentByStreetID :one
SELECT * FROM street_leader_assignments
WHERE street_id = $1 AND is_active = true;

-- name: ListStreetLeaderAssignmentsByStreetID :many
SELECT * FROM street_leader_assignments
WHERE street_id = $1
ORDER BY created_at DESC;

-- name: ListStreetLeaderAssignmentsByUserID :many
SELECT * FROM street_leader_assignments
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: IsUserAssignedToStreet :one
SELECT EXISTS (
    SELECT 1
    FROM street_leader_assignments
    WHERE street_id = $1 AND user_id = $2 AND is_active = true
) AS is_assigned;
