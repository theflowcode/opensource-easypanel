package port

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// ProxyDriverPort defines the contract for dynamic reverse proxy routing.
type ProxyDriverPort interface {
	// ApplyRoute writes or updates the routing configuration for a service.
	ApplyRoute(ctx context.Context, route domain.RouteConfig) error

	// RemoveRoute tears down proxy routing for a service.
	RemoveRoute(ctx context.Context, serviceID string) error

	// SyncAllRoutes applies an atomic batch synchronization of all active routes.
	SyncAllRoutes(ctx context.Context, routes []domain.RouteConfig) error
}
