package households

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/streets"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Service struct {
	store   Store
	streets *streets.Service
}

func NewService(store Store, streetService *streets.Service) *Service {
	return &Service{store: store, streets: streetService}
}

func (s *Service) Create(ctx context.Context, current users.User, streetID uuid.UUID, input CreateHouseholdInput) (Household, error) {
	if current.Role == users.RoleResponsiblePerson {
		return Household{}, ErrForbidden
	}
	street, err := s.streets.Get(ctx, current, streetID)
	if err != nil {
		return Household{}, translateStreetError(err)
	}
	input.StreetID = street.ID
	input.MFYID = street.MFYID
	input.CreatedByUserID = &current.ID
	input = normalizeCreate(input)
	if err := validateCreate(input); err != nil {
		return Household{}, err
	}
	return s.store.Create(ctx, input)
}

func (s *Service) ListByStreet(ctx context.Context, current users.User, streetID uuid.UUID, limit, offset int) ([]Household, error) {
	if _, err := s.streets.Get(ctx, current, streetID); err != nil {
		return nil, translateStreetError(err)
	}
	if err := validatePagination(limit, offset); err != nil {
		return nil, err
	}
	return s.store.ListByStreetID(ctx, streetID, limit, offset)
}

func (s *Service) Get(ctx context.Context, current users.User, id uuid.UUID) (Household, error) {
	household, err := s.store.GetByID(ctx, id)
	if err != nil {
		return Household{}, err
	}
	if err := s.ensureCanAccessHousehold(ctx, current, household); err != nil {
		return Household{}, err
	}
	return household, nil
}

func (s *Service) Update(ctx context.Context, current users.User, id uuid.UUID, input UpdateHouseholdInput) (Household, error) {
	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return Household{}, err
	}
	if err := s.ensureCanAccessHousehold(ctx, current, existing); err != nil {
		return Household{}, err
	}
	input = normalizeUpdate(input)
	if err := validateUpdate(input); err != nil {
		return Household{}, err
	}
	updated, err := s.store.Update(ctx, id, input)
	if err != nil {
		return Household{}, err
	}
	s.writeUpdateLogs(ctx, current.ID, existing, updated)
	return updated, nil
}

func (s *Service) MyHouseholds(ctx context.Context, current users.User, limit, offset int) ([]Household, error) {
	if err := validatePagination(limit, offset); err != nil {
		return nil, err
	}
	switch current.Role {
	case users.RoleResponsiblePerson:
		return s.store.ListByResponsibleUserID(ctx, current.ID, limit, offset)
	case users.RoleStreetLeader:
		streetList, err := s.streets.MyStreets(ctx, current, 200, 0)
		if err != nil {
			return nil, translateStreetError(err)
		}
		ids := make([]uuid.UUID, 0, len(streetList))
		for _, street := range streetList {
			ids = append(ids, street.ID)
		}
		if len(ids) == 0 {
			return []Household{}, nil
		}
		return s.store.ListByStreetIDs(ctx, ids, limit, offset)
	case users.RoleMFYChairman:
		if current.MFYID == nil {
			return nil, ErrForbidden
		}
		return s.store.ListByMFYID(ctx, *current.MFYID, limit, offset)
	default:
		return []Household{}, nil
	}
}

func (s *Service) Logs(ctx context.Context, current users.User, householdID uuid.UUID, limit, offset int) ([]ChangeLog, error) {
	household, err := s.Get(ctx, current, householdID)
	if err != nil {
		return nil, err
	}
	if household.ID == uuid.Nil {
		return nil, ErrHouseholdNotFound
	}
	if err := validatePagination(limit, offset); err != nil {
		return nil, err
	}
	return s.store.ListChangeLogs(ctx, householdID, limit, offset)
}

func (s *Service) AssignResponsibleByNumericRange(ctx context.Context, streetID, responsibleUserID uuid.UUID, from, to int, changedBy *uuid.UUID) error {
	if s.store == nil {
		return ErrInternalStoreNotReady
	}
	return s.store.AssignResponsibleByNumericRange(ctx, streetID, responsibleUserID, from, to, changedBy)
}

