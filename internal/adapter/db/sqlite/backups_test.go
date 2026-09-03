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

func TestSessionsCRUDAndCascades(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	u := &domain.User{
		ID:           "u-sess",
		Email:        "sess@example.com",
		PasswordHash: "bcrypt-hash",
		Role:         domain.RoleAdmin,
	}
	if err := repo.CreateUser(ctx, u); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	now := time.Now().UTC()
	sessActive := &domain.Session{
		ID:        "sess-act",
		UserID:    "u-sess",
		TokenHash: "token-active-hash",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	if err := repo.CreateSession(ctx, sessActive); err != nil {
		t.Fatalf("CreateSession active failed: %v", err)
	}

	sessExpired := &domain.Session{
		ID:        "sess-exp",
		UserID:    "u-sess",
		TokenHash: "token-expired-hash",
		ExpiresAt: now.Add(-1 * time.Hour),
		CreatedAt: now.Add(-2 * time.Hour),
	}
	if err := repo.CreateSession(ctx, sessExpired); err != nil {
		t.Fatalf("CreateSession expired failed: %v", err)
	}

	// Verify GetSession
	got, err := repo.GetSession(ctx, "token-active-hash")
	if err != nil || got.ID != "sess-act" {
		t.Fatalf("GetSession failed: got %+v, err=%v", got, err)
	}

	// Clean expired sessions
	if err := repo.DeleteExpiredSessions(ctx); err != nil {
		t.Fatalf("DeleteExpiredSessions failed: %v", err)
	}
	if _, err := repo.GetSession(ctx, "token-expired-hash"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected expired session to be deleted, got %v", err)
	}
	if _, err := repo.GetSession(ctx, "token-active-hash"); err != nil {
		t.Errorf("expected active session to still exist, got %v", err)
	}

	// Cascade delete when user is deleted
	if err := repo.DeleteUser(ctx, "u-sess"); err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
	if _, err := repo.GetSession(ctx, "token-active-hash"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected active session to be cascade deleted when user deleted, got %v", err)
	}
}
