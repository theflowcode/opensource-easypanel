package noop

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.EventBusPort = (*NoOpEventBus)(nil)

type noOpSubscription struct{}

func (s *noOpSubscription) Unsubscribe() {}

// NoOpEventBus is a Null Object implementation of port.EventBusPort.
type NoOpEventBus struct{}

func NewNoOpEventBus() *NoOpEventBus {
	return &NoOpEventBus{}
}

func (n *NoOpEventBus) Publish(ctx context.Context, event domain.Event) error {
	return nil
}

func (n *NoOpEventBus) Subscribe(ctx context.Context, eventType domain.EventType, handler domain.EventHandler) (domain.Subscription, error) {
	return &noOpSubscription{}, nil
}
