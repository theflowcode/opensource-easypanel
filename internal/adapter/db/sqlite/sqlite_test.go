package sqlite_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/adapter/db/sqlite"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

func setupTestDB(t *testing.T) *sqlite.Repository {
	t.Helper()
	// Use isolated in-memory SQLite database for testing
	dsn := fmt.Sprintf("file:testdb_%d?mode=memory&cache=shared", time.Now().UnixNano())
	repo, err := sqlite.New(dsn)
	if err != nil {
		t.Fatalf("failed to create sqlite repo: %v", err)
	}

	ctx := context.Background()
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	t.Cleanup(func() {
		_ = repo.Close()
	})

	return repo
}

func TestMigrateIdempotency(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Running migration a second time should not fail
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("second migration should be idempotent, got: %v", err)
	}
}

func TestProjectCRUD(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	p := &domain.Project{
		ID:          "p-1",
		Name:        "Production",
		Description: "Main production workloads",
	}

	// Create
	if err := repo.CreateProject(ctx, p); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Duplicate ID or Name should fail
	duplicate := &domain.Project{
		ID:   "p-2",
		Name: "Production",
	}
	if err := repo.CreateProject(ctx, duplicate); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	// Get by ID
	got, err := repo.GetProject(ctx, "p-1")
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if got.Name != p.Name || got.Description != p.Description {
		t.Errorf("GetProject mismatch: got %+v, want %+v", got, p)
	}

	// Get by Name
	gotName, err := repo.GetProjectByName(ctx, "Production")
	if err != nil {
		t.Fatalf("GetProjectByName failed: %v", err)
	}
	if gotName.ID != p.ID {
		t.Errorf("GetProjectByName mismatch: got ID %s, want %s", gotName.ID, p.ID)
	}

	// List
	list, err := repo.ListProjects(ctx)
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 project, got %d", len(list))
	}

	// Update
	p.Description = "Updated description"
	if err := repo.UpdateProject(ctx, p); err != nil {
		t.Fatalf("UpdateProject failed: %v", err)
	}
	updated, _ := repo.GetProject(ctx, "p-1")
	if updated.Description != "Updated description" {
		t.Errorf("UpdateProject description not updated: %s", updated.Description)
	}

	// Delete
	if err := repo.DeleteProject(ctx, "p-1"); err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
	if _, err := repo.GetProject(ctx, "p-1"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestServiceCRUDAndCascades(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Create parent project
	proj := &domain.Project{ID: "p-parent", Name: "Parent Project"}
	if err := repo.CreateProject(ctx, proj); err != nil {
		t.Fatalf("failed to create parent project: %v", err)
	}

	srv := &domain.Service{
		ID:        "s-1",
		ProjectID: "p-parent",
		Name:      "web-api",
		Type:      domain.ServiceTypeApp,
		Image:     "golang:1.24-alpine",
		Command:   "./server",
		Args:      []string{"--port", "8080"},
		EnvVars: []domain.EnvVar{
			{Key: "ENV", Value: "production"},
		},
		Ports: []domain.PortMapping{
			{HostPort: 8080, ContainerPort: 8080, Protocol: "tcp"},
		},
		Volumes: []domain.VolumeMount{
			{Name: "data-vol", ContainerPath: "/data", ReadOnly: false},
		},
		Domains:   []string{"api.example.com"},
		Replicas:  2,
		Resources: domain.ResourceLimits{CPULimit: 1.5, MemoryLimit: 512},
		Status:    domain.ServiceStatusRunning,
	}

	// Create
	if err := repo.CreateService(ctx, srv); err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	// Get
	got, err := repo.GetService(ctx, "s-1")
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}
	if got.Name != srv.Name || got.Image != srv.Image || len(got.Ports) != 1 || got.Replicas != 2 {
		t.Errorf("GetService fields mismatch: %+v", got)
	}

	// Get by Name in project
	gotByName, err := repo.GetServiceByName(ctx, "p-parent", "web-api")
	if err != nil {
		t.Fatalf("GetServiceByName failed: %v", err)
	}
	if gotByName.ID != "s-1" {
		t.Errorf("expected ID s-1, got %s", gotByName.ID)
	}

	// List by project
	list, err := repo.ListServicesByProject(ctx, "p-parent")
	if err != nil || len(list) != 1 {
		t.Fatalf("ListServicesByProject failed: len=%d, err=%v", len(list), err)
	}

	// Add Domain and Deployment to verify cascading deletes
	dom := &domain.Domain{
		ID:         "d-1",
		ServiceID:  "s-1",
		DomainName: "api.example.com",
		Port:       8080,
		HTTPS:      true,
	}
	if err := repo.CreateDomain(ctx, dom); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	dep := &domain.Deployment{
		ID:        "dep-1",
		ServiceID: "s-1",
		Status:    domain.DeploymentStatusRunning,
		Trigger:   "manual",
	}
	if err := repo.CreateDeployment(ctx, dep); err != nil {
		t.Fatalf("CreateDeployment failed: %v", err)
	}

	// Delete Project should cascade delete service, domain, and deployment
	if err := repo.DeleteProject(ctx, "p-parent"); err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}

	if _, err := repo.GetService(ctx, "s-1"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected service to be cascade deleted, got %v", err)
	}
	if _, err := repo.GetDomain(ctx, "d-1"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected domain to be cascade deleted, got %v", err)
	}
	if _, err := repo.GetDeployment(ctx, "dep-1"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected deployment to be cascade deleted, got %v", err)
	}
}

