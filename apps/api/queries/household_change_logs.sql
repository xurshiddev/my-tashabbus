-- name: CreateHouseholdChangeLog :one
INSERT INTO household_change_logs (
    household_id,
    changed_by_user_id,
    field_name,
    old_value,
    new_value,
    note
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ListHouseholdChangeLogsByHouseholdID :many
SELECT * FROM household_change_logs
WHERE household_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
