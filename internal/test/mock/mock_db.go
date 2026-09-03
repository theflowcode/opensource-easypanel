package mock

import (
	"context"
	"sync"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.DatabasePort = (*MockDatabasePort)(nil)

// MockDatabasePort is a thread-safe in-memory implementation of port.DatabasePort.
type MockDatabasePort struct {
	mu sync.RWMutex

	// Projects
	Projects map[string]*domain.Project
	// Services
	Services map[string]*domain.Service
	// Domains
	Domains map[string]*domain.Domain
	// Deployments
	Deployments map[string]*domain.Deployment
	// Users
	Users map[string]*domain.User
	// Settings
	Settings map[string]string

	// Call tracking
	Calls []string

	// Customizable hooks
	MigrateFunc func(ctx context.Context) error
	CloseFunc   func() error
}

func NewMockDatabasePort() *MockDatabasePort {
	return &MockDatabasePort{
		Projects:    make(map[string]*domain.Project),
		Services:    make(map[string]*domain.Service),
		Domains:     make(map[string]*domain.Domain),
		Deployments: make(map[string]*domain.Deployment),
		Users:       make(map[string]*domain.User),
		Settings:    make(map[string]string),
	}
}

func (m *MockDatabasePort) Migrate(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "Migrate")
	if m.MigrateFunc != nil {
		return m.MigrateFunc(ctx)
	}
	return nil
}

func (m *MockDatabasePort) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "Close")
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockDatabasePort) WithTx(ctx context.Context, fn func(tx port.DatabasePort) error) error {
	m.mu.Lock()
	m.Calls = append(m.Calls, "WithTx")
	m.mu.Unlock()

	return fn(m)
}

// Reset clears all in-memory mock state and recorded calls.
func (m *MockDatabasePort) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Projects = make(map[string]*domain.Project)
	m.Services = make(map[string]*domain.Service)
	m.Domains = make(map[string]*domain.Domain)
	m.Deployments = make(map[string]*domain.Deployment)
	m.Users = make(map[string]*domain.User)
	m.Settings = make(map[string]string)
	m.Calls = nil
	m.MigrateFunc = nil
	m.CloseFunc = nil
}
