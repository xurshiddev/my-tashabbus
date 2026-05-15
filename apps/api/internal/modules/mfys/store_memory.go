package mfys

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemoryStore struct {
	mu   sync.RWMutex
	mfys map[uuid.UUID]MFY
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{mfys: make(map[uuid.UUID]MFY)}
}

func (s *MemoryStore) Create(ctx context.Context, input CreateMFYInput) (MFY, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	mfy := MFY{
		ID:              uuid.New(),
		Name:            input.Name,
		Region:          input.Region,
		District:        input.District,
		TargetVotes:     input.TargetVotes,
		Season:          input.Season,
		IsActive:        true,
		CreatedByUserID: input.CreatedByUserID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	s.mfys[mfy.ID] = mfy
	return mfy, nil
}

func (s *MemoryStore) GetByID(ctx context.Context, id uuid.UUID) (MFY, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	mfy, ok := s.mfys[id]
	if !ok {
		return MFY{}, ErrMFYNotFound
	}
	return mfy, nil
}

func (s *MemoryStore) List(ctx context.Context, limit, offset int) ([]MFY, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]MFY, 0, len(s.mfys))
	for _, mfy := range s.mfys {
		list = append(list, mfy)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	if offset >= len(list) {
		return []MFY{}, nil
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end], nil
}

func (s *MemoryStore) Update(ctx context.Context, id uuid.UUID, input UpdateMFYInput) (MFY, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	mfy, ok := s.mfys[id]
	if !ok {
		return MFY{}, ErrMFYNotFound
	}
	mfy.Name = input.Name
	mfy.Region = input.Region
	mfy.District = input.District
	mfy.TargetVotes = input.TargetVotes
	mfy.Season = input.Season
	mfy.IsActive = input.IsActive
	mfy.UpdatedAt = time.Now().UTC()
	s.mfys[id] = mfy
	return mfy, nil
}

func (s *MemoryStore) SetActive(ctx context.Context, id uuid.UUID, active bool) (MFY, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	mfy, ok := s.mfys[id]
	if !ok {
		return MFY{}, ErrMFYNotFound
	}
	mfy.IsActive = active
	mfy.UpdatedAt = time.Now().UTC()
	s.mfys[id] = mfy
	return mfy, nil
}