func (s *Service) ensureCanAccessHousehold(ctx context.Context, current users.User, household Household) error {
	if current.Role == users.RoleResponsiblePerson {
		if household.AssignedResponsibleUserID != nil && *household.AssignedResponsibleUserID == current.ID {
			return nil
		}
		return ErrForbidden
	}
	if _, err := s.streets.Get(ctx, current, household.StreetID); err != nil {
		return translateStreetError(err)
	}
	return nil
}

func (s *Service) writeUpdateLogs(ctx context.Context, userID uuid.UUID, old, next Household) {
	changedBy := &userID
	fields := []struct {
		name string
		old  *string
		next *string
	}{
		{"house_number", str(old.HouseNumber), str(next.HouseNumber)},
		{"total_numbers", str(fmt.Sprint(old.TotalNumbers)), str(fmt.Sprint(next.TotalNumbers))},
		{"contacted_numbers", str(fmt.Sprint(old.ContactedNumbers)), str(fmt.Sprint(next.ContactedNumbers))},
		{"voted_numbers", str(fmt.Sprint(old.VotedNumbers)), str(fmt.Sprint(next.VotedNumbers))},
		{"status", str(string(old.Status)), str(string(next.Status))},
		{"notes", old.Notes, next.Notes},
	}
	for _, field := range fields {
		if ptrValue(field.old) == ptrValue(field.next) {
			continue
		}
		_, _ = s.store.CreateChangeLog(ctx, CreateChangeLogInput{
			HouseholdID:     next.ID,
			ChangedByUserID: changedBy,
			FieldName:       field.name,
			OldValue:        field.old,
			NewValue:        field.next,
		})
	}
}

func normalizeCreate(input CreateHouseholdInput) CreateHouseholdInput {
	input.HouseNumber = strings.TrimSpace(input.HouseNumber)
	input.Status = normalizeStatus(input.Status, input.TotalNumbers, input.VotedNumbers)
	return input
}

func normalizeUpdate(input UpdateHouseholdInput) UpdateHouseholdInput {
	input.HouseNumber = strings.TrimSpace(input.HouseNumber)
	input.Status = normalizeStatus(input.Status, input.TotalNumbers, input.VotedNumbers)
	return input
}

func normalizeStatus(status Status, total, voted int32) Status {
	if status != "" {
		return status
	}
	if total > 0 && voted == total {
		return StatusFullyVoted
	}
	if voted > 0 && voted < total {
		return StatusPartiallyVoted
	}
	return StatusNew
}

func validateCreate(input CreateHouseholdInput) error {
	return validateFields(input.HouseNumber, input.TotalNumbers, input.ContactedNumbers, input.VotedNumbers, input.Status)
}

func validateUpdate(input UpdateHouseholdInput) error {
	return validateFields(input.HouseNumber, input.TotalNumbers, input.ContactedNumbers, input.VotedNumbers, input.Status)
}

func validateFields(houseNumber string, total, contacted, voted int32, status Status) error {
	if strings.TrimSpace(houseNumber) == "" {
		return ErrHouseNumberRequired
	}
	if total < 0 || contacted < 0 || voted < 0 || contacted > total || voted > total {
		return ErrInvalidCounts
	}
	if !IsValidStatus(status) {
		return ErrInvalidStatus
	}
	return nil
}

func IsValidStatus(status Status) bool {
	switch status {
	case StatusNew, StatusVisited, StatusExplained, StatusPartiallyVoted, StatusFullyVoted, StatusNotHome, StatusCallbackNeeded, StatusRefused, StatusInvalidInfo:
		return true
	default:
		return false
	}
}

func validatePagination(limit, offset int) error {
	if limit < 1 || limit > 200 || offset < 0 {
		return ErrInvalidPagination
	}
	return nil
}

func parseNumericHouseNumber(value string) (int, bool) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	return parsed, err == nil
}

func uuidString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	return str(id.String())
}

func str(value string) *string {
	return &value
}

func ptrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func translateStreetError(err error) error {
	if err == streets.ErrForbidden {
		return ErrForbidden
	}
	return err
}