func TestUsersAndSettings(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Users
	user := &domain.User{
		ID:           "u-1",
		Email:        "admin@easypanel.local",
		PasswordHash: "secret-hash",
		Role:         "admin",
	}

	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	gotUser, err := repo.GetUserByEmail(ctx, "admin@easypanel.local")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if gotUser.ID != "u-1" || gotUser.Role != "admin" {
		t.Errorf("GetUserByEmail unexpected user: %+v", gotUser)
	}

	// Settings
	if err := repo.SetSetting(ctx, "panel.name", "OpenSource Easypanel"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	val, err := repo.GetSetting(ctx, "panel.name")
	if err != nil || val != "OpenSource Easypanel" {
		t.Fatalf("GetSetting unexpected: val=%s, err=%v", val, err)
	}

	// Upsert setting
	if err := repo.SetSetting(ctx, "panel.name", "OpenSource Easypanel v2"); err != nil {
		t.Fatalf("SetSetting update failed: %v", err)
	}
	val, _ = repo.GetSetting(ctx, "panel.name")
	if val != "OpenSource Easypanel v2" {
		t.Errorf("expected updated setting, got: %s", val)
	}
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Create base project
	proj := &domain.Project{ID: "p-concurrent", Name: "Concurrent Project"}
	if err := repo.CreateProject(ctx, proj); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	var wg sync.WaitGroup
	workers := 10

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			s := &domain.Service{
				ID:        fmt.Sprintf("s-concurrent-%d", idx),
				ProjectID: "p-concurrent",
				Name:      fmt.Sprintf("service-%d", idx),
				Type:      domain.ServiceTypeApp,
				Image:     "alpine:latest",
				Status:    domain.ServiceStatusRunning,
			}
			_ = repo.CreateService(ctx, s)
			_, _ = repo.GetService(ctx, s.ID)
			_, _ = repo.ListServicesByProject(ctx, "p-concurrent")
		}(i)
	}

	wg.Wait()

	services, err := repo.ListServicesByProject(ctx, "p-concurrent")
	if err != nil {
		t.Fatalf("ListServicesByProject failed after concurrent writes: %v", err)
	}
	if len(services) != workers {
		t.Errorf("expected %d services, got %d", workers, len(services))
	}
}

func TestWithTxCommit(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.WithTx(ctx, func(tx port.DatabasePort) error {
		p := &domain.Project{ID: "p-tx-1", Name: "Tx Project"}
		if err := tx.CreateProject(ctx, p); err != nil {
			return err
		}
		s := &domain.Service{
			ID:        "s-tx-1",
			ProjectID: "p-tx-1",
			Name:      "tx-service",
			Type:      domain.ServiceTypeApp,
			Image:     "alpine",
		}
		return tx.CreateService(ctx, s)
	})
	if err != nil {
		t.Fatalf("WithTx commit failed: %v", err)
	}

	// Verify both were committed
	if _, err := repo.GetProject(ctx, "p-tx-1"); err != nil {
		t.Errorf("expected project to be committed: %v", err)
	}
	if _, err := repo.GetService(ctx, "s-tx-1"); err != nil {
		t.Errorf("expected service to be committed: %v", err)
	}
}

