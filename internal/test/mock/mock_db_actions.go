package mock

import (
	"context"
	"sort"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (m *MockDatabasePort) CreateAction(ctx context.Context, action *domain.Action) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateAction")

	if err := action.Validate(); err != nil {
		return err
	}
	if _, exists := m.Actions[action.ID]; exists {
		return domain.ErrAlreadyExists
	}

	cp := *action
	now := time.Now().UTC()
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = now
	}
	if cp.UpdatedAt.IsZero() {
		cp.UpdatedAt = now
	}
	m.Actions[action.ID] = &cp
	return nil
}

func (m *MockDatabasePort) GetAction(ctx context.Context, id string) (*domain.Action, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "GetAction")

	action, exists := m.Actions[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	cp := *action
	return &cp, nil
}

func (m *MockDatabasePort) UpdateAction(ctx context.Context, action *domain.Action) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "UpdateAction")

	if err := action.Validate(); err != nil {
		return err
	}
	if _, exists := m.Actions[action.ID]; !exists {
		return domain.ErrNotFound
	}

	cp := *action
	cp.UpdatedAt = time.Now().UTC()
	m.Actions[action.ID] = &cp
	return nil
}

func (m *MockDatabasePort) ListActions(ctx context.Context, projectName, serviceName string, limit, offset int) ([]*domain.Action, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "ListActions")

	var result []*domain.Action
	for _, a := range m.Actions {
		if projectName != "" && a.ProjectName != projectName {
			continue
		}
		if serviceName != "" && a.ServiceName != serviceName {
			continue
		}
		cp := *a
		result = append(result, &cp)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	if offset > len(result) {
		return []*domain.Action{}, nil
	}
	result = result[offset:]
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
