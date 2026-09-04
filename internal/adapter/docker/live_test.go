package docker

import (
	"context"
	"testing"
)

func TestLiveDockerAdapter(t *testing.T) {
	adapter, err := New()
	if err != nil {
		t.Skipf("skipping live docker test: docker daemon not accessible: %v", err)
	}
	defer adapter.Close()

	ctx := context.Background()

	// 1. Docker Info
	info, err := adapter.GetDockerInfo(ctx)
	if err != nil {
		t.Fatalf("GetDockerInfo failed: %v", err)
	}
	if info.ServerVersion == "" {
		t.Errorf("expected non-empty ServerVersion")
	}
	t.Logf("Docker Live Server Version: %s, Swarm Active: %t, Containers: %d",
		info.ServerVersion, info.SwarmActive, info.ContainersTotal)

	// 2. Host Metrics
	metrics, err := adapter.GetHostMetrics(ctx)
	if err != nil {
		t.Fatalf("GetHostMetrics failed: %v", err)
	}
	if metrics.MemoryTotalBytes == 0 {
		t.Errorf("expected non-zero MemoryTotalBytes")
	}
	t.Logf("Host Metrics: MemUsed: %d MB / %d MB, DiskUsed: %d MB / %d MB, Uptime: %d s",
		metrics.MemoryUsedBytes/(1024*1024),
		metrics.MemoryTotalBytes/(1024*1024),
		metrics.DiskUsedBytes/(1024*1024),
		metrics.DiskTotalBytes/(1024*1024),
		metrics.UptimeSeconds)

	// 3. Ensure Network (verify idempotency)
	testNet := "easypanel-ci-test-net"
	if err := adapter.EnsureNetwork(ctx, testNet); err != nil {
		t.Errorf("EnsureNetwork failed: %v", err)
	}
	if err := adapter.EnsureNetwork(ctx, testNet); err != nil {
		t.Errorf("EnsureNetwork second call failed: %v", err)
	}
	defer func() {
		_ = adapter.Client().NetworkRemove(context.Background(), testNet)
	}()

	// 4. Ensure Volume (verify idempotency)
	testVol := "easypanel_ci_test_vol"
	if err := adapter.EnsureVolume(ctx, testVol); err != nil {
		t.Errorf("EnsureVolume failed: %v", err)
	}
	if err := adapter.EnsureVolume(ctx, testVol); err != nil {
		t.Errorf("EnsureVolume second call failed: %v", err)
	}
	defer func() {
		_ = adapter.Client().VolumeRemove(context.Background(), testVol, true)
	}()

	// 5. List Containers
	containers, err := adapter.ListContainers(ctx)
	if err != nil {
		t.Errorf("ListContainers failed: %v", err)
	}
	t.Logf("Total active/stopped containers listed: %d", len(containers))
}
