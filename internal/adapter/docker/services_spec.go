package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// formatServiceName generates the standard Easypanel swarm service name.
func formatServiceName(projectName, serviceName string) string {
	if projectName == "" {
		return serviceName
	}
	return fmt.Sprintf("%s_%s", projectName, serviceName)
}

// formatEnvVars converts domain EnvVar slice into KEY=VALUE format.
func formatEnvVars(vars []domain.EnvVar) []string {
	result := make([]string, 0, len(vars))
	for _, v := range vars {
		result = append(result, fmt.Sprintf("%s=%s", v.Key, v.Value))
	}
	return result
}

// prepareMounts transforms domain volume mounts to Swarm mount specs.
func prepareMounts(spec domain.ServiceSpec, projectsDir string) ([]mount.Mount, error) {
	mounts := make([]mount.Mount, 0, len(spec.Volumes))
	serviceDir := filepath.Join(projectsDir, spec.ProjectName, spec.Name)

	for _, v := range spec.Volumes {
		switch v.Type {
		case "file":
			filePath := filepath.Join(serviceDir, v.Name)
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create file mount dir %s: %w", filepath.Dir(filePath), err)
			}
			if err := os.WriteFile(filePath, []byte(v.Content), 0644); err != nil {
				return nil, fmt.Errorf("failed to write mount file %s: %w", filePath, err)
			}
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   filePath,
				Target:   v.ContainerPath,
				ReadOnly: v.ReadOnly,
			})
		case "bind":
			hostPath := v.HostPath
			if hostPath == "" {
				hostPath = filepath.Join(serviceDir, v.Name)
			}
			if err := os.MkdirAll(hostPath, 0755); err != nil {
				return nil, fmt.Errorf("failed to create host bind dir %s: %w", hostPath, err)
			}
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   hostPath,
				Target:   v.ContainerPath,
				ReadOnly: v.ReadOnly,
			})
		default: // "volume" or empty
			volName := v.Name
			if !strings.HasPrefix(volName, spec.ProjectName+"_") {
				volName = fmt.Sprintf("%s_%s_%s", spec.ProjectName, spec.Name, v.Name)
			}
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeVolume,
				Source:   volName,
				Target:   v.ContainerPath,
				ReadOnly: v.ReadOnly,
			})
		}
	}
	return mounts, nil
}

// buildRestartPolicy translates restart policy string into Swarm restart policy.
func buildRestartPolicy(policy string) *swarm.RestartPolicy {
	cond := swarm.RestartPolicyConditionAny
	switch policy {
	case domain.RestartPolicyOnFailure:
		cond = swarm.RestartPolicyConditionOnFailure
	case domain.RestartPolicyNo:
		cond = swarm.RestartPolicyConditionNone
	case domain.RestartPolicyAlways, domain.RestartPolicyUnlessStopped:
		cond = swarm.RestartPolicyConditionAny
	}
	return &swarm.RestartPolicy{
		Condition: cond,
	}
}

// buildHealthConfig converts domain health check to container health config.
func buildHealthConfig(hc *domain.HealthCheckConfig) *container.HealthConfig {
	if hc == nil || len(hc.Test) == 0 {
		return nil
	}
	cfg := &container.HealthConfig{
		Test: hc.Test,
	}
	if hc.IntervalSeconds > 0 {
		cfg.Interval = time.Duration(hc.IntervalSeconds) * time.Second
	}
	if hc.TimeoutSeconds > 0 {
		cfg.Timeout = time.Duration(hc.TimeoutSeconds) * time.Second
	}
	if hc.Retries > 0 {
		cfg.Retries = hc.Retries
	}
	if hc.StartPeriodSeconds > 0 {
		cfg.StartPeriod = time.Duration(hc.StartPeriodSeconds) * time.Second
	}
	return cfg
}

// buildResourceLimits converts domain resource limits into Swarm resource requirements.
func buildResourceLimits(res domain.ResourceLimits) *swarm.ResourceRequirements {
	if res.CPULimit <= 0 && res.MemoryLimit <= 0 {
		return nil
	}
	limits := &swarm.Limit{}
	if res.CPULimit > 0 {
		limits.NanoCPUs = int64(res.CPULimit * 1e9)
	}
	if res.MemoryLimit > 0 {
		limits.MemoryBytes = res.MemoryLimit * 1024 * 1024
	}
	return &swarm.ResourceRequirements{
		Limits: limits,
	}
}

// buildSwarmServiceSpec generates a complete Swarm ServiceSpec from domain.ServiceSpec.
func buildSwarmServiceSpec(spec domain.ServiceSpec, defaultNet, projectsDir string) (swarm.ServiceSpec, error) {
	mounts, err := prepareMounts(spec, projectsDir)
	if err != nil {
		return swarm.ServiceSpec{}, err
	}

	labels := map[string]string{
		"easypanel.project": spec.ProjectName,
		"easypanel.service": spec.Name,
		"easypanel.id":      spec.ID,
		"easypanel.type":    string(spec.Type),
		"easypanel.managed": "true",
	}
	for k, v := range spec.Labels {
		labels[k] = v
	}

	initProcess := true
	containerSpec := &swarm.ContainerSpec{
		Image:       spec.Image,
		Env:         formatEnvVars(spec.EnvVars),
		Mounts:      mounts,
		Labels:      labels,
		Init:        &initProcess,
		Healthcheck: buildHealthConfig(spec.HealthCheck),
	}
	if spec.Command != "" {
		containerSpec.Command = strings.Fields(spec.Command)
	}
	if len(spec.Args) > 0 {
		containerSpec.Args = spec.Args
	}

	networks := []swarm.NetworkAttachmentConfig{
		{Target: defaultNet},
	}
	if spec.ProjectName != "" {
		projNet := fmt.Sprintf("easypanel-%s", spec.ProjectName)
		if projNet != defaultNet {
			networks = append(networks, swarm.NetworkAttachmentConfig{Target: projNet})
		}
	}

	replicas := uint64(spec.Replicas)
	if replicas == 0 {
		replicas = 1
	}

	order := swarm.UpdateOrderStopFirst
	if spec.ZeroDowntime {
		order = swarm.UpdateOrderStartFirst
	}

	portConfigs := make([]swarm.PortConfig, 0, len(spec.Ports))
	for _, p := range spec.Ports {
		if p.HostPort > 0 && p.ContainerPort > 0 {
			proto := swarm.PortConfigProtocolTCP
			if strings.EqualFold(p.Protocol, "udp") {
				proto = swarm.PortConfigProtocolUDP
			}
			portConfigs = append(portConfigs, swarm.PortConfig{
				Protocol:      proto,
				TargetPort:    uint32(p.ContainerPort),
				PublishedPort: uint32(p.HostPort),
				PublishMode:   swarm.PortConfigPublishModeIngress,
			})
		}
	}

	swarmSpec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name:   formatServiceName(spec.ProjectName, spec.Name),
			Labels: labels,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: containerSpec,
			Resources:     buildResourceLimits(spec.Resources),
			RestartPolicy: buildRestartPolicy(spec.RestartPolicy),
			Networks:      networks,
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
		UpdateConfig: &swarm.UpdateConfig{
			Parallelism:   1,
			Delay:         5 * time.Second,
			FailureAction: swarm.UpdateFailureActionPause,
			Order:         order,
		},
		EndpointSpec: &swarm.EndpointSpec{
			Mode:  swarm.ResolutionModeDNSRR,
			Ports: portConfigs,
		},
	}

	return swarmSpec, nil
}
