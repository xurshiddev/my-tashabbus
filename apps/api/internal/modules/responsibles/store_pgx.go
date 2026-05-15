package responsibles

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxStore struct {
	pool *pgxpool.Pool
}

func NewPgxStore(pool *pgxpool.Pool) *PgxStore {
	return &PgxStore{pool: pool}
}

func (s *PgxStore) Create(ctx context.Context, input CreateAssignmentInput) (Assignment, error) {
	const query = `
INSERT INTO responsible_assignments (street_id, responsible_user_id, assigned_by_user_id, from_house_number, to_house_number)
VALUES ($1, $2, $3::uuid, $4, $5)
RETURNING id::text, street_id::text, responsible_user_id::text, assigned_by_user_id::text, from_house_number, to_house_number, is_active, created_at, updated_at`
	return scanAssignment(s.pool.QueryRow(ctx, query, input.StreetID, input.ResponsibleUserID, nullableUUIDArg(input.AssignedByUserID), input.FromHouseNumber, input.ToHouseNumber))
}

func (s *PgxStore) GetByID(ctx context.Context, id uuid.UUID) (Assignment, error) {
	const query = `
SELECT id::text, street_id::text, responsible_user_id::text, assigned_by_user_id::text, from_house_number, to_house_number, is_active, created_at, updated_at
FROM responsible_assignments WHERE id = $1`
	return scanAssignment(s.pool.QueryRow(ctx, query, id))
}

func (s *PgxStore) ListByStreetID(ctx context.Context, streetID uuid.UUID, limit, offset int, activeOnly bool) ([]Assignment, error) {
	const query = `
SELECT id::text, street_id::text, responsible_user_id::text, assigned_by_user_id::text, from_house_number, to_house_number, is_active, created_at, updated_at
FROM responsible_assignments
WHERE street_id = $1 AND ($2 = false OR is_active = true)
ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	return s.list(ctx, query, streetID, activeOnly, limit, offset)
}

func (s *PgxStore) ListByResponsibleUserID(ctx context.Context, responsibleUserID uuid.UUID, limit, offset int) ([]Assignment, error) {
	const query = `
SELECT id::text, street_id::text, responsible_user_id::text, assigned_by_user_id::text, from_house_number, to_house_number, is_active, created_at, updated_at
FROM responsible_assignments
WHERE responsible_user_id = $1
ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	return s.list(ctx, query, responsibleUserID, limit, offset)
}

func (s *PgxStore) Deactivate(ctx context.Context, id uuid.UUID) (Assignment, error) {
	const query = `
UPDATE responsible_assignments SET is_active = false, updated_at = now()
WHERE id = $1
RETURNING id::text, street_id::text, responsible_user_id::text, assigned_by_user_id::text, from_house_number, to_house_number, is_active, created_at, updated_at`
	return scanAssignment(s.pool.QueryRow(ctx, query, id))
}

func (s *PgxStore) IsResponsibleAssignedToStreet(ctx context.Context, streetID, responsibleUserID uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM responsible_assignments WHERE street_id = $1 AND responsible_user_id = $2 AND is_active = true)`
	var assigned bool
	err := s.pool.QueryRow(ctx, query, streetID, responsibleUserID).Scan(&assigned)
	return assigned, err
}

func (s *PgxStore) list(ctx context.Context, query string, args ...any) ([]Assignment, error) {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Assignment, 0)
	for rows.Next() {
		item, err := scanAssignment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanAssignment(row scanner) (Assignment, error) {
	var idRaw, streetRaw, responsibleRaw string
	var assignedByRaw sql.NullString
	var from, to string
	var active bool
	var createdAt, updatedAt time.Time
	err := row.Scan(&idRaw, &streetRaw, &responsibleRaw, &assignedByRaw, &from, &to, &active, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Assignment{}, ErrAssignmentNotFound
		}
		return Assignment{}, err
	}
	id, _ := uuid.Parse(idRaw)
	streetID, _ := uuid.Parse(streetRaw)
	responsibleID, _ := uuid.Parse(responsibleRaw)
	return Assignment{ID: id, StreetID: streetID, ResponsibleUserID: responsibleID, AssignedByUserID: uuidPtr(assignedByRaw), FromHouseNumber: from, ToHouseNumber: to, IsActive: active, CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

func nullableUUIDArg(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return id.String()
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
