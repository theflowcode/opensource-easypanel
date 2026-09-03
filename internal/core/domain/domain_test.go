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
		ID:           "srv-1",
		ProjectID:    "proj-1",
		ProjectName:  "myfirstproject",
		Name:         "web",
		Type:         domain.ServiceTypeApp,
		Image:        "nginx:alpine",
		DeployScript: "bundle exec rails db:migrate",
		Replicas:     0, // should default to 1 in ToSpec
		Ports: []domain.PortMapping{
			{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
		},
	}
	spec := validService.ToSpec()
	if spec.ID != "srv-1" || spec.Replicas != 1 {
		t.Errorf("ToSpec() unexpected spec = %+v", spec)
	}
	if spec.ProjectName != "myfirstproject" {
		t.Errorf("ToSpec() expected ProjectName myfirstproject, got %s", spec.ProjectName)
	}
	if spec.DeployScript != "bundle exec rails db:migrate" {
		t.Errorf("ToSpec() expected DeployScript, got %s", spec.DeployScript)
	}
	if len(spec.NetworkAliases) != 1 || spec.NetworkAliases[0] != "web" {
		t.Errorf("ToSpec() expected NetworkAliases ['web'], got %v", spec.NetworkAliases)
	}
	if spec.Labels["easypanel.project"] != "proj-1" || spec.Labels["easypanel.projectName"] != "myfirstproject" {
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

func TestServiceDeployTokenAndCronJobs(t *testing.T) {
	s := domain.Service{
		ID:        "s-cron",
		ProjectID: "p-1",
		Name:      "cron-app",
		Image:     "alpine:latest",
		CronJobs: []domain.CronJobSpec{
			{ID: "c-1", Name: "daily", Schedule: "@daily", Command: "echo hi"},
		},
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if s.DeployToken == "" {
		t.Error("expected auto-generated DeployToken")
	}

	spec := s.ToSpec()
	if spec.DeployToken != s.DeployToken {
		t.Errorf("expected DeployToken to match in spec: %s != %s", spec.DeployToken, s.DeployToken)
	}
	if len(spec.CronJobs) != 1 || spec.CronJobs[0].Schedule != "@daily" {
		t.Errorf("CronJobs not properly preserved in spec: %+v", spec.CronJobs)
	}
}

func TestBackupValidation(t *testing.T) {
	b := domain.Backup{}
	if err := b.Validate(); err != domain.ErrValidation {
		t.Errorf("expected ErrValidation for empty backup, got %v", err)
	}

	b = domain.Backup{
		ID:        "b-1",
		ServiceID: "s-1",
		FileName:  "db.sql",
	}
	if err := b.Validate(); err != nil {
		t.Errorf("expected valid backup, got %v", err)
	}
}

func TestSessionValidationAndExpiration(t *testing.T) {
	s := domain.Session{}
	if err := s.Validate(); err != domain.ErrValidation {
		t.Errorf("expected ErrValidation for empty session, got %v", err)
	}

	now := time.Now().UTC()
	s = domain.Session{
		ID:        "sess-1",
		UserID:    "u-1",
		TokenHash: "hash123",
		ExpiresAt: now.Add(1 * time.Hour),
		CreatedAt: now,
	}
	if err := s.Validate(); err != nil {
		t.Errorf("expected valid session, got %v", err)
	}
	if s.IsExpired() {
		t.Error("session with future expiry should not be expired")
	}

	s.ExpiresAt = now.Add(-1 * time.Minute)
	if !s.IsExpired() {
		t.Error("session with past expiry should be expired")
	}
}

func TestProductionUIParityAndDatabaseEngines(t *testing.T) {
	// 1. Test IsDatabase helper
	if !domain.ServiceTypePostgres.IsDatabase() || !domain.ServiceTypeRedis.IsDatabase() ||
		!domain.ServiceTypeMySQL.IsDatabase() || !domain.ServiceTypeMariaDB.IsDatabase() ||
		!domain.ServiceTypeMongoDB.IsDatabase() || !domain.ServiceTypeDatabase.IsDatabase() {
		t.Error("expected IsDatabase to be true for database engines")
	}
	if domain.ServiceTypeApp.IsDatabase() || domain.ServiceTypeTemplate.IsDatabase() {
		t.Error("expected IsDatabase to be false for app and template")
	}

	// 2. Test auto image defaulting for database engines
	dbService := domain.Service{
		ID:        "db-1",
		ProjectID: "proj-1",
		Name:      "chatwoot-db",
		Type:      domain.ServiceTypePostgres,
	}
	if err := dbService.Validate(); err != nil {
		t.Fatalf("Validate failed on postgres service: %v", err)
	}
	if dbService.Image != "postgres:16-alpine" {
		t.Errorf("expected default postgres image, got %s", dbService.Image)
	}

	// 3. Test DatabaseConfig and Redirects in ToSpec
	dbService.DatabaseConfig = &domain.DatabaseConfig{
		DatabaseName: "chatwoot_production",
		RootPassword: "supersecretpass",
		ExposePort:   5432,
		IsExposed:    true,
		EnabledTools: []string{"pgweb", "dbgate"},
		InternalURL:  "postgres://postgres:supersecretpass@chatwoot-db:5432/chatwoot_production",
	}
	dbService.Redirects = []domain.RedirectRule{
		{
			ID:        "red-1",
			ServiceID: "db-1",
			Source:    "www.example.com",
			Target:    "https://example.com",
			Permanent: true,
			Enabled:   true,
		},
	}
	spec := dbService.ToSpec()
	if spec.DatabaseConfig == nil || spec.DatabaseConfig.ExposePort != 5432 || len(spec.DatabaseConfig.EnabledTools) != 2 {
		t.Errorf("DatabaseConfig not properly preserved in spec: %+v", spec.DatabaseConfig)
	}
	if len(spec.Redirects) != 1 || !spec.Redirects[0].Permanent {
		t.Errorf("Redirects not properly preserved in spec: %+v", spec.Redirects)
	}
}

func TestActionsStorageProvidersAndMacros(t *testing.T) {
	// 1. Action validation
	a := domain.Action{}
	if err := a.Validate(); err != domain.ErrValidation {
		t.Errorf("expected ErrValidation on empty action, got %v", err)
	}
	a.ID = "act-1"
	a.Type = domain.ActionTypeDeployment
	if err := a.Validate(); err != nil {
		t.Fatalf("Validate failed on valid action: %v", err)
	}
	if a.Status != domain.ActionStatusPending {
		t.Errorf("expected default pending status, got %s", a.Status)
	}

	// 2. StorageProvider validation
	sp := domain.StorageProvider{}
	if err := sp.Validate(); err != domain.ErrValidation {
		t.Errorf("expected ErrValidation on empty storage provider, got %v", err)
	}
	sp.ID = "sp-1"
	sp.Name = "Local Disk"
	if err := sp.Validate(); err != nil {
		t.Fatalf("Validate failed on valid local storage provider: %v", err)
	}
	if sp.Type != domain.StorageProviderTypeLocal || sp.Path != "/etc/easypanel/backups" {
		t.Errorf("unexpected local storage defaults: %+v", sp)
	}

	spS3 := domain.StorageProvider{
		ID:   "sp-2",
		Name: "S3 Remote",
		Type: domain.StorageProviderTypeS3,
	}
	if err := spS3.Validate(); err != domain.ErrValidation {
		t.Error("expected error on S3 provider without endpoint and bucket")
	}
	spS3.Endpoint = "https://s3.amazonaws.com"
	spS3.Bucket = "my-backups"
	if err := spS3.Validate(); err != nil {
		t.Fatalf("Validate failed on valid S3 provider: %v", err)
	}

	// 3. Macro expansion
	envVars := []domain.EnvVar{
		{Key: "FRONTEND_URL", Value: "https://$(PRIMARY_DOMAIN)"},
		{Key: "POSTGRES_HOST", Value: "$(PROJECT_NAME)_chatwoot-db"},
	}
	expanded := domain.ExpandEnvVars(envVars, map[string]string{
		"PRIMARY_DOMAIN": "chatwoot.example.com",
		"PROJECT_NAME":   "myfirstproject",
	})
	if expanded[0].Value != "https://chatwoot.example.com" {
		t.Errorf("macro expansion failed for PRIMARY_DOMAIN: %s", expanded[0].Value)
	}
	if expanded[1].Value != "myfirstproject_chatwoot-db" {
		t.Errorf("macro expansion failed for PROJECT_NAME: %s", expanded[1].Value)
	}

	// 4. VolumeMount type detection
	vmVol, err := domain.ParseVolumeMount("myvol:/app/storage")
	if err != nil || vmVol.Type != "volume" || vmVol.Name != "myvol" {
		t.Errorf("expected volume mount, got %+v", vmVol)
	}
	vmBind, err := domain.ParseVolumeMount("/etc/data:/app/data")
	if err != nil || vmBind.Type != "bind" || vmBind.HostPath != "/etc/data" {
		t.Errorf("expected bind mount, got %+v", vmBind)
	}

	// 5. PrimaryDomainID and ZeroDowntime in Service ToSpec
	srv := domain.Service{
		ID:              "s-zd",
		ProjectID:       "p-1",
		Name:            "web",
		Image:           "nginx:alpine",
		PrimaryDomainID: "dom-123",
		ZeroDowntime:    true,
	}
	spec := srv.ToSpec()
	if spec.PrimaryDomainID != "dom-123" || !spec.ZeroDowntime {
		t.Errorf("ToSpec did not propagate PrimaryDomainID or ZeroDowntime: %+v", spec)
	}
}
