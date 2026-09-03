package noop

import (
	"context"
	"io"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.DockerPort = (*NoOpDocker)(nil)

// NoOpDocker is a Null Object implementation of port.DockerPort for headless/disabled mode.
type NoOpDocker struct{}

func NewNoOpDocker() *NoOpDocker {
	return &NoOpDocker{}
}

func (n *NoOpDocker) DeployService(ctx context.Context, spec domain.ServiceSpec) (*domain.DeployResult, error) {
	return &domain.DeployResult{
		ServiceID:    spec.ID,
		ContainerIDs: []string{"noop-" + spec.ID},
		Status:       "running",
		DeployedAt:   time.Now().UTC(),
	}, nil
}

func (n *NoOpDocker) StopService(ctx context.Context, serviceID string) error {
	return nil
}

func (n *NoOpDocker) RestartService(ctx context.Context, serviceID string) error {
	return nil
}

func (n *NoOpDocker) DeleteService(ctx context.Context, serviceID string) error {
	return nil
}

func (n *NoOpDocker) BuildImage(ctx context.Context, build domain.BuildConfig, logWriter io.Writer) (string, error) {
	tag := build.ImageTag
	if tag == "" {
		tag = "openpanel/" + build.ServiceID + ":latest"
	}
	return tag, nil
}

func (n *NoOpDocker) PullImage(ctx context.Context, image string, auth *domain.RegistryAuth, logWriter io.Writer) error {
	return nil
}

func (n *NoOpDocker) GetServiceStatus(ctx context.Context, serviceID string) (*domain.ServiceStatus, error) {
	st := domain.ServiceStatusRunning
	return &st, nil
}

func (n *NoOpDocker) GetServiceStats(ctx context.Context, serviceID string) (*domain.ServiceStats, error) {
	return &domain.ServiceStats{
		ServiceID:        serviceID,
		CPUPercentage:    0.0,
		MemoryUsageBytes: 0,
		MemoryLimitBytes: 0,
		ReadAt:           time.Now().UTC(),
	}, nil
}

func (n *NoOpDocker) StreamServiceLogs(ctx context.Context, serviceID string, opts domain.LogStreamOptions, stdout, stderr io.Writer) error {
	return nil
}

func (n *NoOpDocker) ExecServiceTerminal(ctx context.Context, serviceID string, stdin io.Reader, stdout, stderr io.Writer, resizeChan <-chan domain.TerminalSize) error {
	return nil
}

func (n *NoOpDocker) EnsureNetwork(ctx context.Context, networkName string) error {
	return nil
}

func (n *NoOpDocker) EnsureVolume(ctx context.Context, volumeName string) error {
	return nil
}

func (n *NoOpDocker) ListContainers(ctx context.Context) ([]domain.ContainerSummary, error) {
	return []domain.ContainerSummary{}, nil
}
