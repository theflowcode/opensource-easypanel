package docker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func TestFormatEnvVars(t *testing.T) {
	vars := []domain.EnvVar{
		{Key: "PORT", Value: "8080"},
		{Key: "DEBUG", Value: "true"},
		{Key: "SECRET", Value: "supersecret", IsSecret: true},
	}

	formatted := formatEnvVars(vars)
	if len(formatted) != 3 {
		t.Fatalf("expected 3 env vars, got %d", len(formatted))
	}
	if formatted[0] != "PORT=8080" || formatted[1] != "DEBUG=true" || formatted[2] != "SECRET=supersecret" {
		t.Errorf("unexpected formatted env vars: %v", formatted)
	}
}

func TestBuildRestartPolicy(t *testing.T) {
	tests := []struct {
		policy   string
		expected swarm.RestartPolicyCondition
	}{
		{domain.RestartPolicyAlways, swarm.RestartPolicyConditionAny},
		{domain.RestartPolicyUnlessStopped, swarm.RestartPolicyConditionAny},
		{domain.RestartPolicyOnFailure, swarm.RestartPolicyConditionOnFailure},
		{domain.RestartPolicyNo, swarm.RestartPolicyConditionNone},
		{"unknown", swarm.RestartPolicyConditionAny},
	}

	for _, tc := range tests {
		rp := buildRestartPolicy(tc.policy)
		if rp.Condition != tc.expected {
			t.Errorf("policy %q expected condition %v, got %v", tc.policy, tc.expected, rp.Condition)
		}
	}
}

func TestBuildResourceLimits(t *testing.T) {
	// Zero resources
	res := buildResourceLimits(domain.ResourceLimits{})
	if res != nil {
		t.Errorf("expected nil requirements for zero limits, got %v", res)
	}

	// Non-zero resources
	res = buildResourceLimits(domain.ResourceLimits{
		CPULimit:    1.5,
		MemoryLimit: 512,
	})
	if res == nil || res.Limits == nil {
		t.Fatalf("expected non-nil resource limits")
	}
	if res.Limits.NanoCPUs != 1500000000 {
		t.Errorf("expected 1.5e9 NanoCPUs, got %d", res.Limits.NanoCPUs)
	}
	if res.Limits.MemoryBytes != 512*1024*1024 {
		t.Errorf("expected 512MB in bytes, got %d", res.Limits.MemoryBytes)
	}
}

func TestBuildSwarmServiceSpec_Complete(t *testing.T) {
	tempDir := t.TempDir()

	spec := domain.ServiceSpec{
		ID:          "svc-123",
		ProjectName: "demo",
		Name:        "api",
		Type:        domain.ServiceTypeApp,
		Image:       "node:20-alpine",
		Command:     "npm start",
		Args:        []string{"--production"},
		Replicas:    2,
		ZeroDowntime: true,
		EnvVars: []domain.EnvVar{
			{Key: "NODE_ENV", Value: "production"},
		},
		Ports: []domain.PortMapping{
			{HostPort: 8080, ContainerPort: 3000, Protocol: "tcp"},
		},
		Volumes: []domain.VolumeMount{
			{Type: "file", Name: "config.json", ContainerPath: "/app/config.json", Content: `{"ok":true}`},
			{Type: "volume", Name: "uploads", ContainerPath: "/app/uploads"},
		},
		Resources: domain.ResourceLimits{
			CPULimit:    0.5,
			MemoryLimit: 256,
		},
		HealthCheck: &domain.HealthCheckConfig{
			Test:            []string{"CMD", "curl", "-f", "http://localhost:3000/health"},
			IntervalSeconds: 10,
			TimeoutSeconds:  3,
			Retries:         3,
		},
	}

	swarmSpec, err := buildSwarmServiceSpec(spec, "easypanel", tempDir)
	if err != nil {
		t.Fatalf("buildSwarmServiceSpec failed: %v", err)
	}

	if swarmSpec.Annotations.Name != "demo_api" {
		t.Errorf("expected service name 'demo_api', got %s", swarmSpec.Annotations.Name)
	}
	if *swarmSpec.Mode.Replicated.Replicas != 2 {
		t.Errorf("expected 2 replicas, got %d", *swarmSpec.Mode.Replicated.Replicas)
	}
	if swarmSpec.UpdateConfig.Order != swarm.UpdateOrderStartFirst {
		t.Errorf("expected start-first update order for zero-downtime")
	}

	taskSpec := swarmSpec.TaskTemplate
	if taskSpec.ContainerSpec.Image != "node:20-alpine" {
		t.Errorf("expected image node:20-alpine, got %s", taskSpec.ContainerSpec.Image)
	}

	// Verify file mount was written
	filePath := filepath.Join(tempDir, "demo", "api", "config.json")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("expected written config.json file: %v", err)
	}
	if string(content) != `{"ok":true}` {
		t.Errorf("expected file content '{\"ok\":true}', got %s", string(content))
	}

	// Verify mounts
	if len(taskSpec.ContainerSpec.Mounts) != 2 {
		t.Fatalf("expected 2 mounts, got %d", len(taskSpec.ContainerSpec.Mounts))
	}
	if taskSpec.ContainerSpec.Mounts[0].Type != mount.TypeBind {
		t.Errorf("expected bind mount for file type, got %s", taskSpec.ContainerSpec.Mounts[0].Type)
	}
	if taskSpec.ContainerSpec.Mounts[1].Type != mount.TypeVolume {
		t.Errorf("expected volume mount, got %s", taskSpec.ContainerSpec.Mounts[1].Type)
	}
}
