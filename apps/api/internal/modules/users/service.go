package users

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type Store interface {
	Create(ctx context.Context, input CreateUserInput) (User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (User, error)
	List(ctx context.Context, limit, offset int) ([]User, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (User, error)
	SetTelegramIdentity(ctx context.Context, id uuid.UUID, input SetTelegramIdentityInput) (User, error)
	AssignToMFY(ctx context.Context, id uuid.UUID, mfyID uuid.UUID) (User, error)
	Deactivate(ctx context.Context, id uuid.UUID) (User, error)
	Activate(ctx context.Context, id uuid.UUID) (User, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Create(ctx context.Context, input CreateUserInput) (User, error) {
	if err := validateCreate(input); err != nil {
		return User{}, err
	}
	if s.store == nil {
		return User{}, ErrStoreNotConfigured
	}
	user, err := s.store.Create(ctx, normalizeCreate(input))
	if errors.Is(err, ErrTelegramIDConflict) {
		return User{}, err
	}
	return user, err
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	if s.store == nil {
		return User{}, ErrStoreNotConfigured
	}
	return s.store.GetByID(ctx, id)
}

func (s *Service) GetByTelegramID(ctx context.Context, telegramID int64) (User, error) {
	if s.store == nil {
		return User{}, ErrStoreNotConfigured
	}
	return s.store.GetByTelegramID(ctx, telegramID)
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]User, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, ErrInvalidPagination
	}
	if s.store == nil {
		return nil, ErrStoreNotConfigured
	}
	return s.store.List(ctx, limit, offset)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (User, error) {
	if strings.TrimSpace(input.FullName) == "" {
		return User{}, ErrFullNameRequired
	}
	if !IsValidRole(input.Role) {
		return User{}, ErrInvalidRole
	}
	if s.store == nil {
		return User{}, ErrStoreNotConfigured
	}
	input.FullName = strings.TrimSpace(input.FullName)
	return s.store.Update(ctx, id, input)
}

func (s *Service) SetTelegramIdentity(ctx context.Context, id uuid.UUID, input SetTelegramIdentityInput) (User, error) {
	if input.TelegramID == 0 {
		return User{}, ErrTelegramIDRequired
	}
	if s.store == nil {
		return User{}, ErrStoreNotConfigured
	}
	return s.store.SetTelegramIdentity(ctx, id, input)
}

func (s *Service) AssignToMFY(ctx context.Context, id uuid.UUID, mfyID uuid.UUID) (User, error) {
	if s.store == nil {
		return User{}, ErrStoreNotConfigured
	}
	return s.store.AssignToMFY(ctx, id, mfyID)
}

func (s *Service) Deactivate(ctx context.Context, id uuid.UUID) (User, error) {
	if s.store == nil {
		return User{}, ErrStoreNotConfigured
	}
	return s.store.Deactivate(ctx, id)
}

func (s *Service) Activate(ctx context.Context, id uuid.UUID) (User, error) {
	if s.store == nil {
		return User{}, ErrStoreNotConfigured
	}
	return s.store.Activate(ctx, id)
}

func validateCreate(input CreateUserInput) error {
	if strings.TrimSpace(input.FullName) == "" {
		return ErrFullNameRequired
	}
	if !IsValidRole(input.Role) {
		return ErrInvalidRole
	}
	return nil
}

func normalizeCreate(input CreateUserInput) CreateUserInput {
	input.FullName = strings.TrimSpace(input.FullName)
	return input
}
