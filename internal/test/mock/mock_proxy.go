package mock

import (
	"context"
	"sync"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.ProxyDriverPort = (*MockProxyDriverPort)(nil)

// MockProxyDriverPort is a thread-safe mock implementing port.ProxyDriverPort.
type MockProxyDriverPort struct {
	mu sync.RWMutex

	// Customizable function hooks
	ApplyRouteFunc    func(ctx context.Context, route domain.RouteConfig) error
	RemoveRouteFunc   func(ctx context.Context, serviceID string) error
	SyncAllRoutesFunc func(ctx context.Context, routes []domain.RouteConfig) error

	// Call tracking
	Calls []string

	// In-memory state
	Routes map[string]domain.RouteConfig
}

func NewMockProxyDriverPort() *MockProxyDriverPort {
	return &MockProxyDriverPort{
		Routes: make(map[string]domain.RouteConfig),
	}
}

// Reset clears all in-memory mock state and recorded calls.
func (m *MockProxyDriverPort) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Routes = make(map[string]domain.RouteConfig)
	m.Calls = nil
	m.ApplyRouteFunc = nil
	m.RemoveRouteFunc = nil
	m.SyncAllRoutesFunc = nil
}

func (m *MockProxyDriverPort) ApplyRoute(ctx context.Context, route domain.RouteConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "ApplyRoute")

	if m.ApplyRouteFunc != nil {
		return m.ApplyRouteFunc(ctx, route)
	}
	m.Routes[route.ServiceID] = route
	return nil
}

func (m *MockProxyDriverPort) RemoveRoute(ctx context.Context, serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "RemoveRoute")

	if m.RemoveRouteFunc != nil {
		return m.RemoveRouteFunc(ctx, serviceID)
	}
	delete(m.Routes, serviceID)
	return nil
}

func (m *MockProxyDriverPort) SyncAllRoutes(ctx context.Context, routes []domain.RouteConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "SyncAllRoutes")

	if m.SyncAllRoutesFunc != nil {
		return m.SyncAllRoutesFunc(ctx, routes)
	}
	m.Routes = make(map[string]domain.RouteConfig)
	for _, r := range routes {
		m.Routes[r.ServiceID] = r
	}
	return nil
}
