package mock

import (
	"context"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// --- Services ---

func (m *MockDatabasePort) CreateService(ctx context.Context, s *domain.Service) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateService")

	if err := s.Validate(); err != nil {
		return err
	}
	if _, ok := m.Projects[s.ProjectID]; !ok {
		return domain.ErrNotFound
	}
	if _, ok := m.Services[s.ID]; ok {
		return domain.ErrAlreadyExists
	}
	for _, existing := range m.Services {
		if existing.ProjectID == s.ProjectID && existing.Name == s.Name {
			return domain.ErrAlreadyExists
		}
	}

	clone := *s
	now := time.Now().UTC()
	if clone.CreatedAt.IsZero() {
		clone.CreatedAt = now
	}
	if clone.UpdatedAt.IsZero() {
		clone.UpdatedAt = now
	}
	m.Services[s.ID] = &clone
	return nil
}

func (m *MockDatabasePort) GetService(ctx context.Context, id string) (*domain.Service, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.Services[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	clone := *s
	return &clone, nil
}

func (m *MockDatabasePort) GetServiceByName(ctx context.Context, projectID, name string) (*domain.Service, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.Services {
		if s.ProjectID == projectID && s.Name == name {
			clone := *s
			return &clone, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockDatabasePort) GetServiceByDeployToken(ctx context.Context, token string) (*domain.Service, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.Services {
		if s.DeployToken == token && token != "" {
			clone := *s
			return &clone, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockDatabasePort) ListServicesByProject(ctx context.Context, projectID string) ([]*domain.Service, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*domain.Service
	for _, s := range m.Services {
		if s.ProjectID == projectID {
			clone := *s
			list = append(list, &clone)
		}
	}
	return list, nil
}

func (m *MockDatabasePort) ListAllServices(ctx context.Context) ([]*domain.Service, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*domain.Service
	for _, s := range m.Services {
		clone := *s
		list = append(list, &clone)
	}
	return list, nil
}

func (m *MockDatabasePort) UpdateService(ctx context.Context, s *domain.Service) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "UpdateService")

	if _, ok := m.Services[s.ID]; !ok {
		return domain.ErrNotFound
	}
	clone := *s
	clone.UpdatedAt = time.Now().UTC()
	m.Services[s.ID] = &clone
	return nil
}

func (m *MockDatabasePort) DeleteService(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteService")

	if _, ok := m.Services[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.Services, id)

	// Cascade delete domains and deployments
	for dID, d := range m.Domains {
		if d.ServiceID == id {
			delete(m.Domains, dID)
		}
	}
	for depID, dep := range m.Deployments {
		if dep.ServiceID == id {
			delete(m.Deployments, depID)
		}
	}
	return nil
}
