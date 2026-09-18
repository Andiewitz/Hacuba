package users

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// MemoryStore is an in-memory Store for unit tests. Production uses
// PostgresStore — this exists so handler tests never need a database.
type MemoryStore struct {
	mu       sync.RWMutex
	byID     map[uuid.UUID]User
	byEmail  map[string]uuid.UUID
	sessions map[string]RefreshSession // keyed by token_hash
}

// NewMemoryStore returns an empty test store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		byID:     make(map[uuid.UUID]User),
		byEmail:  make(map[string]uuid.UUID),
		sessions: make(map[string]RefreshSession),
	}
}

func (m *MemoryStore) CreateUser(_ context.Context, user User) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.byEmail[user.Email]; exists {
		return nil, ErrEmailTaken
	}
	m.byID[user.ID] = user
	m.byEmail[user.Email] = user.ID
	u := user
	return &u, nil
}

func (m *MemoryStore) GetUserByEmail(_ context.Context, email string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.byEmail[NormalizeEmail(email)]
	if !ok {
		return nil, ErrNotFound
	}
	u := m.byID[id]
	return &u, nil
}

func (m *MemoryStore) GetUserByID(_ context.Context, id uuid.UUID) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &u, nil
}

func (m *MemoryStore) CreateRefreshSession(_ context.Context, s RefreshSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.TokenHash] = s
	return nil
}

func (m *MemoryStore) GetRefreshSessionByHash(_ context.Context, tokenHash string) (*RefreshSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[tokenHash]
	if !ok {
		return nil, ErrNotFound
	}
	return &s, nil
}

func (m *MemoryStore) DeleteRefreshSessionByHash(_ context.Context, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, tokenHash)
	return nil
}

func (m *MemoryStore) DeleteUserSessions(_ context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for h, s := range m.sessions {
		if s.UserID == userID {
			delete(m.sessions, h)
		}
	}
	return nil
}
