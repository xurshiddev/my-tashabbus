package streets

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Create(ctx context.Context, input CreateStreetInput) (Street, error)
	GetByID(ctx context.Context, id uuid.UUID) (Street, error)
	ListByMFYID(ctx context.Context, mfyID uuid.UUID, limit, offset int) ([]Street, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateStreetInput) (Street, error)
	ListForStreetLeader(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Street, error)
	IsUserAssignedToStreet(ctx context.Context, streetID, userID uuid.UUID) (bool, error)
	ReassignLeader(ctx context.Context, streetID, userID uuid.UUID, assignedBy *uuid.UUID) (StreetLeaderAssignment, error)
	GetActiveLeader(ctx context.Context, streetID uuid.UUID) (*StreetLeaderAssignment, error)
}

type MemoryStore struct {
	mu          sync.RWMutex
	streets     map[uuid.UUID]Street
	assignments map[uuid.UUID]StreetLeaderAssignment
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		streets:     make(map[uuid.UUID]Street),
		assignments: make(map[uuid.UUID]StreetLeaderAssignment),
	}
}

func (s *MemoryStore) Create(ctx context.Context, input CreateStreetInput) (Street, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, street := range s.streets {
		if street.MFYID == input.MFYID && street.IsActive && strings.EqualFold(street.Name, input.Name) {
			return Street{}, ErrDuplicateStreetName
		}
	}
	now := time.Now().UTC()
	street := Street{
		ID:                     uuid.New(),
		MFYID:                  input.MFYID,
		Name:                   input.Name,
		PlannedHouseholdsCount: input.PlannedHouseholdsCount,
		Notes:                  input.Notes,
		IsActive:               true,
		CreatedByUserID:        input.CreatedByUserID,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	s.streets[street.ID] = street
	return street, nil
}

func (s *MemoryStore) GetByID(ctx context.Context, id uuid.UUID) (Street, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	street, ok := s.streets[id]
	if !ok {
		return Street{}, ErrStreetNotFound
	}
	return street, nil
}

func (s *MemoryStore) ListByMFYID(ctx context.Context, mfyID uuid.UUID, limit, offset int) ([]Street, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]Street, 0)
	for _, street := range s.streets {
		if street.MFYID == mfyID {
			list = append(list, street)
		}
	}
	return paginateStreets(list, limit, offset), nil
}

func (s *MemoryStore) Update(ctx context.Context, id uuid.UUID, input UpdateStreetInput) (Street, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	street, ok := s.streets[id]
	if !ok {
		return Street{}, ErrStreetNotFound
	}
	for _, other := range s.streets {
		if other.ID != id && other.MFYID == street.MFYID && other.IsActive && input.IsActive && strings.EqualFold(other.Name, input.Name) {
			return Street{}, ErrDuplicateStreetName
		}
	}
	street.Name = input.Name
	street.PlannedHouseholdsCount = input.PlannedHouseholdsCount
	street.Notes = input.Notes
	street.IsActive = input.IsActive
	street.UpdatedAt = time.Now().UTC()
	s.streets[id] = street
	return street, nil
}

func (s *MemoryStore) ListForStreetLeader(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Street, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]Street, 0)
	for _, assignment := range s.assignments {
		if assignment.UserID == userID && assignment.IsActive {
			if street, ok := s.streets[assignment.StreetID]; ok && street.IsActive {
				list = append(list, street)
			}
		}
	}
	return paginateStreets(list, limit, offset), nil
}

func (s *MemoryStore) IsUserAssignedToStreet(ctx context.Context, streetID, userID uuid.UUID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, assignment := range s.assignments {
		if assignment.StreetID == streetID && assignment.UserID == userID && assignment.IsActive {
			return true, nil
		}
	}
	return false, nil
}

func (s *MemoryStore) ReassignLeader(ctx context.Context, streetID, userID uuid.UUID, assignedBy *uuid.UUID) (StreetLeaderAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.streets[streetID]; !ok {
		return StreetLeaderAssignment{}, ErrStreetNotFound
	}
	now := time.Now().UTC()
	for id, assignment := range s.assignments {
		if assignment.StreetID == streetID && assignment.IsActive {
			assignment.IsActive = false
			assignment.UpdatedAt = now
			s.assignments[id] = assignment
		}
	}
	assignment := StreetLeaderAssignment{
		ID:               uuid.New(),
		StreetID:         streetID,
		UserID:           userID,
		AssignedByUserID: assignedBy,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.assignments[assignment.ID] = assignment
	return assignment, nil
}

func (s *MemoryStore) GetActiveLeader(ctx context.Context, streetID uuid.UUID) (*StreetLeaderAssignment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, assignment := range s.assignments {
		if assignment.StreetID == streetID && assignment.IsActive {
			copy := assignment
			return &copy, nil
		}
	}
	return nil, nil
}

func paginateStreets(list []Street, limit, offset int) []Street {
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	if offset >= len(list) {
		return []Street{}
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end]
}
