package mock

import (
	"context"
	"sort"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (m *MockDatabasePort) CreateStorageProvider(ctx context.Context, sp *domain.StorageProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateStorageProvider")

	if err := sp.Validate(); err != nil {
		return err
	}
	if _, exists := m.StorageProviders[sp.ID]; exists {
		return domain.ErrAlreadyExists
	}

	cp := *sp
	now := time.Now().UTC()
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = now
	}
	if cp.UpdatedAt.IsZero() {
		cp.UpdatedAt = now
	}
	m.StorageProviders[sp.ID] = &cp
	return nil
}

func (m *MockDatabasePort) GetStorageProvider(ctx context.Context, id string) (*domain.StorageProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "GetStorageProvider")

	sp, exists := m.StorageProviders[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	cp := *sp
	return &cp, nil
}

func (m *MockDatabasePort) ListStorageProviders(ctx context.Context) ([]*domain.StorageProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "ListStorageProviders")

	var result []*domain.StorageProvider
	for _, sp := range m.StorageProviders {
		cp := *sp
		result = append(result, &cp)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (m *MockDatabasePort) DeleteStorageProvider(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteStorageProvider")

	if _, exists := m.StorageProviders[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.StorageProviders, id)
	return nil
}
