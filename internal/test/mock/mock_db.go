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
	// Backups
	Backups map[string]*domain.Backup
	// Sessions
	Sessions map[string]*domain.Session
	// Users
	Users map[string]*domain.User
	// Settings
	Settings map[string]string
	// Actions
	Actions map[string]*domain.Action
	// Storage Providers
	StorageProviders map[string]*domain.StorageProvider

	// Call tracking
	Calls []string

	// Customizable hooks
	MigrateFunc func(ctx context.Context) error
	CloseFunc   func() error
}

func NewMockDatabasePort() *MockDatabasePort {
	return &MockDatabasePort{
		Projects:         make(map[string]*domain.Project),
		Services:         make(map[string]*domain.Service),
		Domains:          make(map[string]*domain.Domain),
		Deployments:      make(map[string]*domain.Deployment),
		Backups:          make(map[string]*domain.Backup),
		Sessions:         make(map[string]*domain.Session),
		Users:            make(map[string]*domain.User),
		Settings:         make(map[string]string),
		Actions:          make(map[string]*domain.Action),
		StorageProviders: make(map[string]*domain.StorageProvider),
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
	m.Backups = make(map[string]*domain.Backup)
	m.Sessions = make(map[string]*domain.Session)
	m.Users = make(map[string]*domain.User)
	m.Settings = make(map[string]string)
	m.Actions = make(map[string]*domain.Action)
	m.StorageProviders = make(map[string]*domain.StorageProvider)
	m.Calls = nil
	m.MigrateFunc = nil
	m.CloseFunc = nil
}
