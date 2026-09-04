package mock

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (m *MockDatabasePort) CreateMiddleware(ctx context.Context, mw *domain.Middleware) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateMiddleware")

	if err := mw.Validate(); err != nil {
		return err
	}
	if _, exists := m.Middlewares[mw.ID]; exists {
		return domain.ErrAlreadyExists
	}
	for _, existing := range m.Middlewares {
		if existing.Name == mw.Name {
			return domain.ErrAlreadyExists
		}
	}
	cp := *mw
	m.Middlewares[mw.ID] = &cp
	return nil
}

func (m *MockDatabasePort) GetMiddleware(ctx context.Context, id string) (*domain.Middleware, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "GetMiddleware")

	mw, exists := m.Middlewares[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	cp := *mw
	return &cp, nil
}

func (m *MockDatabasePort) ListMiddlewares(ctx context.Context) ([]*domain.Middleware, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "ListMiddlewares")

	var list []*domain.Middleware
	for _, mw := range m.Middlewares {
		cp := *mw
		list = append(list, &cp)
	}
	return list, nil
}

func (m *MockDatabasePort) UpdateMiddleware(ctx context.Context, mw *domain.Middleware) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "UpdateMiddleware")

	if err := mw.Validate(); err != nil {
		return err
	}
	if _, exists := m.Middlewares[mw.ID]; !exists {
		return domain.ErrNotFound
	}
	cp := *mw
	m.Middlewares[mw.ID] = &cp
	return nil
}

func (m *MockDatabasePort) DeleteMiddleware(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteMiddleware")

	if _, exists := m.Middlewares[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.Middlewares, id)
	return nil
}