func TestWithTxRollback(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	errExpected := errors.New("simulated failure")

	err := repo.WithTx(ctx, func(tx port.DatabasePort) error {
		p := &domain.Project{ID: "p-tx-rollback", Name: "Rollback Project"}
		if err := tx.CreateProject(ctx, p); err != nil {
			return err
		}
		return errExpected
	})
	if !errors.Is(err, errExpected) {
		t.Fatalf("expected errExpected, got %v", err)
	}

	// Verify project was rolled back
	if _, err := repo.GetProject(ctx, "p-tx-rollback"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected project to be rolled back and not found, got %v", err)
	}
}

func TestAdvancedServiceAndDomainFields(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	p := &domain.Project{ID: "p-adv", Name: "Adv Project"}
	if err := repo.CreateProject(ctx, p); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	srv := &domain.Service{
		ID:           "s-adv",
		ProjectID:    "p-adv",
		ProjectName:  "adv-project",
		Name:         "adv-web",
		Type:         domain.ServiceTypeApp,
		DeployScript: "npm run migrate",
		SourceType:   domain.SourceTypeGit,
		SourceConfig: &domain.SourceConfig{
			RepoURL: "https://github.com/example/repo.git",
			Branch:  "main",
		},
		RestartPolicy: domain.RestartPolicyUnlessStopped,
		HealthCheck: &domain.HealthCheckConfig{
			Test:            []string{"CMD", "curl", "-f", "http://localhost:80/health"},
			IntervalSeconds: 30,
			TimeoutSeconds:  5,
			Retries:         3,
		},
		Labels: map[string]string{
			"traefik.enable": "true",
		},
		EnvVars: []domain.EnvVar{
			{Key: "APP_ENV", Value: "production"},
			{Key: "DB_PASS", Value: "secret123", IsSecret: true},
		},
		DatabaseConfig: &domain.DatabaseConfig{
			DatabaseName: "adv_db",
			RootPassword: "rootpassword",
			ExposePort:   5432,
			IsExposed:    true,
			EnabledTools: []string{"pgweb", "dbgate"},
		},
		Redirects: []domain.RedirectRule{
			{
				ID:        "r-1",
				ServiceID: "s-adv",
				Source:    "old.example.com",
				Target:    "https://new.example.com",
				Permanent: true,
				Enabled:   true,
			},
		},
	}

	if err := repo.CreateService(ctx, srv); err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	got, err := repo.GetService(ctx, "s-adv")
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}
	if got.ProjectName != "adv-project" {
		t.Errorf("ProjectName not preserved: expected adv-project, got %s", got.ProjectName)
	}
	if got.DeployScript != "npm run migrate" {
		t.Errorf("DeployScript not preserved: expected npm run migrate, got %s", got.DeployScript)
	}
	if got.DatabaseConfig == nil || got.DatabaseConfig.ExposePort != 5432 || len(got.DatabaseConfig.EnabledTools) != 2 {
		t.Errorf("DatabaseConfig not preserved: %+v", got.DatabaseConfig)
	}
	if len(got.Redirects) != 1 || got.Redirects[0].Source != "old.example.com" {
		t.Errorf("Redirects not preserved: %+v", got.Redirects)
	}
	if got.SourceType != domain.SourceTypeGit || got.SourceConfig.Branch != "main" {
		t.Errorf("SourceConfig not preserved: %+v", got.SourceConfig)
	}
	if got.HealthCheck == nil || got.HealthCheck.IntervalSeconds != 30 {
		t.Errorf("HealthCheck not preserved: %+v", got.HealthCheck)
	}
	if got.Labels["traefik.enable"] != "true" {
		t.Errorf("Labels not preserved: %+v", got.Labels)
	}
	if len(got.EnvVars) != 2 || !got.EnvVars[1].IsSecret || got.EnvVars[1].MaskedValue() != "••••••••" {
		t.Errorf("Secret env var not preserved or masked properly: %+v", got.EnvVars)
	}

	// Test UpdateService for ProjectName, DeployScript, DatabaseConfig, and Redirects
	got.DeployScript = "npm run migrate && npm run seed"
	got.DatabaseConfig.EnabledTools = append(got.DatabaseConfig.EnabledTools, "redis-commander")
	got.Redirects[0].Target = "https://updated.example.com"
	if err := repo.UpdateService(ctx, got); err != nil {
		t.Fatalf("UpdateService failed: %v", err)
	}
	gotUpdated, err := repo.GetService(ctx, "s-adv")
	if err != nil {
		t.Fatalf("GetService after update failed: %v", err)
	}
	if gotUpdated.DeployScript != "npm run migrate && npm run seed" {
		t.Errorf("Updated DeployScript not preserved: got %s", gotUpdated.DeployScript)
	}
	if gotUpdated.DatabaseConfig == nil || len(gotUpdated.DatabaseConfig.EnabledTools) != 3 {
		t.Errorf("Updated DatabaseConfig not preserved: %+v", gotUpdated.DatabaseConfig)
	}
	if len(gotUpdated.Redirects) != 1 || gotUpdated.Redirects[0].Target != "https://updated.example.com" {
		t.Errorf("Updated Redirects not preserved: %+v", gotUpdated.Redirects)
	}

	// Test Domain with Middlewares, ProjectName, and ServiceName
	dom := &domain.Domain{
		ID:          "d-adv",
		ServiceID:   "s-adv",
		ProjectName: "adv-project",
		ServiceName: "adv-web",
		DomainName:  "adv.example.com",
		Port:        8080,
		HTTPS:       true,
		Middlewares: []string{"rate-limit", "secure-headers"},
	}
	if err := repo.CreateDomain(ctx, dom); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	gotDom, err := repo.GetDomain(ctx, "d-adv")
	if err != nil {
		t.Fatalf("GetDomain failed: %v", err)
	}
	if gotDom.ProjectName != "adv-project" || gotDom.ServiceName != "adv-web" {
		t.Errorf("Domain ProjectName or ServiceName not preserved: %+v", gotDom)
	}
	if len(gotDom.Middlewares) != 2 || gotDom.Middlewares[0] != "rate-limit" {
		t.Errorf("Middlewares not preserved: %+v", gotDom.Middlewares)
	}
}

