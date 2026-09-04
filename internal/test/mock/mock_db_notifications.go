package mock

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (m *MockDatabasePort) CreateNotificationChannel(ctx context.Context, ch *domain.NotificationChannel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "CreateNotificationChannel")

	if err := ch.Validate(); err != nil {
		return err
	}
	if _, exists := m.NotificationChannels[ch.ID]; exists {
		return domain.ErrAlreadyExists
	}
	for _, existing := range m.NotificationChannels {
		if existing.Name == ch.Name {
			return domain.ErrAlreadyExists
		}
	}
	cp := *ch
	m.NotificationChannels[ch.ID] = &cp
	return nil
}

func (m *MockDatabasePort) GetNotificationChannel(ctx context.Context, id string) (*domain.NotificationChannel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "GetNotificationChannel")

	ch, exists := m.NotificationChannels[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	cp := *ch
	return &cp, nil
}

func (m *MockDatabasePort) ListNotificationChannels(ctx context.Context) ([]*domain.NotificationChannel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.Calls = append(m.Calls, "ListNotificationChannels")

	var list []*domain.NotificationChannel
	for _, ch := range m.NotificationChannels {
		cp := *ch
		list = append(list, &cp)
	}
	return list, nil
}

func (m *MockDatabasePort) UpdateNotificationChannel(ctx context.Context, ch *domain.NotificationChannel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "UpdateNotificationChannel")

	if err := ch.Validate(); err != nil {
		return err
	}
	if _, exists := m.NotificationChannels[ch.ID]; !exists {
		return domain.ErrNotFound
	}
	cp := *ch
	m.NotificationChannels[ch.ID] = &cp
	return nil
}

func (m *MockDatabasePort) DeleteNotificationChannel(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteNotificationChannel")

	if _, exists := m.NotificationChannels[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.NotificationChannels, id)
	return nil
}
