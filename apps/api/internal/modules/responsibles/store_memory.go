package responsibles

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Create(ctx context.Context, input CreateAssignmentInput) (Assignment, error)
	GetByID(ctx context.Context, id uuid.UUID) (Assignment, error)
	ListByStreetID(ctx context.Context, streetID uuid.UUID, limit, offset int, activeOnly bool) ([]Assignment, error)
	ListByResponsibleUserID(ctx context.Context, responsibleUserID uuid.UUID, limit, offset int) ([]Assignment, error)
	Deactivate(ctx context.Context, id uuid.UUID) (Assignment, error)
	IsResponsibleAssignedToStreet(ctx context.Context, streetID, responsibleUserID uuid.UUID) (bool, error)
}

type MemoryStore struct {
	mu    sync.RWMutex
	items map[uuid.UUID]Assignment
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[uuid.UUID]Assignment)}
}

func (s *MemoryStore) Create(ctx context.Context, input CreateAssignmentInput) (Assignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	item := Assignment{ID: uuid.New(), StreetID: input.StreetID, ResponsibleUserID: input.ResponsibleUserID, AssignedByUserID: input.AssignedByUserID, FromHouseNumber: input.FromHouseNumber, ToHouseNumber: input.ToHouseNumber, IsActive: true, CreatedAt: now, UpdatedAt: now}
	s.items[item.ID] = item
	return item, nil
}

func (s *MemoryStore) GetByID(ctx context.Context, id uuid.UUID) (Assignment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return Assignment{}, ErrAssignmentNotFound
	}
	return item, nil
}

func (s *MemoryStore) ListByStreetID(ctx context.Context, streetID uuid.UUID, limit, offset int, activeOnly bool) ([]Assignment, error) {
	return s.filter(limit, offset, func(item Assignment) bool { return item.StreetID == streetID && (!activeOnly || item.IsActive) }), nil
}

func (s *MemoryStore) ListByResponsibleUserID(ctx context.Context, responsibleUserID uuid.UUID, limit, offset int) ([]Assignment, error) {
	return s.filter(limit, offset, func(item Assignment) bool { return item.ResponsibleUserID == responsibleUserID }), nil
}

func (s *MemoryStore) Deactivate(ctx context.Context, id uuid.UUID) (Assignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return Assignment{}, ErrAssignmentNotFound
	}
	item.IsActive = false
	item.UpdatedAt = time.Now().UTC()
	s.items[id] = item
	return item, nil
}

func (s *MemoryStore) IsResponsibleAssignedToStreet(ctx context.Context, streetID, responsibleUserID uuid.UUID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.StreetID == streetID && item.ResponsibleUserID == responsibleUserID && item.IsActive {
			return true, nil
		}
	}
	return false, nil
}

func (s *MemoryStore) filter(limit, offset int, fn func(Assignment) bool) []Assignment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]Assignment, 0)
	for _, item := range s.items {
		if fn(item) {
			list = append(list, item)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].CreatedAt.Equal(list[j].CreatedAt) {
			return strings.Compare(list[i].FromHouseNumber, list[j].FromHouseNumber) < 0
		}
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	if offset >= len(list) {
		return []Assignment{}
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end]
}
