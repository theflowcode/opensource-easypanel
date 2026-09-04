package mock

import (
	"context"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

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

func (m *MockDockerPort) PruneSystem(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "PruneSystem")
	if m.PruneSystemFunc != nil {
		return m.PruneSystemFunc(ctx)
	}
	return nil
}

func (m *MockDockerPort) StreamDockerEvents(ctx context.Context, eventChan chan<- domain.DockerEvent) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.StreamDockerEventsFunc != nil {
		return m.StreamDockerEventsFunc(ctx, eventChan)
	}
	return nil
}

func (m *MockDockerPort) GetServiceStorage(ctx context.Context, projectName, serviceName string) (*domain.ServiceStorage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.GetServiceStorageFunc != nil {
		return m.GetServiceStorageFunc(ctx, projectName, serviceName)
	}
	key := projectName + "/" + serviceName
	if s, exists := m.Storage[key]; exists {
		return s, nil
	}
	return &domain.ServiceStorage{
		ProjectName: projectName,
		ServiceName: serviceName,
		SizeBytes:   1024 * 1024,
		Path:        "/etc/easypanel/projects/" + projectName + "/" + serviceName,
	}, nil
}

func (m *MockDockerPort) ListStorageUsage(ctx context.Context) ([]domain.ServiceStorage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.ListStorageUsageFunc != nil {
		return m.ListStorageUsageFunc(ctx)
	}
	var list []domain.ServiceStorage
	for _, s := range m.Storage {
		list = append(list, *s)
	}
	return list, nil
}
