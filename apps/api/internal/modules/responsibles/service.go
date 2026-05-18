package responsibles

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/households"
	"github.com/my-tashabbus/api/internal/modules/streets"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Service struct {
	store      Store
	streets    *streets.Service
	users      *users.Service
	households *households.Service
}

func NewService(store Store, streetService *streets.Service, userService *users.Service, householdService *households.Service) *Service {
	return &Service{store: store, streets: streetService, users: userService, households: householdService}
}

func (s *Service) Create(ctx context.Context, current users.User, streetID uuid.UUID, input CreateAssignmentInput) (Assignment, error) {
	if current.Role == users.RoleResponsiblePerson {
		return Assignment{}, ErrForbidden
	}
	street, err := s.streets.Get(ctx, current, streetID)
	if err != nil {
		return Assignment{}, translateStreetError(err)
	}
	input.StreetID = street.ID
	input.AssignedByUserID = &current.ID
	input.FromHouseNumber = strings.TrimSpace(input.FromHouseNumber)
	input.ToHouseNumber = strings.TrimSpace(input.ToHouseNumber)
	if err := validateCreate(input); err != nil {
		return Assignment{}, err
	}
	responsible, err := s.responsibleUser(ctx, input)
	if err != nil {
		return Assignment{}, err
	}
	input.ResponsibleUserID = responsible.ID
	if responsible.Role != users.RoleResponsiblePerson {
		return Assignment{}, ErrResponsibleRoleRequired
	}
	if !responsible.IsActive {
		return Assignment{}, ErrResponsibleUserInactive
	}
	if responsible.MFYID != nil && *responsible.MFYID != street.MFYID {
		return Assignment{}, ErrResponsibleWrongMFY
	}
	if responsible.MFYID == nil {
		if _, err := s.users.AssignToMFY(ctx, responsible.ID, street.MFYID); err != nil {
			return Assignment{}, err
		}
	}
	assignment, err := s.store.Create(ctx, input)
	if err != nil {
		return Assignment{}, err
	}
	if from, to, ok := numericRange(input.FromHouseNumber, input.ToHouseNumber); ok {
		_ = s.households.AssignResponsibleByNumericRange(ctx, street.ID, responsible.ID, from, to, &current.ID)
	}
	return assignment, nil
}

func (s *Service) responsibleUser(ctx context.Context, input CreateAssignmentInput) (users.User, error) {
	if input.TelegramID != nil {
		return s.users.GetByTelegramID(ctx, *input.TelegramID)
	}
	return s.users.GetByID(ctx, input.ResponsibleUserID)
}

func (s *Service) ListByStreet(ctx context.Context, current users.User, streetID uuid.UUID, limit, offset int) ([]Assignment, error) {
	if _, err := s.streets.Get(ctx, current, streetID); err != nil {
		return nil, translateStreetError(err)
	}
	if err := validatePagination(limit, offset); err != nil {
		return nil, err
	}
	return s.store.ListByStreetID(ctx, streetID, limit, offset, true)
}

func (s *Service) Deactivate(ctx context.Context, current users.User, id uuid.UUID) (Assignment, error) {
	assignment, err := s.store.GetByID(ctx, id)
	if err != nil {
		return Assignment{}, err
	}
	if current.Role == users.RoleResponsiblePerson {
		return Assignment{}, ErrForbidden
	}
	if _, err := s.streets.Get(ctx, current, assignment.StreetID); err != nil {
		return Assignment{}, translateStreetError(err)
	}
	return s.store.Deactivate(ctx, id)
}

func validateCreate(input CreateAssignmentInput) error {
	if strings.TrimSpace(input.FromHouseNumber) == "" || strings.TrimSpace(input.ToHouseNumber) == "" {
		return ErrHouseRangeRequired
	}
	if from, to, ok := numericRange(input.FromHouseNumber, input.ToHouseNumber); ok && from > to {
		return ErrInvalidHouseRange
	}
	return nil
}

func validatePagination(limit, offset int) error {
	if limit < 1 || limit > 200 || offset < 0 {
		return ErrInvalidPagination
	}
	return nil
}

func numericRange(fromRaw, toRaw string) (int, int, bool) {
	from, errFrom := strconv.Atoi(strings.TrimSpace(fromRaw))
	to, errTo := strconv.Atoi(strings.TrimSpace(toRaw))
	return from, to, errFrom == nil && errTo == nil
}

func translateStreetError(err error) error {
	if err == streets.ErrForbidden {
		return ErrForbidden
	}
	return err
}
