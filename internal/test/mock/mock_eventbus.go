package mock

import (
	"context"
	"sync"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.EventBusPort = (*MockEventBusPort)(nil)

type mockSubscription struct {
	onUnsubscribe func()
}

func (s *mockSubscription) Unsubscribe() {
	if s.onUnsubscribe != nil {
		s.onUnsubscribe()
	}
}

// MockEventBusPort is a thread-safe mock implementing port.EventBusPort.
type MockEventBusPort struct {
	mu sync.RWMutex

	subscribers map[domain.EventType][]domain.EventHandler
	Events      []domain.Event
	Calls       []string

	// Customizable hooks
	PublishFunc   func(ctx context.Context, event domain.Event) error
	SubscribeFunc func(ctx context.Context, eventType domain.EventType, handler domain.EventHandler) (domain.Subscription, error)
}

func NewMockEventBusPort() *MockEventBusPort {
	return &MockEventBusPort{
		subscribers: make(map[domain.EventType][]domain.EventHandler),
	}
}

// Reset clears all in-memory mock state, subscribers, and recorded calls.
func (m *MockEventBusPort) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers = make(map[domain.EventType][]domain.EventHandler)
	m.Events = nil
	m.Calls = nil
	m.PublishFunc = nil
	m.SubscribeFunc = nil
}

func (m *MockEventBusPort) Publish(ctx context.Context, event domain.Event) error {
	m.mu.Lock()
	m.Calls = append(m.Calls, "Publish")
	m.Events = append(m.Events, event)
	handlers := append([]domain.EventHandler(nil), m.subscribers[event.Type]...)
	wildcardHandlers := append([]domain.EventHandler(nil), m.subscribers["*"]...)
	publishFunc := m.PublishFunc
	m.mu.Unlock()

	if publishFunc != nil {
		return publishFunc(ctx, event)
	}

	for _, h := range handlers {
		h(event)
	}
	for _, h := range wildcardHandlers {
		h(event)
	}
	return nil
}

func (m *MockEventBusPort) Subscribe(ctx context.Context, eventType domain.EventType, handler domain.EventHandler) (domain.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "Subscribe")

	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(ctx, eventType, handler)
	}

	m.subscribers[eventType] = append(m.subscribers[eventType], handler)
	sub := &mockSubscription{
		onUnsubscribe: func() {
			m.mu.Lock()
			defer m.mu.Unlock()
			handlers := m.subscribers[eventType]
			for i, h := range handlers {
				// Compare function pointers
				if &h == &handler {
					m.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
					break
				}
			}
		},
	}
	return sub, nil
}
