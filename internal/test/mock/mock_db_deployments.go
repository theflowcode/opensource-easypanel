package mock

import (
	"context"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// --- Deployments ---

func (m *MockDatabasePort) CreateDeployment(ctx context.Context, d *domain.Deployment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateDeployment")

	if err := d.Validate(); err != nil {
		return err
	}
	if _, ok := m.Services[d.ServiceID]; !ok {
		return domain.ErrNotFound
	}
	if _, ok := m.Deployments[d.ID]; ok {
		return domain.ErrAlreadyExists
	}

	clone := *d
	if clone.StartedAt.IsZero() {
		clone.StartedAt = time.Now().UTC()
	}
	m.Deployments[d.ID] = &clone
	return nil
}

func (m *MockDatabasePort) GetDeployment(ctx context.Context, id string) (*domain.Deployment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	d, ok := m.Deployments[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	clone := *d
	return &clone, nil
}

func (m *MockDatabasePort) ListDeploymentsByService(ctx context.Context, serviceID string, limit, offset int) ([]*domain.Deployment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []*domain.Deployment
	for _, d := range m.Deployments {
		if d.ServiceID == serviceID {
			clone := *d
			all = append(all, &clone)
		}
	}

	if offset >= len(all) {
		return []*domain.Deployment{}, nil
	}
	end := len(all)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return all[offset:end], nil
}

func (m *MockDatabasePort) UpdateDeployment(ctx context.Context, d *domain.Deployment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "UpdateDeployment")

	if _, ok := m.Deployments[d.ID]; !ok {
		return domain.ErrNotFound
	}
	clone := *d
	m.Deployments[d.ID] = &clone
	return nil
}
