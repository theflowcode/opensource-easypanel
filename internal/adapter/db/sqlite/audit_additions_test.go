package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func TestBackupSchedulesCRUD(t *testing.T) {
	ctx := context.Background()
	repo := setupTestDB(t)

	bs := &domain.BackupSchedule{
		ID:                  "bs-db-1",
		ProjectName:         "myproj",
		ServiceName:         "mydb",
		Type:                "database",
		TargetName:          "prod_db",
		Schedule:            "0 3 * * *",
		Enabled:             true,
		StorageProviderID:   "sp-loc",
		StorageProviderPath: "/backups/mydb",
		Retention:           14,
	}

	if err := repo.CreateBackupSchedule(ctx, bs); err != nil {
		t.Fatalf("CreateBackupSchedule failed: %v", err)
	}

	got, err := repo.GetBackupSchedule(ctx, "bs-db-1")
	if err != nil {
		t.Fatalf("GetBackupSchedule failed: %v", err)
	}
	if got.Retention != 14 || got.Schedule != "0 3 * * *" || !got.Enabled {
		t.Errorf("unexpected retrieved schedule: %+v", got)
	}

	// Update schedule
	bs.Schedule = "0 4 * * *"
	bs.Retention = 30
	bs.Enabled = false
	if err := repo.UpdateBackupSchedule(ctx, bs); err != nil {
		t.Fatalf("UpdateBackupSchedule failed: %v", err)
	}

	gotUpdated, err := repo.GetBackupSchedule(ctx, "bs-db-1")
	if err != nil {
		t.Fatalf("GetBackupSchedule after update failed: %v", err)
	}
	if gotUpdated.Retention != 30 || gotUpdated.Enabled {
		t.Errorf("expected retention=30, enabled=false, got: %+v", gotUpdated)
	}

	// List schedules by service
	list, err := repo.ListBackupSchedulesByService(ctx, "myproj", "mydb")
	if err != nil {
		t.Fatalf("ListBackupSchedulesByService failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(list))
	}

	// Delete
	if err := repo.DeleteBackupSchedule(ctx, "bs-db-1"); err != nil {
		t.Fatalf("DeleteBackupSchedule failed: %v", err)
	}
	if _, err := repo.GetBackupSchedule(ctx, "bs-db-1"); err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestStorageProviderExtendedFields(t *testing.T) {
	ctx := context.Background()
	repo := setupTestDB(t)

	sp := &domain.StorageProvider{
		ID:           "sp-sftp-1",
		Name:         "Offsite SFTP",
		Type:         domain.StorageProviderTypeSFTP,
		Host:         "192.168.1.50",
		Port:         2222,
		Username:     "backupadmin",
		Password:     "securepass",
		Path:         "/mnt/backups",
		StorageClass: "COLD",
	}

	if err := repo.CreateStorageProvider(ctx, sp); err != nil {
		t.Fatalf("CreateStorageProvider failed: %v", err)
	}

	got, err := repo.GetStorageProvider(ctx, "sp-sftp-1")
	if err != nil {
		t.Fatalf("GetStorageProvider failed: %v", err)
	}
	if got.Host != "192.168.1.50" || got.Port != 2222 || got.Username != "backupadmin" || got.StorageClass != "COLD" {
		t.Errorf("retrieved storage provider mismatch: %+v", got)
	}

	list, err := repo.ListStorageProviders(ctx)
	if err != nil {
		t.Fatalf("ListStorageProviders failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(list))
	}
}

func TestMiddlewaresCRUD(t *testing.T) {
	ctx := context.Background()
	repo := setupTestDB(t)

	mw := &domain.Middleware{
		ID:   "mw-rate-1",
		Name: "api-rate-limit",
		Type: "rateLimit",
		Config: map[string]interface{}{
			"average": float64(100),
			"burst":   float64(50),
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := repo.CreateMiddleware(ctx, mw); err != nil {
		t.Fatalf("CreateMiddleware failed: %v", err)
	}

	got, err := repo.GetMiddleware(ctx, "mw-rate-1")
	if err != nil {
		t.Fatalf("GetMiddleware failed: %v", err)
	}
	if got.Name != "api-rate-limit" || got.Type != "rateLimit" {
		t.Errorf("unexpected retrieved middleware: %+v", got)
	}
	if got.Config["burst"] != float64(50) {
		t.Errorf("expected burst=50, got %v", got.Config["burst"])
	}

	// Update middleware
	mw.Config["burst"] = float64(80)
	if err := repo.UpdateMiddleware(ctx, mw); err != nil {
		t.Fatalf("UpdateMiddleware failed: %v", err)
	}

	gotUpdated, err := repo.GetMiddleware(ctx, "mw-rate-1")
	if err != nil {
		t.Fatalf("GetMiddleware after update failed: %v", err)
	}
	if gotUpdated.Config["burst"] != float64(80) {
		t.Errorf("expected burst=80, got %v", gotUpdated.Config["burst"])
	}

	list, err := repo.ListMiddlewares(ctx)
	if err != nil {
		t.Fatalf("ListMiddlewares failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 middleware, got %d", len(list))
	}

	// Delete middleware
	if err := repo.DeleteMiddleware(ctx, "mw-rate-1"); err != nil {
		t.Fatalf("DeleteMiddleware failed: %v", err)
	}
	if _, err := repo.GetMiddleware(ctx, "mw-rate-1"); err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}
}
