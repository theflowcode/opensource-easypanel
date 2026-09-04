package mock

import (
	"context"
	"sync"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.NotifierPort = (*MockNotifierPort)(nil)

// SentNotification records a notification dispatched to a target.
type SentNotification struct {
	Target  domain.NotificationTarget
	Payload domain.NotificationPayload
}

// MockNotifierPort is a thread-safe mock implementation of port.NotifierPort for testing.
type MockNotifierPort struct {
	mu sync.RWMutex

	Calls []string
	Sent  []SentNotification
	Tests []domain.NotificationTarget

	SendNotificationFunc     func(ctx context.Context, target domain.NotificationTarget, payload domain.NotificationPayload) error
	SendTestNotificationFunc func(ctx context.Context, target domain.NotificationTarget) error
}

func NewMockNotifierPort() *MockNotifierPort {
	return &MockNotifierPort{}
}

func (m *MockNotifierPort) SendNotification(ctx context.Context, target domain.NotificationTarget, payload domain.NotificationPayload) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "SendNotification")
	m.Sent = append(m.Sent, SentNotification{Target: target, Payload: payload})

	if m.SendNotificationFunc != nil {
		return m.SendNotificationFunc(ctx, target, payload)
	}
	return nil
}

func (m *MockNotifierPort) SendTestNotification(ctx context.Context, target domain.NotificationTarget) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "SendTestNotification")
	m.Tests = append(m.Tests, target)

	if m.SendTestNotificationFunc != nil {
		return m.SendTestNotificationFunc(ctx, target)
	}
	return nil
}

// Reset clears recorded calls and custom hooks.
func (m *MockNotifierPort) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = nil
	m.Sent = nil
	m.Tests = nil
	m.SendNotificationFunc = nil
	m.SendTestNotificationFunc = nil
}
