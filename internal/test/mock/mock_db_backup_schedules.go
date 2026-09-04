package mock

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (m *MockDatabasePort) CreateBackupSchedule(ctx context.Context, bs *domain.BackupSchedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateBackupSchedule")

	if err := bs.Validate(); err != nil {
		return err
	}
	if _, exists := m.BackupSchedules[bs.ID]; exists {
		return domain.ErrAlreadyExists
	}
	cp := *bs
	m.BackupSchedules[bs.ID] = &cp
	return nil
}

func (m *MockDatabasePort) GetBackupSchedule(ctx context.Context, id string) (*domain.BackupSchedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "GetBackupSchedule")

	bs, exists := m.BackupSchedules[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	cp := *bs
	return &cp, nil
}

func (m *MockDatabasePort) ListBackupSchedulesByService(ctx context.Context, projectName, serviceName string) ([]*domain.BackupSchedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "ListBackupSchedulesByService")

	var list []*domain.BackupSchedule
	for _, bs := range m.BackupSchedules {
		if bs.ProjectName == projectName && bs.ServiceName == serviceName {
			cp := *bs
			list = append(list, &cp)
		}
	}
	return list, nil
}

func (m *MockDatabasePort) UpdateBackupSchedule(ctx context.Context, bs *domain.BackupSchedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "UpdateBackupSchedule")

	if err := bs.Validate(); err != nil {
		return err
	}
	if _, exists := m.BackupSchedules[bs.ID]; !exists {
		return domain.ErrNotFound
	}
	cp := *bs
	m.BackupSchedules[bs.ID] = &cp
	return nil
}

func (m *MockDatabasePort) DeleteBackupSchedule(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteBackupSchedule")

	if _, exists := m.BackupSchedules[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.BackupSchedules, id)
	return nil
}
