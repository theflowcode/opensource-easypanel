package mock

import (
	"context"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// --- Projects ---

func (m *MockDatabasePort) CreateProject(ctx context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateProject")

	if err := p.Validate(); err != nil {
		return err
	}
	if _, exists := m.Projects[p.ID]; exists {
		return domain.ErrAlreadyExists
	}
	for _, existing := range m.Projects {
		if existing.Name == p.Name {
			return domain.ErrAlreadyExists
		}
	}

	clone := *p
	now := time.Now().UTC()
	if clone.CreatedAt.IsZero() {
		clone.CreatedAt = now
	}
	if clone.UpdatedAt.IsZero() {
		clone.UpdatedAt = now
	}
	m.Projects[p.ID] = &clone
	return nil
}

func (m *MockDatabasePort) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.Projects[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	clone := *p
	return &clone, nil
}

func (m *MockDatabasePort) GetProjectByName(ctx context.Context, name string) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.Projects {
		if p.Name == name {
			clone := *p
			return &clone, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockDatabasePort) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*domain.Project
	for _, p := range m.Projects {
		clone := *p
		list = append(list, &clone)
	}
	return list, nil
}

func (m *MockDatabasePort) UpdateProject(ctx context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "UpdateProject")

	if _, ok := m.Projects[p.ID]; !ok {
		return domain.ErrNotFound
	}
	clone := *p
	clone.UpdatedAt = time.Now().UTC()
	m.Projects[p.ID] = &clone
	return nil
}

func (m *MockDatabasePort) DeleteProject(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteProject")

	if _, ok := m.Projects[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.Projects, id)

	// Cascade delete services
	for sID, s := range m.Services {
		if s.ProjectID == id {
			delete(m.Services, sID)
		}
	}
	return nil
}
