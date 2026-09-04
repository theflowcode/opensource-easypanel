package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// DeployService provisions or updates a Docker Swarm service for the given specification.
func (a *DockerAdapter) DeployService(ctx context.Context, spec domain.ServiceSpec) (*domain.DeployResult, error) {
	if err := a.EnsureNetwork(ctx, a.defaultNetwork); err != nil {
		return nil, fmt.Errorf("failed to ensure default network %s: %w", a.defaultNetwork, err)
	}

	if spec.ProjectName != "" {
		projNet := fmt.Sprintf("easypanel-%s", spec.ProjectName)
		if err := a.EnsureNetwork(ctx, projNet); err != nil {
			return nil, fmt.Errorf("failed to ensure project network %s: %w", projNet, err)
		}
	}

	swarmSpec, err := buildSwarmServiceSpec(spec, a.defaultNetwork, a.projectsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to build swarm service spec: %w", err)
	}

	svcName := formatServiceName(spec.ProjectName, spec.Name)
	existingSvc, err := a.findService(ctx, svcName)

	var targetServiceID string
	if err == nil && existingSvc != nil {
		targetServiceID = existingSvc.ID
		_, updateErr := a.cli.ServiceUpdate(ctx, existingSvc.ID, existingSvc.Version, swarmSpec, types.ServiceUpdateOptions{})
		if updateErr != nil {
			return nil, fmt.Errorf("failed to update swarm service %s: %w", svcName, updateErr)
		}
	} else {
		resp, createErr := a.cli.ServiceCreate(ctx, swarmSpec, types.ServiceCreateOptions{})
		if createErr != nil {
			return nil, fmt.Errorf("failed to create swarm service %s: %w", svcName, createErr)
		}
		targetServiceID = resp.ID
	}

	containerIDs := a.getServiceContainerIDs(ctx, targetServiceID, svcName)

	return &domain.DeployResult{
		ServiceID:    spec.ID,
		ContainerIDs: containerIDs,
		Status:       "running",
		DeployedAt:   time.Now().UTC(),
	}, nil
}

// StopService stops the service by scaling its replicas down to zero.
func (a *DockerAdapter) StopService(ctx context.Context, serviceID string) error {
	svc, err := a.findService(ctx, serviceID)
	if err != nil {
		return err
	}

	zero := uint64(0)
	if svc.Spec.Mode.Replicated != nil {
		svc.Spec.Mode.Replicated.Replicas = &zero
	}

	_, err = a.cli.ServiceUpdate(ctx, svc.ID, svc.Version, svc.Spec, types.ServiceUpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale service %s to 0: %w", serviceID, err)
	}
	return nil
}

// RestartService triggers a rolling restart by incrementing the ForceUpdate sequence counter.
func (a *DockerAdapter) RestartService(ctx context.Context, serviceID string) error {
	svc, err := a.findService(ctx, serviceID)
	if err != nil {
		return err
	}

	svc.Spec.TaskTemplate.ForceUpdate++
	_, err = a.cli.ServiceUpdate(ctx, svc.ID, svc.Version, svc.Spec, types.ServiceUpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to restart service %s: %w", serviceID, err)
	}
	return nil
}

// DeleteService removes the service from Docker Swarm and cleans up task containers.
func (a *DockerAdapter) DeleteService(ctx context.Context, serviceID string) error {
	svc, err := a.findService(ctx, serviceID)
	if err != nil {
		return err
	}

	if err := a.cli.ServiceRemove(ctx, svc.ID); err != nil {
		return fmt.Errorf("failed to remove service %s: %w", serviceID, err)
	}
	return nil
}

// findService retrieves a Swarm service by exact name/ID or Easypanel ID label.
func (a *DockerAdapter) findService(ctx context.Context, identifier string) (*swarm.Service, error) {
	inspect, _, err := a.cli.ServiceInspectWithRaw(ctx, identifier, types.ServiceInspectOptions{})
	if err == nil {
		return &inspect, nil
	}

	services, listErr := a.cli.ServiceList(ctx, types.ServiceListOptions{
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("easypanel.id=%s", identifier))),
	})
	if listErr == nil && len(services) > 0 {
		return &services[0], nil
	}

	return nil, fmt.Errorf("service not found: %s", identifier)
}

// getServiceContainerIDs finds all active container IDs belonging to a service.
func (a *DockerAdapter) getServiceContainerIDs(ctx context.Context, serviceID, serviceName string) []string {
	tasks, err := a.cli.TaskList(ctx, types.TaskListOptions{
		Filters: filters.NewArgs(filters.Arg("service", serviceID)),
	})
	if err != nil || len(tasks) == 0 {
		tasks, _ = a.cli.TaskList(ctx, types.TaskListOptions{
			Filters: filters.NewArgs(filters.Arg("service", serviceName)),
		})
	}

	var containerIDs []string
	for _, task := range tasks {
		if task.Status.ContainerStatus != nil && task.Status.ContainerStatus.ContainerID != "" {
			containerIDs = append(containerIDs, task.Status.ContainerStatus.ContainerID)
		}
	}

	if len(containerIDs) == 0 {
		containers, _ := a.cli.ContainerList(ctx, container.ListOptions{
			Filters: filters.NewArgs(filters.Arg("name", serviceName)),
		})
		for _, c := range containers {
			containerIDs = append(containerIDs, c.ID)
		}
	}

	return containerIDs
}
