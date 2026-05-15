package users

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemoryStore struct {
	mu    sync.RWMutex
	users map[uuid.UUID]User
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{users: make(map[uuid.UUID]User)}
}

func (s *MemoryStore) Create(ctx context.Context, input CreateUserInput) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if input.TelegramID != nil {
		for _, user := range s.users {
			if user.TelegramID != nil && *user.TelegramID == *input.TelegramID {
				return User{}, ErrTelegramIDConflict
			}
		}
	}
	now := time.Now().UTC()
	user := User{
		ID:               uuid.New(),
		FullName:         input.FullName,
		Phone:            input.Phone,
		TelegramID:       input.TelegramID,
		TelegramUsername: input.TelegramUsername,
		Role:             input.Role,
		MFYID:            input.MFYID,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.users[user.ID] = user
	return user, nil
}

func (s *MemoryStore) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *MemoryStore) GetByTelegramID(ctx context.Context, telegramID int64) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.users {
		if user.TelegramID != nil && *user.TelegramID == telegramID {
			return user, nil
		}
	}
	return User{}, ErrUserNotFound
}

func (s *MemoryStore) List(ctx context.Context, limit, offset int) ([]User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]User, 0, len(s.users))
	for _, user := range s.users {
		list = append(list, user)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	if offset >= len(list) {
		return []User{}, nil
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end], nil
}

func (s *MemoryStore) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	user.FullName = input.FullName
	user.Phone = input.Phone
	user.Role = input.Role
	user.MFYID = input.MFYID
	user.IsActive = input.IsActive
	user.UpdatedAt = time.Now().UTC()
	s.users[id] = user
	return user, nil
}

func (s *MemoryStore) SetTelegramIdentity(ctx context.Context, id uuid.UUID, input SetTelegramIdentityInput) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for userID, user := range s.users {
		if userID != id && user.TelegramID != nil && *user.TelegramID == input.TelegramID {
			return User{}, ErrTelegramIDConflict
		}
	}
	user, ok := s.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	user.TelegramID = &input.TelegramID
	user.TelegramUsername = input.TelegramUsername
	user.UpdatedAt = time.Now().UTC()
	s.users[id] = user
	return user, nil
}

func (s *MemoryStore) AssignToMFY(ctx context.Context, id uuid.UUID, mfyID uuid.UUID) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	user.MFYID = &mfyID
	user.UpdatedAt = time.Now().UTC()
	s.users[id] = user
	return user, nil
}

func (s *MemoryStore) Deactivate(ctx context.Context, id uuid.UUID) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	user.IsActive = false
	user.UpdatedAt = time.Now().UTC()
	s.users[id] = user
	return user, nil
}

func (s *MemoryStore) Activate(ctx context.Context, id uuid.UUID) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	user.IsActive = true
	user.UpdatedAt = time.Now().UTC()
	s.users[id] = user
	return user, nil
}
