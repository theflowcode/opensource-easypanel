package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/test/mock"
)

func TestMockBackupSchedulesAndMiddlewares(t *testing.T) {
	ctx := context.Background()
	db := mock.NewMockDatabasePort()

	// 1. BackupSchedules
	bs := &domain.BackupSchedule{
		ID:                  "bs-m-1",
		ProjectName:         "testproj",
		ServiceName:         "testsvc",
		Type:                "volume",
		TargetName:          "data-vol",
		Schedule:            "0 1 * * *",
		Enabled:             true,
		StorageProviderID:   "sp-1",
		StorageProviderPath: "/vols",
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	if err := db.CreateBackupSchedule(ctx, bs); err != nil {
		t.Fatalf("CreateBackupSchedule failed: %v", err)
	}

	gotBS, err := db.GetBackupSchedule(ctx, "bs-m-1")
	if err != nil {
		t.Fatalf("GetBackupSchedule failed: %v", err)
	}
	if gotBS.TargetName != "data-vol" {
		t.Errorf("expected TargetName=data-vol, got %s", gotBS.TargetName)
	}

	listBS, err := db.ListBackupSchedulesByService(ctx, "testproj", "testsvc")
	if err != nil {
		t.Fatalf("ListBackupSchedulesByService failed: %v", err)
	}
	if len(listBS) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(listBS))
	}

	// 2. Middlewares
	mw := &domain.Middleware{
		ID:   "mw-m-1",
		Name: "ip-allow",
		Type: "ipAllowList",
		Config: map[string]interface{}{
			"sourceRange": []string{"10.0.0.0/8"},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := db.CreateMiddleware(ctx, mw); err != nil {
		t.Fatalf("CreateMiddleware failed: %v", err)
	}

	gotMW, err := db.GetMiddleware(ctx, "mw-m-1")
	if err != nil {
		t.Fatalf("GetMiddleware failed: %v", err)
	}
	if gotMW.Name != "ip-allow" {
		t.Errorf("expected Name=ip-allow, got %s", gotMW.Name)
	}

	listMW, err := db.ListMiddlewares(ctx)
	if err != nil {
		t.Fatalf("ListMiddlewares failed: %v", err)
	}
	if len(listMW) != 1 {
		t.Fatalf("expected 1 middleware, got %d", len(listMW))
	}

	// 3. Reset
	db.Reset()
	if len(db.BackupSchedules) != 0 || len(db.Middlewares) != 0 {
		t.Errorf("expected Reset to clear BackupSchedules and Middlewares")
	}
}
