package zephyrix

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type MemorySessionStorage struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewMemorySessionStorage() *MemorySessionStorage {
	return &MemorySessionStorage{
		sessions: make(map[string]*Session),
	}
}

func (m *MemorySessionStorage) Create(ctx context.Context, session *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	return nil
}

func (m *MemorySessionStorage) Get(ctx context.Context, sessionID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}

func (m *MemorySessionStorage) Update(ctx context.Context, session *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	return nil
}

func (m *MemorySessionStorage) Delete(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
	return nil
}

func (m *MemorySessionStorage) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for id, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			delete(m.sessions, id)
		}
	}
	return nil
}
