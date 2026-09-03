package domain_test

import (
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func TestProjectValidation(t *testing.T) {
	tests := []struct {
		name    string
		project domain.Project
		wantErr bool
	}{
		{
			name: "valid project",
			project: domain.Project{
				ID:          "proj-1",
				Name:        "My Project",
				Description: "Test project",
			},
			wantErr: false,
		},
		{
			name: "missing id",
			project: domain.Project{
				Name: "My Project",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			project: domain.Project{
				ID: "proj-1",
			},
			wantErr: true,
		},
		{
			name: "empty name with spaces",
			project: domain.Project{
				ID:   "proj-1",
				Name: "   ",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceValidationAndToSpec(t *testing.T) {
	tests := []struct {
		name    string
		service domain.Service
		wantErr bool
	}{
		{
			name: "valid service",
			service: domain.Service{
				ID:        "srv-1",
				ProjectID: "proj-1",
				Name:      "web",
				Type:      domain.ServiceTypeApp,
				Image:     "nginx:alpine",
				Replicas:  1,
			},
			wantErr: false,
		},
		{
			name: "missing id",
			service: domain.Service{
				ProjectID: "proj-1",
				Name:      "web",
				Image:     "nginx:alpine",
			},
			wantErr: true,
		},
		{
			name: "missing project id",
			service: domain.Service{
				ID:    "srv-1",
				Name:  "web",
				Image: "nginx:alpine",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			service: domain.Service{
				ID:        "srv-1",
				ProjectID: "proj-1",
				Image:     "nginx:alpine",
			},
			wantErr: true,
		},
		{
			name: "missing image",
			service: domain.Service{
				ID:        "srv-1",
				ProjectID: "proj-1",
				Name:      "web",
			},
			wantErr: true,
		},
		{
			name: "negative replicas",
			service: domain.Service{
				ID:        "srv-1",
				ProjectID: "proj-1",
				Name:      "web",
				Image:     "nginx:alpine",
				Replicas:  -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}

	// Test ToSpec
	validService := domain.Service{
		ID:        "srv-1",
		ProjectID: "proj-1",
		Name:      "web",
		Type:      domain.ServiceTypeApp,
		Image:     "nginx:alpine",
		Replicas:  0, // should default to 1 in ToSpec
		Ports: []domain.PortMapping{
			{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
		},
	}
	spec := validService.ToSpec()
	if spec.ID != "srv-1" || spec.Replicas != 1 {
		t.Errorf("ToSpec() unexpected spec = %+v", spec)
	}
	if spec.Labels["easypanel.project"] != "proj-1" {
		t.Errorf("ToSpec() missing or wrong label: %v", spec.Labels)
	}
}

func TestDomainValidationAndRoute(t *testing.T) {
	d := domain.Domain{
		ID:         "dom-1",
		ServiceID:  "srv-1",
		DomainName: "example.com",
		Port:       80,
		HTTPS:      true,
		CertMode:   "letsencrypt",
	}

	if err := d.Validate(); err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	route := d.ToRouteConfig()
	if route.Domain != "example.com" || route.TargetPort != 80 || route.CertResolver != "letsencrypt" {
		t.Errorf("ToRouteConfig() unexpected route = %+v", route)
	}

	// Test invalid port
	d.Port = 999999
	if err := d.Validate(); err == nil {
		t.Error("expected error for invalid port > 65535, got nil")
	}
}

func TestDeploymentValidation(t *testing.T) {
	now := time.Now()
	dep := domain.Deployment{
		ID:        "dep-1",
		ServiceID: "srv-1",
		Status:    domain.DeploymentStatusRunning,
		StartedAt: now,
	}

	if err := dep.Validate(); err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	dep.ID = ""
	if err := dep.Validate(); err == nil {
		t.Error("expected validation error for empty ID, got nil")
	}
}

func TestUserValidation(t *testing.T) {
	u := domain.User{
		ID:           "usr-1",
		Email:        "admin@example.com",
		PasswordHash: "argon2id$hashed",
		Role:         "admin",
	}

	if err := u.Validate(); err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	// Invalid email
	u.Email = "invalid-email"
	if err := u.Validate(); err == nil {
		t.Error("expected validation error for invalid email, got nil")
	}
}

func TestNewID(t *testing.T) {
	id1 := domain.NewID()
	id2 := domain.NewID()

	if len(id1) != 24 {
		t.Errorf("expected 24 character hex ID, got %d characters: %s", len(id1), id1)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, got duplicate: %s", id1)
	}
}

func TestEnvVarConversions(t *testing.T) {
	vars := []domain.EnvVar{
		{Key: "PORT", Value: "8080"},
		{Key: "NODE_ENV", Value: "production"},
	}

	// ToSlice
	slice := domain.EnvVarsToSlice(vars)
	if len(slice) != 2 || slice[0] != "PORT=8080" || slice[1] != "NODE_ENV=production" {
		t.Errorf("EnvVarsToSlice unexpected: %v", slice)
	}

	// FromSlice
	rawLines := []string{"# comment", "PORT=8080", "", "EMPTY=", "KEY_ONLY", "NODE_ENV=production"}
	parsed := domain.EnvVarsFromSlice(rawLines)
	if len(parsed) != 4 {
		t.Fatalf("expected 4 parsed env vars, got %d: %v", len(parsed), parsed)
	}
	if parsed[0].Key != "PORT" || parsed[0].Value != "8080" {
		t.Errorf("parsed[0] unexpected: %+v", parsed[0])
	}
	if parsed[1].Key != "EMPTY" || parsed[1].Value != "" {
		t.Errorf("parsed[1] unexpected: %+v", parsed[1])
	}

	// ToMap & FromMap
	m := domain.EnvVarsToMap(vars)
	if m["PORT"] != "8080" || m["NODE_ENV"] != "production" {
		t.Errorf("EnvVarsToMap unexpected: %v", m)
	}

	roundtrip := domain.EnvVarsFromMap(m)
	if len(roundtrip) != 2 {
		t.Errorf("EnvVarsFromMap unexpected length: %d", len(roundtrip))
	}
}

func TestPortMappingParsingAndFormatting(t *testing.T) {
	tests := []struct {
		input       string
		wantHost    int
		wantCont    int
		wantProto   string
		wantErr     bool
		wantString  string
	}{
		{"8080:80", 8080, 80, "tcp", false, "8080:80/tcp"},
		{"3000:3000/tcp", 3000, 3000, "tcp", false, "3000:3000/tcp"},
		{"53:53/udp", 53, 53, "udp", false, "53:53/udp"},
		{"invalid", 0, 0, "", true, ""},
		{"-1:80", 0, 0, "", true, ""},
		{"8080:70000", 0, 0, "", true, ""},
		{"8080:80/sctp", 0, 0, "", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p, err := domain.ParsePortMapping(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParsePortMapping(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr {
				if p.HostPort != tt.wantHost || p.ContainerPort != tt.wantCont || p.Protocol != tt.wantProto {
					t.Errorf("ParsePortMapping(%q) = %+v, want host=%d cont=%d proto=%s", tt.input, p, tt.wantHost, tt.wantCont, tt.wantProto)
				}
				if p.String() != tt.wantString {
					t.Errorf("PortMapping.String() = %q, want %q", p.String(), tt.wantString)
				}
			}
		})
	}
}

func TestVolumeMountParsingAndFormatting(t *testing.T) {
	tests := []struct {
		input       string
		wantName    string
		wantHost    string
		wantCont    string
		wantRO      bool
		wantErr     bool
		wantString  string
	}{
		{"myvol:/var/lib/data", "myvol", "", "/var/lib/data", false, false, "myvol:/var/lib/data:rw"},
		{"myvol:/var/lib/data:ro", "myvol", "", "/var/lib/data", true, false, "myvol:/var/lib/data:ro"},
		{"/opt/data:/app/data:rw", "", "/opt/data", "/app/data", false, false, "/opt/data:/app/data:rw"},
		{"invalid-no-colon", "", "", "", false, true, ""},
		{":/var/data", "", "", "", false, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			vm, err := domain.ParseVolumeMount(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseVolumeMount(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr {
				if vm.Name != tt.wantName || vm.HostPath != tt.wantHost || vm.ContainerPath != tt.wantCont || vm.ReadOnly != tt.wantRO {
					t.Errorf("ParseVolumeMount(%q) = %+v", tt.input, vm)
				}
				if vm.String() != tt.wantString {
					t.Errorf("VolumeMount.String() = %q, want %q", vm.String(), tt.wantString)
				}
			}
		})
	}
}

func TestServiceStateTransitions(t *testing.T) {
	srv := domain.Service{
		Status: domain.ServiceStatusStopped,
	}

	if !srv.CanTransitionTo(domain.ServiceStatusStarting) {
		t.Error("stopped should be able to transition to starting")
	}
	if !srv.CanTransitionTo(domain.ServiceStatusDeploying) {
		t.Error("stopped should be able to transition to deploying")
	}
	if srv.CanTransitionTo(domain.ServiceStatusRunning) {
		t.Error("stopped should not directly transition to running")
	}
	if srv.IsRunning() {
		t.Error("stopped service should not report IsRunning == true")
	}

	srv.Status = domain.ServiceStatusRunning
	if !srv.IsRunning() {
		t.Error("running service should report IsRunning == true")
	}
}

func TestDeploymentLifecycle(t *testing.T) {
	dep := domain.Deployment{
		ID:        "dep-lifecycle",
		ServiceID: "srv-1",
		Status:    domain.DeploymentStatusQueued,
		StartedAt: time.Now(),
	}

	dep.Complete(domain.DeploymentStatusRunning, "Build successful. Container running.")
	if dep.Status != domain.DeploymentStatusRunning || dep.Logs == "" || dep.FinishedAt == nil {
		t.Errorf("Complete() did not properly update deployment: %+v", dep)
	}

	dep.Fail("Build failed with exit code 1")
	if dep.Status != domain.DeploymentStatusFailed {
		t.Errorf("Fail() did not set status to failed: %v", dep.Status)
	}
}

func TestSecretEnvVarMasking(t *testing.T) {
	vPlain := domain.EnvVar{Key: "PORT", Value: "8080", IsSecret: false}
	if vPlain.MaskedValue() != "8080" {
		t.Errorf("plain env var should not be masked: %s", vPlain.MaskedValue())
	}

	vSecret := domain.EnvVar{Key: "DB_PASS", Value: "supersecret", IsSecret: true}
	if vSecret.MaskedValue() != "••••••••" {
		t.Errorf("secret env var should be masked: %s", vSecret.MaskedValue())
	}
}

func TestServiceSourceTypeValidation(t *testing.T) {
	// Git service without repoURL should fail
	sGit := domain.Service{
		ID:         "s-git",
		ProjectID:  "p-1",
		Name:       "git-app",
		Type:       domain.ServiceTypeApp,
		SourceType: domain.SourceTypeGit,
	}
	if err := sGit.Validate(); err != domain.ErrValidation {
		t.Errorf("expected ErrValidation for git service without RepoURL, got %v", err)
	}

	// Git service with valid repoURL should pass
	sGit.SourceConfig = &domain.SourceConfig{
		RepoURL: "https://github.com/example/app.git",
		Branch:  "main",
	}
	if err := sGit.Validate(); err != nil {
		t.Errorf("expected valid git service, got err %v", err)
	}

	// Spec conversion should set default restart policy and merge labels
	spec := sGit.ToSpec()
	if spec.RestartPolicy != domain.RestartPolicyUnlessStopped {
		t.Errorf("expected default restart policy 'unless-stopped', got %s", spec.RestartPolicy)
	}
	if spec.Labels["easypanel.service"] != "s-git" {
		t.Errorf("expected easypanel.service label, got %+v", spec.Labels)
	}
}
