package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// GetDockerInfo returns current Docker Engine and Swarm cluster metadata.
func (a *DockerAdapter) GetDockerInfo(ctx context.Context) (*domain.DockerInfo, error) {
	info, err := a.cli.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve docker info: %w", err)
	}

	return &domain.DockerInfo{
		ServerVersion:     info.ServerVersion,
		SwarmActive:       info.Swarm.LocalNodeState == "active",
		IsManager:         info.Swarm.ControlAvailable,
		NodeID:            info.Swarm.NodeID,
		ContainersTotal:   info.Containers,
		ContainersRunning: info.ContainersRunning,
		ContainersStopped: info.ContainersStopped,
		ImagesTotal:       info.Images,
	}, nil
}

// GetServiceStatus returns the real-time lifecycle status of a Swarm service.
func (a *DockerAdapter) GetServiceStatus(ctx context.Context, serviceID string) (*domain.ServiceStatus, error) {
	svc, err := a.findService(ctx, serviceID)
	if err != nil {
		st := domain.ServiceStatusStopped
		return &st, nil
	}

	if svc.Spec.Mode.Replicated != nil && svc.Spec.Mode.Replicated.Replicas != nil && *svc.Spec.Mode.Replicated.Replicas == 0 {
		st := domain.ServiceStatusStopped
		return &st, nil
	}

	tasks, err := a.cli.TaskList(ctx, types.TaskListOptions{
		Filters: filters.NewArgs(filters.Arg("service", svc.ID)),
	})
	if err != nil || len(tasks) == 0 {
		st := domain.ServiceStatusStarting
		return &st, nil
	}

	for _, t := range tasks {
		switch t.Status.State {
		case swarm.TaskStateRunning:
			st := domain.ServiceStatusRunning
			return &st, nil
		case swarm.TaskStatePreparing, swarm.TaskStateStarting, swarm.TaskStateReady:
			st := domain.ServiceStatusStarting
			return &st, nil
		case swarm.TaskStateFailed, swarm.TaskStateRejected:
			st := domain.ServiceStatusFailed
			return &st, nil
		}
	}

	st := domain.ServiceStatusRunning
	return &st, nil
}

// GetServiceStats gathers real-time CPU, memory, and network telemetry for a service container.
func (a *DockerAdapter) GetServiceStats(ctx context.Context, serviceID string) (*domain.ServiceStats, error) {
	containerIDs := a.getServiceContainerIDs(ctx, serviceID, serviceID)
	if len(containerIDs) == 0 {
		return &domain.ServiceStats{
			ServiceID: serviceID,
			ReadAt:    time.Now().UTC(),
		}, nil
	}

	statsResp, err := a.cli.ContainerStatsOneShot(ctx, containerIDs[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read container stats for %s: %w", containerIDs[0], err)
	}
	defer statsResp.Body.Close()

	var stats container.StatsResponse
	if err := json.NewDecoder(statsResp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode container stats json: %w", err)
	}

	cpuPercent := calculateCPUPercent(&stats)
	memUsage := stats.MemoryStats.Usage
	if v, ok := stats.MemoryStats.Stats["inactive_file"]; ok && memUsage > v {
		memUsage -= v
	}

	var rxBytes, txBytes uint64
	for _, net := range stats.Networks {
		rxBytes += net.RxBytes
		txBytes += net.TxBytes
	}

	return &domain.ServiceStats{
		ServiceID:          serviceID,
		CPUPercentage:      cpuPercent,
		MemoryUsageBytes:   memUsage,
		MemoryLimitBytes:   stats.MemoryStats.Limit,
		NetworkRxBytes:     rxBytes,
		NetworkTxBytes:     txBytes,
		NetworkInputBytes:  rxBytes,
		NetworkOutputBytes: txBytes,
		ReadAt:             time.Now().UTC(),
	}, nil
}

// calculateCPUPercent computes CPU utilization percentage from Docker stats.
func calculateCPUPercent(stats *container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)
	onlineCPUs := float64(stats.CPUStats.OnlineCPUs)
	if onlineCPUs == 0 {
		onlineCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}
	if onlineCPUs == 0 {
		onlineCPUs = 1.0
	}

	if systemDelta > 0 && cpuDelta > 0 {
		return (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}
	return 0.0
}
