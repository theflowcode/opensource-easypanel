package mock

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (m *MockDatabasePort) CreateBackup(ctx context.Context, backup *domain.Backup) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateBackup")

	if err := backup.Validate(); err != nil {
		return err
	}
	if _, exists := m.Backups[backup.ID]; exists {
		return domain.ErrAlreadyExists
	}

	clone := *backup
	m.Backups[backup.ID] = &clone
	return nil
}

func (m *MockDatabasePort) GetBackup(ctx context.Context, id string) (*domain.Backup, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "GetBackup")

	b, exists := m.Backups[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	clone := *b
	return &clone, nil
}

func (m *MockDatabasePort) ListBackupsByService(ctx context.Context, serviceID string, limit, offset int) ([]*domain.Backup, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "ListBackupsByService")

	var all []*domain.Backup
	for _, b := range m.Backups {
		if b.ServiceID == serviceID {
			clone := *b
			all = append(all, &clone)
		}
	}

	if offset >= len(all) {
		return []*domain.Backup{}, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (m *MockDatabasePort) DeleteBackup(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteBackup")

	if _, exists := m.Backups[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.Backups, id)
	return nil
}
