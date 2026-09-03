package port

import (
	"context"
	"io"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// DockerPort defines outbound contracts for container and swarm lifecycle management.
type DockerPort interface {
	// Service Lifecycle
	DeployService(ctx context.Context, spec domain.ServiceSpec) (*domain.DeployResult, error)
	StopService(ctx context.Context, serviceID string) error
	RestartService(ctx context.Context, serviceID string) error
	DeleteService(ctx context.Context, serviceID string) error

	// Image Building & Pulling
	BuildImage(ctx context.Context, build domain.BuildConfig, logWriter io.Writer) (imageTag string, err error)
	PullImage(ctx context.Context, image string, auth *domain.RegistryAuth, logWriter io.Writer) error

	// Inspection, Monitoring & Streaming
	GetDockerInfo(ctx context.Context) (*domain.DockerInfo, error)
	GetHostMetrics(ctx context.Context) (*domain.HostMetrics, error)
	GetServiceStatus(ctx context.Context, serviceID string) (*domain.ServiceStatus, error)
	GetServiceStats(ctx context.Context, serviceID string) (*domain.ServiceStats, error)
	StreamServiceLogs(ctx context.Context, serviceID string, opts domain.LogStreamOptions, stdout, stderr io.Writer) error
	ExecServiceTerminal(ctx context.Context, serviceID string, stdin io.Reader, stdout, stderr io.Writer, resizeChan <-chan domain.TerminalSize) error

	// Infrastructure Management
	EnsureNetwork(ctx context.Context, networkName string) error
	EnsureVolume(ctx context.Context, volumeName string) error
	ListContainers(ctx context.Context) ([]domain.ContainerSummary, error)
}
