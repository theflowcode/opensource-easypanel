package mock

import (
	"context"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// --- Users & Auth ---

func (m *MockDatabasePort) CreateUser(ctx context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateUser")

	if err := u.Validate(); err != nil {
		return err
	}
	if _, ok := m.Users[u.ID]; ok {
		return domain.ErrAlreadyExists
	}
	for _, existing := range m.Users {
		if existing.Email == u.Email {
			return domain.ErrAlreadyExists
		}
	}

	clone := *u
	now := time.Now().UTC()
	if clone.CreatedAt.IsZero() {
		clone.CreatedAt = now
	}
	if clone.UpdatedAt.IsZero() {
		clone.UpdatedAt = now
	}
	m.Users[u.ID] = &clone
	return nil
}

func (m *MockDatabasePort) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, ok := m.Users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	clone := *u
	return &clone, nil
}

func (m *MockDatabasePort) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.Users {
		if u.Email == email {
			clone := *u
			return &clone, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockDatabasePort) ListUsers(ctx context.Context) ([]*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*domain.User
	for _, u := range m.Users {
		clone := *u
		list = append(list, &clone)
	}
	return list, nil
}

func (m *MockDatabasePort) UpdateUser(ctx context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "UpdateUser")

	if _, ok := m.Users[u.ID]; !ok {
		return domain.ErrNotFound
	}
	clone := *u
	clone.UpdatedAt = time.Now().UTC()
	m.Users[u.ID] = &clone
	return nil
}

func (m *MockDatabasePort) DeleteUser(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteUser")

	if _, ok := m.Users[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.Users, id)
	return nil
}
