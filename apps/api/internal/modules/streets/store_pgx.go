package streets

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxStore struct {
	pool *pgxpool.Pool
}

func NewPgxStore(pool *pgxpool.Pool) *PgxStore {
	return &PgxStore{pool: pool}
}

func (s *PgxStore) Create(ctx context.Context, input CreateStreetInput) (Street, error) {
	const query = `
INSERT INTO streets (mfy_id, name, planned_households_count, notes, created_by_user_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id::text, mfy_id::text, name, planned_households_count, notes, is_active, created_by_user_id::text, created_at, updated_at`
	street, err := scanStreet(s.pool.QueryRow(ctx, query, input.MFYID, input.Name, input.PlannedHouseholdsCount, input.Notes, input.CreatedByUserID))
	return street, normalizeErr(err)
}

func (s *PgxStore) GetByID(ctx context.Context, id uuid.UUID) (Street, error) {
	const query = `
SELECT id::text, mfy_id::text, name, planned_households_count, notes, is_active, created_by_user_id::text, created_at, updated_at
FROM streets WHERE id = $1`
	return scanStreet(s.pool.QueryRow(ctx, query, id))
}

func (s *PgxStore) ListByMFYID(ctx context.Context, mfyID uuid.UUID, limit, offset int) ([]Street, error) {
	const query = `
SELECT id::text, mfy_id::text, name, planned_households_count, notes, is_active, created_by_user_id::text, created_at, updated_at
FROM streets WHERE mfy_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := s.pool.Query(ctx, query, mfyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Street, 0)
	for rows.Next() {
		item, err := scanStreet(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PgxStore) Update(ctx context.Context, id uuid.UUID, input UpdateStreetInput) (Street, error) {
	const query = `
UPDATE streets
SET name = $2, planned_households_count = $3, notes = $4, is_active = $5, updated_at = now()
WHERE id = $1
RETURNING id::text, mfy_id::text, name, planned_households_count, notes, is_active, created_by_user_id::text, created_at, updated_at`
	street, err := scanStreet(s.pool.QueryRow(ctx, query, id, input.Name, input.PlannedHouseholdsCount, input.Notes, input.IsActive))
	return street, normalizeErr(err)
}

func (s *PgxStore) ListForStreetLeader(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Street, error) {
	const query = `
SELECT st.id::text, st.mfy_id::text, st.name, st.planned_households_count, st.notes, st.is_active, st.created_by_user_id::text, st.created_at, st.updated_at
FROM streets st
JOIN street_leader_assignments sla ON sla.street_id = st.id
WHERE sla.user_id = $1 AND sla.is_active = true AND st.is_active = true
ORDER BY st.created_at DESC LIMIT $2 OFFSET $3`
	rows, err := s.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Street, 0)
	for rows.Next() {
		item, err := scanStreet(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PgxStore) IsUserAssignedToStreet(ctx context.Context, streetID, userID uuid.UUID) (bool, error) {
	const query = `
SELECT EXISTS (
    SELECT 1 FROM street_leader_assignments
    WHERE street_id = $1 AND user_id = $2 AND is_active = true
)`
	var assigned bool
	err := s.pool.QueryRow(ctx, query, streetID, userID).Scan(&assigned)
	return assigned, err
}

func (s *PgxStore) ReassignLeader(ctx context.Context, streetID, userID uuid.UUID, assignedBy *uuid.UUID) (StreetLeaderAssignment, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return StreetLeaderAssignment{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE street_leader_assignments SET is_active = false, updated_at = now() WHERE street_id = $1 AND is_active = true`, streetID); err != nil {
		return StreetLeaderAssignment{}, err
	}
	const query = `
INSERT INTO street_leader_assignments (street_id, user_id, assigned_by_user_id)
VALUES ($1, $2, $3)
RETURNING id::text, street_id::text, user_id::text, assigned_by_user_id::text, is_active, created_at, updated_at`
	assignment, err := scanAssignment(tx.QueryRow(ctx, query, streetID, userID, assignedBy))
	if err != nil {
		return StreetLeaderAssignment{}, normalizeErr(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return StreetLeaderAssignment{}, err
	}
	return assignment, nil
}

func (s *PgxStore) GetActiveLeader(ctx context.Context, streetID uuid.UUID) (*StreetLeaderAssignment, error) {
	const query = `
SELECT id::text, street_id::text, user_id::text, assigned_by_user_id::text, is_active, created_at, updated_at
FROM street_leader_assignments
WHERE street_id = $1 AND is_active = true`
	assignment, err := scanAssignment(s.pool.QueryRow(ctx, query, streetID))
	if errors.Is(err, ErrAssignmentNotFound) {
		return nil, nil
	}
	return &assignment, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanStreet(row scanner) (Street, error) {
	var idRaw, mfyIDRaw string
	var name string
	var planned int32
	var notes sql.NullString
	var active bool
	var createdBy sql.NullString
	var createdAt, updatedAt time.Time
	if err := row.Scan(&idRaw, &mfyIDRaw, &name, &planned, &notes, &active, &createdBy, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Street{}, ErrStreetNotFound
		}
		return Street{}, err
	}
	id, _ := uuid.Parse(idRaw)
	mfyID, _ := uuid.Parse(mfyIDRaw)
	return Street{ID: id, MFYID: mfyID, Name: name, PlannedHouseholdsCount: planned, Notes: textPtr(notes), IsActive: active, CreatedByUserID: uuidPtr(createdBy), CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

func scanAssignment(row scanner) (StreetLeaderAssignment, error) {
	var idRaw, streetIDRaw, userIDRaw string
	var assignedBy sql.NullString
	var active bool
	var createdAt, updatedAt time.Time
	if err := row.Scan(&idRaw, &streetIDRaw, &userIDRaw, &assignedBy, &active, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StreetLeaderAssignment{}, ErrAssignmentNotFound
		}
		return StreetLeaderAssignment{}, err
	}
	id, _ := uuid.Parse(idRaw)
	streetID, _ := uuid.Parse(streetIDRaw)
	userID, _ := uuid.Parse(userIDRaw)
	return StreetLeaderAssignment{ID: id, StreetID: streetID, UserID: userID, AssignedByUserID: uuidPtr(assignedBy), IsActive: active, CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

func textPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func uuidPtr(value sql.NullString) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed, err := uuid.Parse(value.String)
	if err != nil {
		return nil
	}
	return &parsed
}

func normalizeErr(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "streets_mfy_lower_name") {
		return ErrDuplicateStreetName
	}
	return fmt.Errorf("street store: %w", err)
}
