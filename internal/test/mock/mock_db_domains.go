package mock

import (
	"context"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// --- Domains ---

func (m *MockDatabasePort) CreateDomain(ctx context.Context, d *domain.Domain) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateDomain")

	if err := d.Validate(); err != nil {
		return err
	}
	if _, ok := m.Services[d.ServiceID]; !ok {
		return domain.ErrNotFound
	}
	if _, ok := m.Domains[d.ID]; ok {
		return domain.ErrAlreadyExists
	}
	for _, existing := range m.Domains {
		if existing.DomainName == d.DomainName {
			return domain.ErrAlreadyExists
		}
	}

	clone := *d
	now := time.Now().UTC()
	if clone.CreatedAt.IsZero() {
		clone.CreatedAt = now
	}
	if clone.UpdatedAt.IsZero() {
		clone.UpdatedAt = now
	}
	m.Domains[d.ID] = &clone
	return nil
}

func (m *MockDatabasePort) GetDomain(ctx context.Context, id string) (*domain.Domain, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	d, ok := m.Domains[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	clone := *d
	return &clone, nil
}

func (m *MockDatabasePort) ListDomainsByService(ctx context.Context, serviceID string) ([]*domain.Domain, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*domain.Domain
	for _, d := range m.Domains {
		if d.ServiceID == serviceID {
			clone := *d
			list = append(list, &clone)
		}
	}
	return list, nil
}

func (m *MockDatabasePort) ListAllDomains(ctx context.Context) ([]*domain.Domain, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*domain.Domain
	for _, d := range m.Domains {
		clone := *d
		list = append(list, &clone)
	}
	return list, nil
}

func (m *MockDatabasePort) UpdateDomain(ctx context.Context, d *domain.Domain) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "UpdateDomain")

	if err := d.Validate(); err != nil {
		return err
	}
	existing, ok := m.Domains[d.ID]
	if !ok {
		return domain.ErrNotFound
	}
	for id, other := range m.Domains {
		if id != d.ID && other.DomainName == d.DomainName {
			return domain.ErrAlreadyExists
		}
	}
	clone := *d
	clone.CreatedAt = existing.CreatedAt
	clone.UpdatedAt = time.Now().UTC()
	m.Domains[d.ID] = &clone
	return nil
}

func (m *MockDatabasePort) DeleteDomain(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteDomain")

	if _, ok := m.Domains[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.Domains, id)
	return nil
}
