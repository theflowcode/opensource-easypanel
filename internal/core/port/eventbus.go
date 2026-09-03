package port

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// EventBusPort defines pub/sub contracts for cross-module async notifications.
type EventBusPort interface {
	// Publish dispatches an event asynchronously to all registered subscribers.
	Publish(ctx context.Context, event domain.Event) error

	// Subscribe registers an event handler for a specific event type.
	Subscribe(ctx context.Context, eventType domain.EventType, handler domain.EventHandler) (domain.Subscription, error)
}
