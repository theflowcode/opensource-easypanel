package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// GetServiceStorage calculates the host disk usage of a specific service.
func (a *DockerAdapter) GetServiceStorage(ctx context.Context, projectName, serviceName string) (*domain.ServiceStorage, error) {
	servicePath := filepath.Join(a.projectsDir, projectName, serviceName)
	size := calculateDirSize(servicePath)

	return &domain.ServiceStorage{
		ProjectName: projectName,
		ServiceName: serviceName,
		SizeBytes:   size,
		Path:        servicePath,
	}, nil
}

// ListStorageUsage scans the base projects directory and calculates disk consumption per service.
func (a *DockerAdapter) ListStorageUsage(ctx context.Context) ([]domain.ServiceStorage, error) {
	var results []domain.ServiceStorage

	projEntries, err := os.ReadDir(a.projectsDir)
	if err != nil {
		return results, nil
	}

	for _, pEntry := range projEntries {
		if !pEntry.IsDir() {
			continue
		}
		projectName := pEntry.Name()
		projPath := filepath.Join(a.projectsDir, projectName)

		svcEntries, err := os.ReadDir(projPath)
		if err != nil {
			continue
		}

		for _, sEntry := range svcEntries {
			if !sEntry.IsDir() {
				continue
			}
			serviceName := sEntry.Name()
			svcPath := filepath.Join(projPath, serviceName)
			results = append(results, domain.ServiceStorage{
				ProjectName: projectName,
				ServiceName: serviceName,
				SizeBytes:   calculateDirSize(svcPath),
				Path:        svcPath,
			})
		}
	}

	return results, nil
}

// EnsureNetwork ensures the requested network exists, creating an attachable overlay or bridge network.
func (a *DockerAdapter) EnsureNetwork(ctx context.Context, networkName string) error {
	if networkName == "" {
		return nil
	}

	_, err := a.cli.NetworkInspect(ctx, networkName, network.InspectOptions{})
	if err == nil {
		return nil
	}

	driver := "bridge"
	if a.isSwarm {
		driver = "overlay"
	}

	opts := network.CreateOptions{
		Driver:     driver,
		Attachable: true,
	}

	_, err = a.cli.NetworkCreate(ctx, networkName, opts)
	if err != nil {
		// If another concurrent call created it, verify inspection again
		if _, inspectErr := a.cli.NetworkInspect(ctx, networkName, network.InspectOptions{}); inspectErr == nil {
			return nil
		}
		return fmt.Errorf("failed to create network %s: %w", networkName, err)
	}

	return nil
}

// EnsureVolume ensures that a named Docker volume exists.
func (a *DockerAdapter) EnsureVolume(ctx context.Context, volumeName string) error {
	if volumeName == "" {
		return nil
	}

	_, err := a.cli.VolumeInspect(ctx, volumeName)
	if err == nil {
		return nil
	}

	_, err = a.cli.VolumeCreate(ctx, volume.CreateOptions{
		Name: volumeName,
	})
	if err != nil {
		if _, inspectErr := a.cli.VolumeInspect(ctx, volumeName); inspectErr == nil {
			return nil
		}
		return fmt.Errorf("failed to create volume %s: %w", volumeName, err)
	}
	return nil
}

// ListContainers returns lightweight summaries of all containers on the host.
func (a *DockerAdapter) ListContainers(ctx context.Context) ([]domain.ContainerSummary, error) {
	containers, err := a.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	summaries := make([]domain.ContainerSummary, 0, len(containers))
	for _, c := range containers {
		summaries = append(summaries, domain.ContainerSummary{
			ID:        c.ID,
			Names:     c.Names,
			Image:     c.Image,
			Status:    c.Status,
			State:     c.State,
			CreatedAt: time.Unix(c.Created, 0),
		})
	}
	return summaries, nil
}

// PruneSystem purges unused containers, images, volumes, and networks.
func (a *DockerAdapter) PruneSystem(ctx context.Context) error {
	emptyFilters := filters.Args{}
	_, _ = a.cli.ContainersPrune(ctx, emptyFilters)
	_, _ = a.cli.ImagesPrune(ctx, emptyFilters)
	_, _ = a.cli.VolumesPrune(ctx, emptyFilters)
	_, _ = a.cli.NetworksPrune(ctx, emptyFilters)
	return nil
}

func calculateDirSize(path string) int64 {
	var total int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}
