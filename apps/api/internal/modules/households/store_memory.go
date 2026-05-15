package households

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Create(ctx context.Context, input CreateHouseholdInput) (Household, error)
	GetByID(ctx context.Context, id uuid.UUID) (Household, error)
	ListByStreetID(ctx context.Context, streetID uuid.UUID, limit, offset int) ([]Household, error)
	ListByResponsibleUserID(ctx context.Context, responsibleUserID uuid.UUID, limit, offset int) ([]Household, error)
	ListByMFYID(ctx context.Context, mfyID uuid.UUID, limit, offset int) ([]Household, error)
	ListByStreetIDs(ctx context.Context, streetIDs []uuid.UUID, limit, offset int) ([]Household, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateHouseholdInput) (Household, error)
	UpdateAssignment(ctx context.Context, id uuid.UUID, responsibleUserID *uuid.UUID) (Household, error)
	CreateChangeLog(ctx context.Context, input CreateChangeLogInput) (ChangeLog, error)
	ListChangeLogs(ctx context.Context, householdID uuid.UUID, limit, offset int) ([]ChangeLog, error)
	AssignResponsibleByNumericRange(ctx context.Context, streetID, responsibleUserID uuid.UUID, from, to int, changedBy *uuid.UUID) error
}

type MemoryStore struct {
	mu         sync.RWMutex
	items      map[uuid.UUID]Household
	changeLogs map[uuid.UUID]ChangeLog
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[uuid.UUID]Household), changeLogs: make(map[uuid.UUID]ChangeLog)}
}

func (s *MemoryStore) Create(ctx context.Context, input CreateHouseholdInput) (Household, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.items {
		if item.StreetID == input.StreetID && strings.EqualFold(item.HouseNumber, input.HouseNumber) {
			return Household{}, ErrDuplicateHousehold
		}
	}
	now := time.Now().UTC()
	item := Household{
		ID:                        uuid.New(),
		MFYID:                     input.MFYID,
		StreetID:                  input.StreetID,
		HouseNumber:               input.HouseNumber,
		TotalNumbers:              input.TotalNumbers,
		ContactedNumbers:          input.ContactedNumbers,
		VotedNumbers:              input.VotedNumbers,
		Status:                    input.Status,
		Notes:                     input.Notes,
		AssignedResponsibleUserID: input.AssignedResponsibleUserID,
		CreatedByUserID:           input.CreatedByUserID,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
	s.items[item.ID] = item
	return item, nil
}

func (s *MemoryStore) GetByID(ctx context.Context, id uuid.UUID) (Household, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return Household{}, ErrHouseholdNotFound
	}
	return item, nil
}

func (s *MemoryStore) ListByStreetID(ctx context.Context, streetID uuid.UUID, limit, offset int) ([]Household, error) {
	return s.filter(limit, offset, func(item Household) bool { return item.StreetID == streetID }), nil
}

func (s *MemoryStore) ListByResponsibleUserID(ctx context.Context, responsibleUserID uuid.UUID, limit, offset int) ([]Household, error) {
	return s.filter(limit, offset, func(item Household) bool {
		return item.AssignedResponsibleUserID != nil && *item.AssignedResponsibleUserID == responsibleUserID
	}), nil
}

func (s *MemoryStore) ListByMFYID(ctx context.Context, mfyID uuid.UUID, limit, offset int) ([]Household, error) {
	return s.filter(limit, offset, func(item Household) bool { return item.MFYID == mfyID }), nil
}

func (s *MemoryStore) ListByStreetIDs(ctx context.Context, streetIDs []uuid.UUID, limit, offset int) ([]Household, error) {
	set := make(map[uuid.UUID]struct{}, len(streetIDs))
	for _, id := range streetIDs {
		set[id] = struct{}{}
	}
	return s.filter(limit, offset, func(item Household) bool {
		_, ok := set[item.StreetID]
		return ok
	}), nil
}

func (s *MemoryStore) Update(ctx context.Context, id uuid.UUID, input UpdateHouseholdInput) (Household, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return Household{}, ErrHouseholdNotFound
	}
	for _, other := range s.items {
		if other.ID != id && other.StreetID == item.StreetID && strings.EqualFold(other.HouseNumber, input.HouseNumber) {
			return Household{}, ErrDuplicateHousehold
		}
	}
	item.HouseNumber = input.HouseNumber
	item.TotalNumbers = input.TotalNumbers
	item.ContactedNumbers = input.ContactedNumbers
	item.VotedNumbers = input.VotedNumbers
	item.Status = input.Status
	item.Notes = input.Notes
	item.UpdatedAt = time.Now().UTC()
	s.items[id] = item
	return item, nil
}

func (s *MemoryStore) UpdateAssignment(ctx context.Context, id uuid.UUID, responsibleUserID *uuid.UUID) (Household, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return Household{}, ErrHouseholdNotFound
	}
	item.AssignedResponsibleUserID = responsibleUserID
	item.UpdatedAt = time.Now().UTC()
	s.items[id] = item
	return item, nil
}

func (s *MemoryStore) CreateChangeLog(ctx context.Context, input CreateChangeLogInput) (ChangeLog, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	log := ChangeLog{ID: uuid.New(), HouseholdID: input.HouseholdID, ChangedByUserID: input.ChangedByUserID, FieldName: input.FieldName, OldValue: input.OldValue, NewValue: input.NewValue, Note: input.Note, CreatedAt: now}
	s.changeLogs[log.ID] = log
	return log, nil
}

func (s *MemoryStore) ListChangeLogs(ctx context.Context, householdID uuid.UUID, limit, offset int) ([]ChangeLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]ChangeLog, 0)
	for _, log := range s.changeLogs {
		if log.HouseholdID == householdID {
			list = append(list, log)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
	if offset >= len(list) {
		return []ChangeLog{}, nil
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end], nil
}

func (s *MemoryStore) AssignResponsibleByNumericRange(ctx context.Context, streetID, responsibleUserID uuid.UUID, from, to int, changedBy *uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, item := range s.items {
		number, ok := parseNumericHouseNumber(item.HouseNumber)
		if item.StreetID == streetID && ok && number >= from && number <= to {
			old := uuidString(item.AssignedResponsibleUserID)
			item.AssignedResponsibleUserID = &responsibleUserID
			item.UpdatedAt = time.Now().UTC()
			s.items[id] = item
			newValue := responsibleUserID.String()
			log := ChangeLog{ID: uuid.New(), HouseholdID: item.ID, ChangedByUserID: changedBy, FieldName: "assigned_responsible_user_id", OldValue: old, NewValue: &newValue, CreatedAt: time.Now().UTC()}
			s.changeLogs[log.ID] = log
		}
	}
	return nil
}

func (s *MemoryStore) filter(limit, offset int, fn func(Household) bool) []Household {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]Household, 0)
	for _, item := range s.items {
		if fn(item) {
			list = append(list, item)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
	if offset >= len(list) {
		return []Household{}
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end]
}
