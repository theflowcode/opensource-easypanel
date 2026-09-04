package docker

import (
	"testing"

	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

func TestDockerAdapter_InterfaceCompliance(t *testing.T) {
	var _ port.DockerPort = (*DockerAdapter)(nil)
}

func TestDockerAdapter_Options(t *testing.T) {
	customDir := "/custom/projects"
	customNet := "custom-network"

	adapter := &DockerAdapter{
		projectsDir:    "/default/projects",
		defaultNetwork: "default-net",
	}

	WithProjectsDir(customDir)(adapter)
	WithDefaultNetwork(customNet)(adapter)

	if adapter.ProjectsDir() != customDir {
		t.Errorf("expected projectsDir %s, got %s", customDir, adapter.ProjectsDir())
	}
	if adapter.DefaultNetwork() != customNet {
		t.Errorf("expected defaultNetwork %s, got %s", customNet, adapter.DefaultNetwork())
	}
}

func TestFormatServiceName(t *testing.T) {
	tests := []struct {
		project  string
		service  string
		expected string
	}{
		{"myproject", "web", "myproject_web"},
		{"", "standalone", "standalone"},
		{"prod", "api-gateway", "prod_api-gateway"},
	}

	for _, tc := range tests {
		got := formatServiceName(tc.project, tc.service)
		if got != tc.expected {
			t.Errorf("formatServiceName(%q, %q) = %q; want %q", tc.project, tc.service, got, tc.expected)
		}
	}
}
