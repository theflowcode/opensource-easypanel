package mock

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.DockerPort = (*MockDockerPort)(nil)

// MockDockerPort is a thread-safe mock implementing port.DockerPort.
type MockDockerPort struct {
	mu sync.RWMutex

	// Customizable function hooks
	DeployServiceFunc       func(ctx context.Context, spec domain.ServiceSpec) (*domain.DeployResult, error)
	StopServiceFunc         func(ctx context.Context, serviceID string) error
	RestartServiceFunc      func(ctx context.Context, serviceID string) error
	DeleteServiceFunc       func(ctx context.Context, serviceID string) error
	BuildImageFunc          func(ctx context.Context, build domain.BuildConfig, logWriter io.Writer) (string, error)
	PullImageFunc           func(ctx context.Context, image string, auth *domain.RegistryAuth, logWriter io.Writer) error
	GetDockerInfoFunc       func(ctx context.Context) (*domain.DockerInfo, error)
	GetHostMetricsFunc      func(ctx context.Context) (*domain.HostMetrics, error)
	GetServiceStatusFunc    func(ctx context.Context, serviceID string) (*domain.ServiceStatus, error)
	GetServiceStatsFunc     func(ctx context.Context, serviceID string) (*domain.ServiceStats, error)
	StreamServiceLogsFunc   func(ctx context.Context, serviceID string, opts domain.LogStreamOptions, stdout, stderr io.Writer) error
	ExecServiceTerminalFunc func(ctx context.Context, serviceID string, stdin io.Reader, stdout, stderr io.Writer, resizeChan <-chan domain.TerminalSize) error
	EnsureNetworkFunc       func(ctx context.Context, networkName string) error
	EnsureVolumeFunc        func(ctx context.Context, volumeName string) error
	ListContainersFunc      func(ctx context.Context) ([]domain.ContainerSummary, error)

	// Call tracking
	Calls []string

	// In-memory state for sensible default behavior
	DeployedServices map[string]domain.ServiceSpec
	Networks         map[string]bool
	Volumes          map[string]bool
}

func NewMockDockerPort() *MockDockerPort {
	return &MockDockerPort{
		DeployedServices: make(map[string]domain.ServiceSpec),
		Networks:         make(map[string]bool),
		Volumes:          make(map[string]bool),
	}
}

// Reset clears all in-memory mock state and recorded calls.
func (m *MockDockerPort) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeployedServices = make(map[string]domain.ServiceSpec)
	m.Networks = make(map[string]bool)
	m.Volumes = make(map[string]bool)
	m.Calls = nil
	m.DeployServiceFunc = nil
	m.StopServiceFunc = nil
	m.RestartServiceFunc = nil
	m.DeleteServiceFunc = nil
	m.BuildImageFunc = nil
	m.PullImageFunc = nil
	m.GetDockerInfoFunc = nil
	m.GetHostMetricsFunc = nil
	m.GetServiceStatusFunc = nil
	m.GetServiceStatsFunc = nil
	m.StreamServiceLogsFunc = nil
	m.ExecServiceTerminalFunc = nil
	m.EnsureNetworkFunc = nil
	m.EnsureVolumeFunc = nil
	m.ListContainersFunc = nil
}

func (m *MockDockerPort) DeployService(ctx context.Context, spec domain.ServiceSpec) (*domain.DeployResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeployService")

	if m.DeployServiceFunc != nil {
		return m.DeployServiceFunc(ctx, spec)
	}

	m.DeployedServices[spec.ID] = spec
	return &domain.DeployResult{
		ServiceID:    spec.ID,
		ContainerIDs: []string{"mock-container-" + spec.ID},
		Status:       "running",
		DeployedAt:   time.Now().UTC(),
	}, nil
}

func (m *MockDockerPort) StopService(ctx context.Context, serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "StopService")

	if m.StopServiceFunc != nil {
		return m.StopServiceFunc(ctx, serviceID)
	}
	return nil
}

func (m *MockDockerPort) RestartService(ctx context.Context, serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "RestartService")

	if m.RestartServiceFunc != nil {
		return m.RestartServiceFunc(ctx, serviceID)
	}
	return nil
}

func (m *MockDockerPort) DeleteService(ctx context.Context, serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "DeleteService")

	if m.DeleteServiceFunc != nil {
		return m.DeleteServiceFunc(ctx, serviceID)
	}
	delete(m.DeployedServices, serviceID)
	return nil
}

func (m *MockDockerPort) BuildImage(ctx context.Context, build domain.BuildConfig, logWriter io.Writer) (string, error) {
	m.mu.Lock()
	m.Calls = append(m.Calls, "BuildImage")
	m.mu.Unlock()

	if m.BuildImageFunc != nil {
		return m.BuildImageFunc(ctx, build, logWriter)
	}
	tag := build.ImageTag
	if tag == "" {
		tag = "openpanel/" + build.ServiceID + ":latest"
	}
	if logWriter != nil {
		_, _ = logWriter.Write([]byte("Successfully built image: " + tag + "\n"))
	}
	return tag, nil
}

