package mock

import (
	"context"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (m *MockDatabasePort) CreateSession(ctx context.Context, session *domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateSession")

	if err := session.Validate(); err != nil {
		return err
	}
	if _, exists := m.Sessions[session.ID]; exists {
		return domain.ErrAlreadyExists
	}
	for _, s := range m.Sessions {
		if s.TokenHash == session.TokenHash {
			return domain.ErrAlreadyExists
		}
	}

	clone := *session
	m.Sessions[session.ID] = &clone
	return nil
}

func (m *MockDatabasePort) GetSession(ctx context.Context, tokenHash string) (*domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "GetSession")

	for _, s := range m.Sessions {
		if s.TokenHash == tokenHash {
			clone := *s
			return &clone, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockDatabasePort) DeleteSession(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteSession")

	if _, exists := m.Sessions[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.Sessions, id)
	return nil
}

func (m *MockDatabasePort) DeleteExpiredSessions(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteExpiredSessions")

	now := time.Now().UTC()
	for id, s := range m.Sessions {
		if now.After(s.ExpiresAt) {
			delete(m.Sessions, id)
		}
	}
	return nil
}
