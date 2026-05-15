package mfys

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (s *PgxStore) Create(ctx context.Context, input CreateMFYInput) (MFY, error) {
	const query = `
INSERT INTO mfys (name, region, district, target_votes, season, created_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id::text, name, region, district, target_votes, season, is_active, created_by_user_id::text, created_at, updated_at`
	return scanMFY(s.pool.QueryRow(ctx, query, input.Name, input.Region, input.District, input.TargetVotes, input.Season, input.CreatedByUserID))
}

func (s *PgxStore) GetByID(ctx context.Context, id uuid.UUID) (MFY, error) {
	const query = `
SELECT id::text, name, region, district, target_votes, season, is_active, created_by_user_id::text, created_at, updated_at
FROM mfys WHERE id = $1`
	return scanMFY(s.pool.QueryRow(ctx, query, id))
}

func (s *PgxStore) List(ctx context.Context, limit, offset int) ([]MFY, error) {
	const query = `
SELECT id::text, name, region, district, target_votes, season, is_active, created_by_user_id::text, created_at, updated_at
FROM mfys ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := s.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]MFY, 0)
	for rows.Next() {
		item, err := scanMFY(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PgxStore) Update(ctx context.Context, id uuid.UUID, input UpdateMFYInput) (MFY, error) {
	const query = `
UPDATE mfys
SET name = $2, region = $3, district = $4, target_votes = $5, season = $6, is_active = $7, updated_at = now()
WHERE id = $1
RETURNING id::text, name, region, district, target_votes, season, is_active, created_by_user_id::text, created_at, updated_at`
	return scanMFY(s.pool.QueryRow(ctx, query, id, input.Name, input.Region, input.District, input.TargetVotes, input.Season, input.IsActive))
}

func (s *PgxStore) SetActive(ctx context.Context, id uuid.UUID, active bool) (MFY, error) {
	const query = `
UPDATE mfys SET is_active = $2, updated_at = now()
WHERE id = $1
RETURNING id::text, name, region, district, target_votes, season, is_active, created_by_user_id::text, created_at, updated_at`
	return scanMFY(s.pool.QueryRow(ctx, query, id, active))
}

type scanner interface {
	Scan(dest ...any) error
}

func scanMFY(row scanner) (MFY, error) {
	var (
		idRaw        string
		name         string
		region       sql.NullString
		district     sql.NullString
		targetVotes  sql.NullInt32
		season       sql.NullString
		isActive     bool
		createdByRaw sql.NullString
		createdAt    time.Time
		updatedAt    time.Time
	)
	if err := row.Scan(&idRaw, &name, &region, &district, &targetVotes, &season, &isActive, &createdByRaw, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MFY{}, ErrMFYNotFound
		}
		return MFY{}, fmt.Errorf("scan mfy: %w", err)
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return MFY{}, err
	}
	return MFY{
		ID:              id,
		Name:            name,
		Region:          textPtr(region),
		District:        textPtr(district),
		TargetVotes:     int32Ptr(targetVotes),
		Season:          textPtr(season),
		IsActive:        isActive,
		CreatedByUserID: uuidPtr(createdByRaw),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}

func textPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func int32Ptr(value sql.NullInt32) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
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
