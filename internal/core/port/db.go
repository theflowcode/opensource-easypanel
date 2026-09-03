package port

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// DatabasePort defines metadata storage contracts for OpenSource Easypanel.
type DatabasePort interface {
	// Lifecycle & Migration
	Migrate(ctx context.Context) error
	Close() error
	WithTx(ctx context.Context, fn func(tx DatabasePort) error) error

	// Projects
	CreateProject(ctx context.Context, project *domain.Project) error
	GetProject(ctx context.Context, id string) (*domain.Project, error)
	GetProjectByName(ctx context.Context, name string) (*domain.Project, error)
	ListProjects(ctx context.Context) ([]*domain.Project, error)
	UpdateProject(ctx context.Context, project *domain.Project) error
	DeleteProject(ctx context.Context, id string) error

	// Services
	CreateService(ctx context.Context, service *domain.Service) error
	GetService(ctx context.Context, id string) (*domain.Service, error)
	GetServiceByName(ctx context.Context, projectID, name string) (*domain.Service, error)
	GetServiceByDeployToken(ctx context.Context, token string) (*domain.Service, error)
	ListServicesByProject(ctx context.Context, projectID string) ([]*domain.Service, error)
	ListAllServices(ctx context.Context) ([]*domain.Service, error)
	UpdateService(ctx context.Context, service *domain.Service) error
	DeleteService(ctx context.Context, id string) error

	// Domains
	CreateDomain(ctx context.Context, dom *domain.Domain) error
	GetDomain(ctx context.Context, id string) (*domain.Domain, error)
	ListDomainsByService(ctx context.Context, serviceID string) ([]*domain.Domain, error)
	ListAllDomains(ctx context.Context) ([]*domain.Domain, error)
	DeleteDomain(ctx context.Context, id string) error

	// Deployments
	CreateDeployment(ctx context.Context, deployment *domain.Deployment) error
	GetDeployment(ctx context.Context, id string) (*domain.Deployment, error)
	ListDeploymentsByService(ctx context.Context, serviceID string, limit, offset int) ([]*domain.Deployment, error)
	UpdateDeployment(ctx context.Context, deployment *domain.Deployment) error

	// Backups
	CreateBackup(ctx context.Context, backup *domain.Backup) error
	GetBackup(ctx context.Context, id string) (*domain.Backup, error)
	ListBackupsByService(ctx context.Context, serviceID string, limit, offset int) ([]*domain.Backup, error)
	DeleteBackup(ctx context.Context, id string) error

	// Users & Auth
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	ListUsers(ctx context.Context) ([]*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id string) error

	// Sessions & API Tokens
	CreateSession(ctx context.Context, session *domain.Session) error
	GetSession(ctx context.Context, tokenHash string) (*domain.Session, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteExpiredSessions(ctx context.Context) error

	// Settings
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, val string) error
	ListSettings(ctx context.Context) (map[string]string, error)

	// Actions & Audit Trail
	CreateAction(ctx context.Context, action *domain.Action) error
	GetAction(ctx context.Context, id string) (*domain.Action, error)
	UpdateAction(ctx context.Context, action *domain.Action) error
	ListActions(ctx context.Context, projectName, serviceName string, limit, offset int) ([]*domain.Action, error)

	// Storage Providers (Backup Targets)
	CreateStorageProvider(ctx context.Context, sp *domain.StorageProvider) error
	GetStorageProvider(ctx context.Context, id string) (*domain.StorageProvider, error)
	ListStorageProviders(ctx context.Context) ([]*domain.StorageProvider, error)
	DeleteStorageProvider(ctx context.Context, id string) error
}
