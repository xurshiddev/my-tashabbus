package mfys

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Store interface {
	Create(ctx context.Context, input CreateMFYInput) (MFY, error)
	GetByID(ctx context.Context, id uuid.UUID) (MFY, error)
	List(ctx context.Context, limit, offset int) ([]MFY, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateMFYInput) (MFY, error)
	SetActive(ctx context.Context, id uuid.UUID, active bool) (MFY, error)
}

type Service struct {
	store Store
	users *users.Service
}

func NewService(store Store, usersService *users.Service) *Service {
	return &Service{store: store, users: usersService}
}

func (s *Service) Create(ctx context.Context, current users.User, input CreateMFYInput) (MFY, error) {
	if current.Role != users.RoleSuperAdmin {
		return MFY{}, ErrForbidden
	}
	if err := validateCreate(input); err != nil {
		return MFY{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	input.CreatedByUserID = &current.ID
	return s.store.Create(ctx, input)
}

func (s *Service) Get(ctx context.Context, current users.User, id uuid.UUID) (MFY, error) {
	mfy, err := s.store.GetByID(ctx, id)
	if err != nil {
		return MFY{}, err
	}
	if !s.canAccessMFY(current, id) {
		return MFY{}, ErrForbidden
	}
	return mfy, nil
}

func (s *Service) List(ctx context.Context, current users.User, limit, offset int) ([]MFY, error) {
	if current.Role != users.RoleSuperAdmin {
		return nil, ErrForbidden
	}
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, ErrInvalidPagination
	}
	return s.store.List(ctx, limit, offset)
}

func (s *Service) Update(ctx context.Context, current users.User, id uuid.UUID, input UpdateMFYInput) (MFY, error) {
	if !s.canAccessMFY(current, id) {
		return MFY{}, ErrForbidden
	}
	if err := validateUpdate(input); err != nil {
		return MFY{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	return s.store.Update(ctx, id, input)
}

func (s *Service) AssignChairman(ctx context.Context, current users.User, mfyID, userID uuid.UUID) (users.User, error) {
	if current.Role != users.RoleSuperAdmin {
		return users.User{}, ErrForbidden
	}
	if _, err := s.store.GetByID(ctx, mfyID); err != nil {
		return users.User{}, err
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return users.User{}, err
	}
	if user.Role != users.RoleMFYChairman {
		return users.User{}, ErrChairmanRoleRequired
	}
	return s.users.AssignToMFY(ctx, userID, mfyID)
}

func (s *Service) CanAccessMFY(current users.User, mfyID uuid.UUID) bool {
	return s.canAccessMFY(current, mfyID)
}

func (s *Service) canAccessMFY(current users.User, mfyID uuid.UUID) bool {
	if current.Role == users.RoleSuperAdmin {
		return true
	}
	if current.Role == users.RoleMFYChairman && current.MFYID != nil && *current.MFYID == mfyID {
		return true
	}
	return false
}

func validateCreate(input CreateMFYInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return ErrNameRequired
	}
	if input.TargetVotes != nil && *input.TargetVotes < 0 {
		return ErrTargetVotesNegative
	}
	return nil
}

func validateUpdate(input UpdateMFYInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return ErrNameRequired
	}
	if input.TargetVotes != nil && *input.TargetVotes < 0 {
		return ErrTargetVotesNegative
	}
	return nil
}
