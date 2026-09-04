package noop

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.NotifierPort = (*NoOpNotifier)(nil)

// NoOpNotifier is a Null Object implementation of port.NotifierPort when outbound notifications are disabled.
type NoOpNotifier struct{}

func NewNoOpNotifier() *NoOpNotifier {
	return &NoOpNotifier{}
}

func (n *NoOpNotifier) SendNotification(ctx context.Context, target domain.NotificationTarget, payload domain.NotificationPayload) error {
	return nil
}

func (n *NoOpNotifier) SendTestNotification(ctx context.Context, target domain.NotificationTarget) error {
	return nil
}
