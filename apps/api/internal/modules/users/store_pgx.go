package users

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

func (s *PgxStore) Create(ctx context.Context, input CreateUserInput) (User, error) {
	const query = `
INSERT INTO users (full_name, phone, telegram_id, telegram_username, role, mfy_id)
VALUES ($1, $2, $3, $4, $5, $6::uuid)
RETURNING id::text, full_name, phone, telegram_id, telegram_username, role, mfy_id::text, is_active, created_at, updated_at`
	return scanUser(s.pool.QueryRow(ctx, query, input.FullName, input.Phone, input.TelegramID, input.TelegramUsername, input.Role, uuidString(input.MFYID)))
}

func (s *PgxStore) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	const query = `
SELECT id::text, full_name, phone, telegram_id, telegram_username, role, mfy_id::text, is_active, created_at, updated_at
FROM users WHERE id = $1`
	return scanUser(s.pool.QueryRow(ctx, query, id))
}

func (s *PgxStore) GetByTelegramID(ctx context.Context, telegramID int64) (User, error) {
	const query = `
SELECT id::text, full_name, phone, telegram_id, telegram_username, role, mfy_id::text, is_active, created_at, updated_at
FROM users WHERE telegram_id = $1`
	return scanUser(s.pool.QueryRow(ctx, query, telegramID))
}

func (s *PgxStore) List(ctx context.Context, limit, offset int) ([]User, error) {
	const query = `
SELECT id::text, full_name, phone, telegram_id, telegram_username, role, mfy_id::text, is_active, created_at, updated_at
FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := s.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *PgxStore) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (User, error) {
	const query = `
UPDATE users
SET full_name = $2, phone = $3, role = $4, mfy_id = $5::uuid, is_active = $6, updated_at = now()
WHERE id = $1
RETURNING id::text, full_name, phone, telegram_id, telegram_username, role, mfy_id::text, is_active, created_at, updated_at`
	return scanUser(s.pool.QueryRow(ctx, query, id, input.FullName, input.Phone, input.Role, uuidString(input.MFYID), input.IsActive))
}

func (s *PgxStore) SetTelegramIdentity(ctx context.Context, id uuid.UUID, input SetTelegramIdentityInput) (User, error) {
	const query = `
UPDATE users
SET telegram_id = $2, telegram_username = $3, updated_at = now()
WHERE id = $1
RETURNING id::text, full_name, phone, telegram_id, telegram_username, role, mfy_id::text, is_active, created_at, updated_at`
	return scanUser(s.pool.QueryRow(ctx, query, id, input.TelegramID, input.TelegramUsername))
}

func (s *PgxStore) Deactivate(ctx context.Context, id uuid.UUID) (User, error) {
	const query = `
UPDATE users SET is_active = false, updated_at = now()
WHERE id = $1
RETURNING id::text, full_name, phone, telegram_id, telegram_username, role, mfy_id::text, is_active, created_at, updated_at`
	return scanUser(s.pool.QueryRow(ctx, query, id))
}

func (s *PgxStore) Activate(ctx context.Context, id uuid.UUID) (User, error) {
	const query = `
UPDATE users SET is_active = true, updated_at = now()
WHERE id = $1
RETURNING id::text, full_name, phone, telegram_id, telegram_username, role, mfy_id::text, is_active, created_at, updated_at`
	return scanUser(s.pool.QueryRow(ctx, query, id))
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (User, error) {
	var (
		idRaw      string
		fullName   string
		phone      sql.NullString
		telegramID sql.NullInt64
		username   sql.NullString
		role       string
		mfyIDRaw   sql.NullString
		isActive   bool
		createdAt  time.Time
		updatedAt  time.Time
	)
	err := row.Scan(&idRaw, &fullName, &phone, &telegramID, &username, &role, &mfyIDRaw, &isActive, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		if isUniqueViolation(err) {
			return User{}, ErrTelegramIDConflict
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}

	id, err := uuid.Parse(idRaw)
	if err != nil {
		return User{}, fmt.Errorf("parse user id: %w", err)
	}
	var mfyID *uuid.UUID
	if mfyIDRaw.Valid {
		parsed, err := uuid.Parse(mfyIDRaw.String)
		if err != nil {
			return User{}, fmt.Errorf("parse mfy id: %w", err)
		}
		mfyID = &parsed
	}

	return User{
		ID:               id,
		FullName:         fullName,
		Phone:            stringPtr(phone),
		TelegramID:       int64Ptr(telegramID),
		TelegramUsername: stringPtr(username),
		Role:             Role(role),
		MFYID:            mfyID,
		IsActive:         isActive,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}

func uuidString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	value := id.String()
	return &value
}

func stringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func int64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "telegram")
	}
	return false
}
