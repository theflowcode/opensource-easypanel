package noop

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.ProxyDriverPort = (*NoOpProxyDriver)(nil)

// NoOpProxyDriver is a Null Object implementation of port.ProxyDriverPort when reverse proxy routing is disabled.
type NoOpProxyDriver struct{}

func NewNoOpProxyDriver() *NoOpProxyDriver {
	return &NoOpProxyDriver{}
}

func (n *NoOpProxyDriver) ApplyRoute(ctx context.Context, route domain.RouteConfig) error {
	return nil
}

func (n *NoOpProxyDriver) RemoveRoute(ctx context.Context, serviceID string) error {
	return nil
}

func (n *NoOpProxyDriver) SyncAllRoutes(ctx context.Context, routes []domain.RouteConfig) error {
	return nil
}
