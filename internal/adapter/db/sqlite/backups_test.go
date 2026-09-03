package sqlite_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func TestBackupsCRUDAndCascades(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	p := &domain.Project{ID: "p-bk", Name: "Backup Proj"}
	if err := repo.CreateProject(ctx, p); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	s := &domain.Service{
		ID:          "s-bk",
		ProjectID:   "p-bk",
		Name:        "bk-postgres",
		Type:        domain.ServiceTypeDatabase,
		Image:       "postgres:16-alpine",
		DeployToken: "deploy-token-12345",
		CronJobs: []domain.CronJobSpec{
			{ID: "cron-1", Name: "nightly-cleanup", Schedule: "0 0 * * *", Command: "vacuumdb -a"},
		},
	}
	if err := repo.CreateService(ctx, s); err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	// Verify GetServiceByDeployToken
	gotByToken, err := repo.GetServiceByDeployToken(ctx, "deploy-token-12345")
	if err != nil || gotByToken.ID != "s-bk" {
		t.Fatalf("GetServiceByDeployToken failed: got %+v, err=%v", gotByToken, err)
	}
	if len(gotByToken.CronJobs) != 1 || gotByToken.CronJobs[0].Name != "nightly-cleanup" {
		t.Errorf("CronJobs not properly retrieved: %+v", gotByToken.CronJobs)
	}

	// Create 12 backups to test pagination
	for i := 0; i < 12; i++ {
		b := &domain.Backup{
			ID:        fmt.Sprintf("bk-%d", i),
			ServiceID: "s-bk",
			Status:    domain.BackupStatusCompleted,
			FileName:  fmt.Sprintf("backup-%d.sql.gz", i),
			SizeBytes: int64(1024 * (i + 1)),
			StartedAt: time.Now().UTC().Add(time.Duration(i) * time.Hour),
		}
		if err := repo.CreateBackup(ctx, b); err != nil {
			t.Fatalf("CreateBackup failed: %v", err)
		}
	}

	// Verify GetBackup
	b0, err := repo.GetBackup(ctx, "bk-0")
	if err != nil || b0.FileName != "backup-0.sql.gz" {
		t.Fatalf("GetBackup failed: got %+v, err=%v", b0, err)
	}

	// Test pagination
	page1, err := repo.ListBackupsByService(ctx, "s-bk", 5, 0)
	if err != nil || len(page1) != 5 {
		t.Fatalf("page 1 failed: len=%d, err=%v", len(page1), err)
	}
	page2, err := repo.ListBackupsByService(ctx, "s-bk", 5, 5)
	if err != nil || len(page2) != 5 {
		t.Fatalf("page 2 failed: len=%d, err=%v", len(page2), err)
	}
	if page1[0].ID == page2[0].ID {
		t.Errorf("pagination overlap detected in backups")
	}

	// Delete single backup
	if err := repo.DeleteBackup(ctx, "bk-0"); err != nil {
		t.Fatalf("DeleteBackup failed: %v", err)
	}
	if _, err := repo.GetBackup(ctx, "bk-0"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for deleted backup, got %v", err)
	}

	// Test cascade delete when project is deleted
	if err := repo.DeleteProject(ctx, "p-bk"); err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
	if _, err := repo.GetBackup(ctx, "bk-1"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected backup to be cascade deleted, got %v", err)
	}
}
