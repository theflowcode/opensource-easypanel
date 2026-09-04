package docker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestStorageCalculations(t *testing.T) {
	tempProjects := t.TempDir()

	proj1 := filepath.Join(tempProjects, "proj1", "web")
	if err := os.MkdirAll(proj1, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	testData := []byte("hello-easypanel-storage-test")
	filePath := filepath.Join(proj1, "data.txt")
	if err := os.WriteFile(filePath, testData, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	adapter := &DockerAdapter{
		projectsDir: tempProjects,
	}

	ctx := context.Background()
	svcStorage, err := adapter.GetServiceStorage(ctx, "proj1", "web")
	if err != nil {
		t.Fatalf("GetServiceStorage failed: %v", err)
	}

	if svcStorage.SizeBytes != int64(len(testData)) {
		t.Errorf("expected size %d bytes, got %d", len(testData), svcStorage.SizeBytes)
	}

	allUsage, err := adapter.ListStorageUsage(ctx)
	if err != nil {
		t.Fatalf("ListStorageUsage failed: %v", err)
	}

	if len(allUsage) != 1 {
		t.Fatalf("expected 1 storage entry, got %d", len(allUsage))
	}
	if allUsage[0].ProjectName != "proj1" || allUsage[0].ServiceName != "web" {
		t.Errorf("unexpected storage entry: %+v", allUsage[0])
	}
}