func TestDeploymentsPagination(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	p := &domain.Project{ID: "p-pag", Name: "Pag Project"}
	_ = repo.CreateProject(ctx, p)
	s := &domain.Service{ID: "s-pag", ProjectID: "p-pag", Name: "pag-srv", Image: "alpine"}
	_ = repo.CreateService(ctx, s)

	for i := 0; i < 15; i++ {
		_ = repo.CreateDeployment(ctx, &domain.Deployment{
			ID:        fmt.Sprintf("dep-%d", i),
			ServiceID: "s-pag",
			Status:    domain.DeploymentStatusRunning,
			StartedAt: time.Now().UTC().Add(time.Duration(i) * time.Minute),
		})
	}

	page1, err := repo.ListDeploymentsByService(ctx, "s-pag", 5, 0)
	if err != nil || len(page1) != 5 {
		t.Fatalf("page 1 failed: len=%d, err=%v", len(page1), err)
	}

	page2, err := repo.ListDeploymentsByService(ctx, "s-pag", 5, 5)
	if err != nil || len(page2) != 5 {
		t.Fatalf("page 2 failed: len=%d, err=%v", len(page2), err)
	}

	if page1[0].ID == page2[0].ID {
		t.Errorf("pagination overlap detected")
	}
}

func TestActionsAndStorageProvidersCRUD(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// 1. Actions CRUD
	act := &domain.Action{
		ID:          "act-101",
		ProjectName: "myfirstproject",
		ServiceName: "chatwoot",
		Type:        domain.ActionTypeDeployment,
		Status:      domain.ActionStatusRunning,
		Description: "Deploy chatwoot v4.13.0",
		NoKill:      true,
		NoLogs:      false,
		UserID:      "usr-1",
		IsAPIAction: true,
		Meta: map[string]interface{}{
			"commit": "abc1234",
		},
	}
	if err := repo.CreateAction(ctx, act); err != nil {
		t.Fatalf("CreateAction failed: %v", err)
	}

	gotAct, err := repo.GetAction(ctx, "act-101")
	if err != nil {
		t.Fatalf("GetAction failed: %v", err)
	}
	if gotAct.Description != "Deploy chatwoot v4.13.0" || !gotAct.NoKill || !gotAct.IsAPIAction {
		t.Errorf("Action fields not preserved: %+v", gotAct)
	}
	if gotAct.Meta["commit"] != "abc1234" {
		t.Errorf("Action meta not preserved: %+v", gotAct.Meta)
	}

	// Update action
	gotAct.Status = domain.ActionStatusDone
	gotAct.Description = "Deploy finished successfully"
	if err := repo.UpdateAction(ctx, gotAct); err != nil {
		t.Fatalf("UpdateAction failed: %v", err)
	}
	gotUpdatedAct, err := repo.GetAction(ctx, "act-101")
	if err != nil || gotUpdatedAct.Status != domain.ActionStatusDone {
		t.Fatalf("UpdateAction not verified: %+v, err: %v", gotUpdatedAct, err)
	}

	// List actions
	actions, err := repo.ListActions(ctx, "myfirstproject", "chatwoot", 10, 0)
	if err != nil || len(actions) != 1 {
		t.Fatalf("ListActions failed: len=%d, err=%v", len(actions), err)
	}

	// 2. StorageProviders CRUD
	spLocal := &domain.StorageProvider{
		ID:   "sp-loc",
		Name: "Local Disk",
		Type: domain.StorageProviderTypeLocal,
		Path: "/etc/easypanel/backups",
	}
	if err := repo.CreateStorageProvider(ctx, spLocal); err != nil {
		t.Fatalf("CreateStorageProvider failed: %v", err)
	}

	spS3 := &domain.StorageProvider{
		ID:        "sp-s3",
		Name:      "AWS S3 Offsite",
		Type:      domain.StorageProviderTypeS3,
		Endpoint:  "https://s3.amazonaws.com",
		Bucket:    "company-backups",
		Region:    "us-east-1",
		AccessKey: "AKIAIOSFODNN7EXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}
	if err := repo.CreateStorageProvider(ctx, spS3); err != nil {
		t.Fatalf("CreateStorageProvider S3 failed: %v", err)
	}

	gotSP, err := repo.GetStorageProvider(ctx, "sp-s3")
	if err != nil || gotSP.Bucket != "company-backups" {
		t.Fatalf("GetStorageProvider failed: %+v, err=%v", gotSP, err)
	}

	sps, err := repo.ListStorageProviders(ctx)
	if err != nil || len(sps) != 2 {
		t.Fatalf("ListStorageProviders failed: len=%d, err=%v", len(sps), err)
	}

	if err := repo.DeleteStorageProvider(ctx, "sp-loc"); err != nil {
		t.Fatalf("DeleteStorageProvider failed: %v", err)
	}
	if _, err := repo.GetStorageProvider(ctx, "sp-loc"); err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound for deleted storage provider, got %v", err)
	}

	// 3. Service PrimaryDomainID and ZeroDowntime
	p := &domain.Project{ID: "p-zd", Name: "ZD Project"}
	_ = repo.CreateProject(ctx, p)
	s := &domain.Service{
		ID:              "s-zd",
		ProjectID:       "p-zd",
		Name:            "zd-app",
		Image:           "nginx:alpine",
		PrimaryDomainID: "dom-primary-1",
		ZeroDowntime:    true,
	}
	if err := repo.CreateService(ctx, s); err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}
	gotS, err := repo.GetService(ctx, "s-zd")
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}
	if gotS.PrimaryDomainID != "dom-primary-1" || !gotS.ZeroDowntime {
		t.Errorf("PrimaryDomainID or ZeroDowntime not preserved: %+v", gotS)
	}

	gotS.PrimaryDomainID = "dom-primary-updated"
	gotS.ZeroDowntime = false
	if err := repo.UpdateService(ctx, gotS); err != nil {
		t.Fatalf("UpdateService failed: %v", err)
	}
	gotUpdatedS, err := repo.GetService(ctx, "s-zd")
	if err != nil || gotUpdatedS.PrimaryDomainID != "dom-primary-updated" || gotUpdatedS.ZeroDowntime {
		t.Errorf("Updated PrimaryDomainID or ZeroDowntime not preserved: %+v, err=%v", gotUpdatedS, err)
	}
}
