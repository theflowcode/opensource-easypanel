package port

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// NotifierPort defines contracts for outbound alert and notification dispatch.
type NotifierPort interface {
	// SendNotification dispatches an alert message to the specified target.
	SendNotification(ctx context.Context, target domain.NotificationTarget, payload domain.NotificationPayload) error

	// SendTestNotification verifies target connectivity by sending a test ping.
	SendTestNotification(ctx context.Context, target domain.NotificationTarget) error
}