func (m *MockDockerPort) PullImage(ctx context.Context, image string, auth *domain.RegistryAuth, logWriter io.Writer) error {
	m.mu.Lock()
	m.Calls = append(m.Calls, "PullImage")
	m.mu.Unlock()

	if m.PullImageFunc != nil {
		return m.PullImageFunc(ctx, image, auth, logWriter)
	}
	if logWriter != nil {
		_, _ = logWriter.Write([]byte("Successfully pulled image: " + image + "\n"))
	}
	return nil
}

func (m *MockDockerPort) GetServiceStatus(ctx context.Context, serviceID string) (*domain.ServiceStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetServiceStatusFunc != nil {
		return m.GetServiceStatusFunc(ctx, serviceID)
	}
	st := domain.ServiceStatusRunning
	return &st, nil
}

func (m *MockDockerPort) GetServiceStats(ctx context.Context, serviceID string) (*domain.ServiceStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetServiceStatsFunc != nil {
		return m.GetServiceStatsFunc(ctx, serviceID)
	}
	return &domain.ServiceStats{
		ServiceID:        serviceID,
		CPUPercentage:    1.5,
		MemoryUsageBytes: 32 * 1024 * 1024,
		MemoryLimitBytes: 512 * 1024 * 1024,
		ReadAt:           time.Now().UTC(),
	}, nil
}

func (m *MockDockerPort) StreamServiceLogs(ctx context.Context, serviceID string, opts domain.LogStreamOptions, stdout, stderr io.Writer) error {
	m.mu.Lock()
	m.Calls = append(m.Calls, "StreamServiceLogs")
	m.mu.Unlock()

	if m.StreamServiceLogsFunc != nil {
		return m.StreamServiceLogsFunc(ctx, serviceID, opts, stdout, stderr)
	}
	_, err := stdout.Write([]byte("mock service log output\n"))
	return err
}

func (m *MockDockerPort) ExecServiceTerminal(ctx context.Context, serviceID string, stdin io.Reader, stdout, stderr io.Writer, resizeChan <-chan domain.TerminalSize) error {
	m.mu.Lock()
	m.Calls = append(m.Calls, "ExecServiceTerminal")
	m.mu.Unlock()

	if m.ExecServiceTerminalFunc != nil {
		return m.ExecServiceTerminalFunc(ctx, serviceID, stdin, stdout, stderr, resizeChan)
	}
	_, err := stdout.Write([]byte("mock terminal connected\n"))
	return err
}

func (m *MockDockerPort) EnsureNetwork(ctx context.Context, networkName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "EnsureNetwork")

	if m.EnsureNetworkFunc != nil {
		return m.EnsureNetworkFunc(ctx, networkName)
	}
	m.Networks[networkName] = true
	return nil
}

func (m *MockDockerPort) EnsureVolume(ctx context.Context, volumeName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "EnsureVolume")

	if m.EnsureVolumeFunc != nil {
		return m.EnsureVolumeFunc(ctx, volumeName)
	}
	m.Volumes[volumeName] = true
	return nil
}

func (m *MockDockerPort) ListContainers(ctx context.Context) ([]domain.ContainerSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.ListContainersFunc != nil {
		return m.ListContainersFunc(ctx)
	}

	var list []domain.ContainerSummary
	for id, spec := range m.DeployedServices {
		list = append(list, domain.ContainerSummary{
			ID:        "mock-container-" + id,
			Names:     []string{"/" + spec.Name},
			Image:     spec.Image,
			Status:    "Up 2 hours",
			State:     "running",
			CreatedAt: time.Now().UTC(),
		})
	}
	return list, nil
}

func (m *MockDockerPort) GetDockerInfo(ctx context.Context) (*domain.DockerInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.GetDockerInfoFunc != nil {
		return m.GetDockerInfoFunc(ctx)
	}
	return &domain.DockerInfo{
		ServerVersion:     "mock-docker/24.0.0",
		ContainersTotal:   len(m.DeployedServices),
		ContainersRunning: len(m.DeployedServices),
	}, nil
}

func (m *MockDockerPort) GetHostMetrics(ctx context.Context) (*domain.HostMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.GetHostMetricsFunc != nil {
		return m.GetHostMetricsFunc(ctx)
	}
	return &domain.HostMetrics{
		CPUPercent:       12.5,
		MemoryUsedBytes:  512 * 1024 * 1024,
		MemoryTotalBytes: 4 * 1024 * 1024 * 1024,
		DiskUsedBytes:    10 * 1024 * 1024 * 1024,
		DiskTotalBytes:   50 * 1024 * 1024 * 1024,
		UptimeSeconds:    86400,
		ReadAt:           time.Now().UTC(),
	}, nil
}
