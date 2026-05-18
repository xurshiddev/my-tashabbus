package streets

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/mfys"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Service struct {
	store Store
	mfys  *mfys.Service
	users *users.Service
}

func NewService(store Store, mfysService *mfys.Service, usersService *users.Service) *Service {
	return &Service{store: store, mfys: mfysService, users: usersService}
}

func (s *Service) Create(ctx context.Context, current users.User, input CreateStreetInput) (Street, error) {
	if !s.canManageMFY(current, input.MFYID) {
		return Street{}, ErrForbidden
	}
	if _, err := s.mfys.Get(ctx, current, input.MFYID); err != nil {
		return Street{}, translateMFYError(err)
	}
	if err := validateCreate(input); err != nil {
		return Street{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	input.CreatedByUserID = &current.ID
	return s.store.Create(ctx, input)
}

func (s *Service) ListByMFY(ctx context.Context, current users.User, mfyID uuid.UUID, limit, offset int) ([]Street, error) {
	if !s.canManageMFY(current, mfyID) {
		return nil, ErrForbidden
	}
	if err := validatePagination(limit, offset, 200); err != nil {
		return nil, err
	}
	return s.store.ListByMFYID(ctx, mfyID, limit, offset)
}

func (s *Service) Get(ctx context.Context, current users.User, id uuid.UUID) (Street, error) {
	street, err := s.store.GetByID(ctx, id)
	if err != nil {
		return Street{}, err
	}
	if s.canManageMFY(current, street.MFYID) {
		return street, nil
	}
	if current.Role == users.RoleStreetLeader {
		assigned, err := s.store.IsUserAssignedToStreet(ctx, id, current.ID)
		if err != nil {
			return Street{}, err
		}
		if assigned {
			return street, nil
		}
	}
	return Street{}, ErrForbidden
}

func (s *Service) Update(ctx context.Context, current users.User, id uuid.UUID, input UpdateStreetInput) (Street, error) {
	street, err := s.store.GetByID(ctx, id)
	if err != nil {
		return Street{}, err
	}
	if !s.canManageMFY(current, street.MFYID) {
		return Street{}, ErrForbidden
	}
	if err := validateUpdate(input); err != nil {
		return Street{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	return s.store.Update(ctx, id, input)
}

func (s *Service) AssignLeader(ctx context.Context, current users.User, streetID, userID uuid.UUID) (StreetLeaderAssignment, error) {
	street, err := s.store.GetByID(ctx, streetID)
	if err != nil {
		return StreetLeaderAssignment{}, err
	}
	if !s.canManageMFY(current, street.MFYID) {
		return StreetLeaderAssignment{}, ErrForbidden
	}
	leader, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return StreetLeaderAssignment{}, err
	}
	if leader.Role != users.RoleStreetLeader {
		return StreetLeaderAssignment{}, ErrStreetLeaderRoleRequired
	}
	if leader.MFYID != nil && *leader.MFYID != street.MFYID {
		return StreetLeaderAssignment{}, ErrStreetLeaderWrongMFY
	}
	if leader.MFYID == nil {
		if _, err := s.users.AssignToMFY(ctx, leader.ID, street.MFYID); err != nil {
			return StreetLeaderAssignment{}, err
		}
	}
	return s.store.ReassignLeader(ctx, streetID, userID, &current.ID)
}

func (s *Service) AssignLeaderByTelegramID(ctx context.Context, current users.User, streetID uuid.UUID, telegramID int64) (StreetLeaderAssignment, error) {
	leader, err := s.users.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return StreetLeaderAssignment{}, err
	}
	return s.AssignLeader(ctx, current, streetID, leader.ID)
}

func (s *Service) GetLeader(ctx context.Context, current users.User, streetID uuid.UUID) (AssignmentWithUser, error) {
	street, err := s.Get(ctx, current, streetID)
	if err != nil {
		return AssignmentWithUser{}, err
	}
	assignment, err := s.store.GetActiveLeader(ctx, street.ID)
	if err != nil {
		return AssignmentWithUser{}, err
	}
	if assignment == nil {
		return AssignmentWithUser{}, nil
	}
	user, err := s.users.GetByID(ctx, assignment.UserID)
	if err != nil {
		return AssignmentWithUser{Assignment: assignment}, nil
	}
	return AssignmentWithUser{Assignment: assignment, User: &user}, nil
}

func (s *Service) MyStreets(ctx context.Context, current users.User, limit, offset int) ([]Street, error) {
	if err := validatePagination(limit, offset, 200); err != nil {
		return nil, err
	}
	switch current.Role {
	case users.RoleStreetLeader:
		return s.store.ListForStreetLeader(ctx, current.ID, limit, offset)
	case users.RoleMFYChairman:
		if current.MFYID == nil {
			return nil, ErrForbidden
		}
		return s.store.ListByMFYID(ctx, *current.MFYID, limit, offset)
	default:
		return []Street{}, nil
	}
}

func (s *Service) canManageMFY(current users.User, mfyID uuid.UUID) bool {
	return s.mfys.CanAccessMFY(current, mfyID)
}

func validateCreate(input CreateStreetInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return ErrNameRequired
	}
	if input.PlannedHouseholdsCount < 0 {
		return ErrPlannedHouseholdsCount
	}
	return nil
}

func validateUpdate(input UpdateStreetInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return ErrNameRequired
	}
	if input.PlannedHouseholdsCount < 0 {
		return ErrPlannedHouseholdsCount
	}
	return nil
}

func validatePagination(limit, offset, max int) error {
	if limit < 1 || limit > max || offset < 0 {
		return ErrInvalidPagination
	}
	return nil
}

func translateMFYError(err error) error {
	if err == mfys.ErrForbidden {
		return ErrForbidden
	}
	return err
}
