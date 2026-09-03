package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/test/mock"
)

func TestMockBackupsAndDeployToken(t *testing.T) {
	ctx := context.Background()
	db := mock.NewMockDatabasePort()

	// DeployToken test
	p := &domain.Project{ID: "p-tok", Name: "Tok Proj"}
	_ = db.CreateProject(ctx, p)
	s := &domain.Service{
		ID:          "s-tok",
		ProjectID:   "p-tok",
		Name:        "tok-srv",
		Image:       "nginx:alpine",
		DeployToken: "tok-secret-xyz",
	}
	if err := db.CreateService(ctx, s); err != nil {
		t.Fatalf("CreateService failed: %v", err)
	}

	gotSrv, err := db.GetServiceByDeployToken(ctx, "tok-secret-xyz")
	if err != nil || gotSrv.ID != "s-tok" {
		t.Errorf("GetServiceByDeployToken failed: got %+v, err=%v", gotSrv, err)
	}

	// Backup test
	b := &domain.Backup{
		ID:        "b-1",
		ServiceID: "s-tok",
		Status:    domain.BackupStatusCompleted,
		FileName:  "dump-2026.sql.gz",
		SizeBytes: 1048576,
		StartedAt: time.Now().UTC(),
	}
	if err := db.CreateBackup(ctx, b); err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	gotB, err := db.GetBackup(ctx, "b-1")
	if err != nil || gotB.FileName != "dump-2026.sql.gz" {
		t.Errorf("GetBackup failed: got %+v, err=%v", gotB, err)
	}

	list, err := db.ListBackupsByService(ctx, "s-tok", 10, 0)
	if err != nil || len(list) != 1 {
		t.Errorf("ListBackupsByService failed: len=%d, err=%v", len(list), err)
	}

	if err := db.DeleteBackup(ctx, "b-1"); err != nil {
		t.Fatalf("DeleteBackup failed: %v", err)
	}
	if _, err := db.GetBackup(ctx, "b-1"); err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound for deleted backup, got %v", err)
	}
}

func TestMockDockerInfoAndHostMetrics(t *testing.T) {
	ctx := context.Background()
	d := mock.NewMockDockerPort()

	info, err := d.GetDockerInfo(ctx)
	if err != nil || info == nil || info.ServerVersion == "" {
		t.Errorf("GetDockerInfo failed: %v", err)
	}

	metrics, err := d.GetHostMetrics(ctx)
	if err != nil || metrics == nil || metrics.MemoryTotalBytes == 0 {
		t.Errorf("GetHostMetrics failed: %v", err)
	}
}
