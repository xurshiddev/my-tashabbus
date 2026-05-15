package households

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

func (s *PgxStore) Create(ctx context.Context, input CreateHouseholdInput) (Household, error) {
	const query = `
INSERT INTO households (mfy_id, street_id, house_number, total_numbers, contacted_numbers, voted_numbers, status, notes, assigned_responsible_user_id, created_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::uuid, $10::uuid)
RETURNING id::text, mfy_id::text, street_id::text, house_number, total_numbers, contacted_numbers, voted_numbers, status, notes, assigned_responsible_user_id::text, created_by_user_id::text, created_at, updated_at`
	item, err := scanHousehold(s.pool.QueryRow(ctx, query,
		input.MFYID,
		input.StreetID,
		input.HouseNumber,
		input.TotalNumbers,
		input.ContactedNumbers,
		input.VotedNumbers,
		string(input.Status),
		nullableStringArg(input.Notes),
		nullableUUIDArg(input.AssignedResponsibleUserID),
		nullableUUIDArg(input.CreatedByUserID),
	))
	return item, normalizeErr(err)
}

func (s *PgxStore) GetByID(ctx context.Context, id uuid.UUID) (Household, error) {
	const query = `
SELECT id::text, mfy_id::text, street_id::text, house_number, total_numbers, contacted_numbers, voted_numbers, status, notes, assigned_responsible_user_id::text, created_by_user_id::text, created_at, updated_at
FROM households WHERE id = $1`
	return scanHousehold(s.pool.QueryRow(ctx, query, id))
}

func (s *PgxStore) ListByStreetID(ctx context.Context, streetID uuid.UUID, limit, offset int) ([]Household, error) {
	const query = `
SELECT id::text, mfy_id::text, street_id::text, house_number, total_numbers, contacted_numbers, voted_numbers, status, notes, assigned_responsible_user_id::text, created_by_user_id::text, created_at, updated_at
FROM households WHERE street_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	return s.list(ctx, query, streetID, limit, offset)
}

func (s *PgxStore) ListByResponsibleUserID(ctx context.Context, responsibleUserID uuid.UUID, limit, offset int) ([]Household, error) {
	const query = `
SELECT id::text, mfy_id::text, street_id::text, house_number, total_numbers, contacted_numbers, voted_numbers, status, notes, assigned_responsible_user_id::text, created_by_user_id::text, created_at, updated_at
FROM households WHERE assigned_responsible_user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	return s.list(ctx, query, responsibleUserID, limit, offset)
}

func (s *PgxStore) ListByMFYID(ctx context.Context, mfyID uuid.UUID, limit, offset int) ([]Household, error) {
	const query = `
SELECT id::text, mfy_id::text, street_id::text, house_number, total_numbers, contacted_numbers, voted_numbers, status, notes, assigned_responsible_user_id::text, created_by_user_id::text, created_at, updated_at
FROM households WHERE mfy_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	return s.list(ctx, query, mfyID, limit, offset)
}

func (s *PgxStore) ListByStreetIDs(ctx context.Context, streetIDs []uuid.UUID, limit, offset int) ([]Household, error) {
	if len(streetIDs) == 0 {
		return []Household{}, nil
	}
	rawIDs := make([]string, 0, len(streetIDs))
	for _, id := range streetIDs {
		rawIDs = append(rawIDs, id.String())
	}
	const query = `
SELECT id::text, mfy_id::text, street_id::text, house_number, total_numbers, contacted_numbers, voted_numbers, status, notes, assigned_responsible_user_id::text, created_by_user_id::text, created_at, updated_at
FROM households WHERE street_id = ANY($1::uuid[]) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	return s.list(ctx, query, rawIDs, limit, offset)
}

func (s *PgxStore) Update(ctx context.Context, id uuid.UUID, input UpdateHouseholdInput) (Household, error) {
	const query = `
UPDATE households
SET house_number = $2, total_numbers = $3, contacted_numbers = $4, voted_numbers = $5, status = $6, notes = $7, updated_at = now()
WHERE id = $1
RETURNING id::text, mfy_id::text, street_id::text, house_number, total_numbers, contacted_numbers, voted_numbers, status, notes, assigned_responsible_user_id::text, created_by_user_id::text, created_at, updated_at`
	item, err := scanHousehold(s.pool.QueryRow(ctx, query, id, input.HouseNumber, input.TotalNumbers, input.ContactedNumbers, input.VotedNumbers, string(input.Status), nullableStringArg(input.Notes)))
	return item, normalizeErr(err)
}

func (s *PgxStore) UpdateAssignment(ctx context.Context, id uuid.UUID, responsibleUserID *uuid.UUID) (Household, error) {
	const query = `
UPDATE households
SET assigned_responsible_user_id = $2::uuid, updated_at = now()
WHERE id = $1
RETURNING id::text, mfy_id::text, street_id::text, house_number, total_numbers, contacted_numbers, voted_numbers, status, notes, assigned_responsible_user_id::text, created_by_user_id::text, created_at, updated_at`
	return scanHousehold(s.pool.QueryRow(ctx, query, id, nullableUUIDArg(responsibleUserID)))
}

func (s *PgxStore) CreateChangeLog(ctx context.Context, input CreateChangeLogInput) (ChangeLog, error) {
	const query = `
INSERT INTO household_change_logs (household_id, changed_by_user_id, field_name, old_value, new_value, note)
VALUES ($1, $2::uuid, $3, $4, $5, $6)
RETURNING id::text, household_id::text, changed_by_user_id::text, field_name, old_value, new_value, note, created_at`
	return scanChangeLog(s.pool.QueryRow(ctx, query, input.HouseholdID, nullableUUIDArg(input.ChangedByUserID), input.FieldName, nullableStringArg(input.OldValue), nullableStringArg(input.NewValue), nullableStringArg(input.Note)))
}

func (s *PgxStore) ListChangeLogs(ctx context.Context, householdID uuid.UUID, limit, offset int) ([]ChangeLog, error) {
	const query = `
SELECT id::text, household_id::text, changed_by_user_id::text, field_name, old_value, new_value, note, created_at
FROM household_change_logs WHERE household_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := s.pool.Query(ctx, query, householdID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ChangeLog, 0)
	for rows.Next() {
		item, err := scanChangeLog(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PgxStore) AssignResponsibleByNumericRange(ctx context.Context, streetID, responsibleUserID uuid.UUID, from, to int, changedBy *uuid.UUID) error {
	const query = `
WITH changed AS (
    SELECT id, assigned_responsible_user_id::text AS old_value
    FROM households
    WHERE street_id = $1
      AND trim(house_number) ~ '^[0-9]+$'
      AND trim(house_number)::int BETWEEN $3 AND $4
      AND (assigned_responsible_user_id IS NULL OR assigned_responsible_user_id <> $2)
),
updated AS (
    UPDATE households h
    SET assigned_responsible_user_id = $2, updated_at = now()
    WHERE h.id IN (SELECT id FROM changed)
    RETURNING h.id
)
INSERT INTO household_change_logs (household_id, changed_by_user_id, field_name, old_value, new_value, note)
SELECT changed.id, $5::uuid, 'assigned_responsible_user_id', changed.old_value, $2::text, 'Assigned by responsible range'
FROM changed`
	_, err := s.pool.Exec(ctx, query, streetID, responsibleUserID, from, to, nullableUUIDArg(changedBy))
	return err
}

func (s *PgxStore) list(ctx context.Context, query string, args ...any) ([]Household, error) {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Household, 0)
	for rows.Next() {
		item, err := scanHousehold(rows)
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

func scanHousehold(row scanner) (Household, error) {
	var idRaw, mfyRaw, streetRaw string
	var houseNumber, status string
	var total, contacted, voted int32
	var notes, assignedRaw, createdByRaw sql.NullString
	var createdAt, updatedAt time.Time
	err := row.Scan(&idRaw, &mfyRaw, &streetRaw, &houseNumber, &total, &contacted, &voted, &status, &notes, &assignedRaw, &createdByRaw, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Household{}, ErrHouseholdNotFound
		}
		return Household{}, err
	}
	id, _ := uuid.Parse(idRaw)
	mfyID, _ := uuid.Parse(mfyRaw)
	streetID, _ := uuid.Parse(streetRaw)
	return Household{ID: id, MFYID: mfyID, StreetID: streetID, HouseNumber: houseNumber, TotalNumbers: total, ContactedNumbers: contacted, VotedNumbers: voted, Status: Status(status), Notes: textPtr(notes), AssignedResponsibleUserID: uuidPtr(assignedRaw), CreatedByUserID: uuidPtr(createdByRaw), CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

func scanChangeLog(row scanner) (ChangeLog, error) {
	var idRaw, householdRaw string
	var changedRaw, oldValue, newValue, note sql.NullString
	var fieldName string
	var createdAt time.Time
	err := row.Scan(&idRaw, &householdRaw, &changedRaw, &fieldName, &oldValue, &newValue, &note, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ChangeLog{}, ErrChangeLogNotFound
		}
		return ChangeLog{}, err
	}
	id, _ := uuid.Parse(idRaw)
	householdID, _ := uuid.Parse(householdRaw)
	return ChangeLog{ID: id, HouseholdID: householdID, ChangedByUserID: uuidPtr(changedRaw), FieldName: fieldName, OldValue: textPtr(oldValue), NewValue: textPtr(newValue), Note: textPtr(note), CreatedAt: createdAt}, nil
}

func nullableStringArg(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableUUIDArg(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return id.String()
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
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "households_street_house_number") {
		return ErrDuplicateHousehold
	}
	return fmt.Errorf("household store: %w", err)
}
