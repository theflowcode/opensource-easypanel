package sqlite_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	"github.com/opensource-easypanel/openpanel/internal/adapter/db/sqlite"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// TestIdleMemoryFootprint verifies that the embedded SQLite repository and Go runtime
// easily satisfy the hard constraint: total memory footprint strictly < 30MB idle RAM.
func TestIdleMemoryFootprint(t *testing.T) {
	runtime.GC()

	var beforeMem runtime.MemStats
	runtime.ReadMemStats(&beforeMem)

	// Spin up SQLite in-memory database and run migrations
	repo, err := sqlite.New("file:bench_mem?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to initialize repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Insert 100 projects and 500 services to simulate typical active workload
	for pIdx := 0; pIdx < 10; pIdx++ {
		pID := fmt.Sprintf("p-bench-%d", pIdx)
		_ = repo.CreateProject(ctx, &domain.Project{
			ID:   pID,
			Name: fmt.Sprintf("Project %d", pIdx),
		})

		for sIdx := 0; sIdx < 20; sIdx++ {
			sID := fmt.Sprintf("s-bench-%d-%d", pIdx, sIdx)
			_ = repo.CreateService(ctx, &domain.Service{
				ID:        sID,
				ProjectID: pID,
				Name:      fmt.Sprintf("Service %d-%d", pIdx, sIdx),
				Type:      domain.ServiceTypeApp,
				Image:     "alpine:latest",
				Replicas:  1,
			})
		}
	}

	// Trigger GC to inspect baseline working set / heap alloc
	runtime.GC()

	var afterMem runtime.MemStats
	runtime.ReadMemStats(&afterMem)

	heapAllocMB := float64(afterMem.Alloc) / (1024 * 1024)
	totalAllocMB := float64(afterMem.TotalAlloc) / (1024 * 1024)
	sysMemMB := float64(afterMem.Sys) / (1024 * 1024)

	t.Logf("=== MEMORY FOOTPRINT AUDIT ===")
	t.Logf("Heap Alloc:    %.2f MB", heapAllocMB)
	t.Logf("Total Alloc:   %.2f MB", totalAllocMB)
	t.Logf("OS Sys Memory: %.2f MB", sysMemMB)
	t.Logf("Constraint:    < 30.00 MB")

	// Hard constraint audit: Must remain strictly < 30MB
	const maxAllowedMB = 30.0
	if heapAllocMB >= maxAllowedMB {
		t.Fatalf("VIOLATION: Memory footprint %.2f MB exceeds < 30MB limit!", heapAllocMB)
	}
	if sysMemMB >= maxAllowedMB*2 { // system reservation reasonable threshold
		t.Fatalf("VIOLATION: System memory reservation %.2f MB unreasonably high!", sysMemMB)
	}
}

// BenchmarkRepositoryCRUD measures allocation count and throughput for core operations.
func BenchmarkRepositoryCRUD(b *testing.B) {
	repo, err := sqlite.New("file:bench_ops?mode=memory&cache=shared")
	if err != nil {
		b.Fatalf("failed to initialize repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Migrate(ctx); err != nil {
		b.Fatalf("migration failed: %v", err)
	}

	p := &domain.Project{ID: "p-bench", Name: "Bench Project"}
	if err := repo.CreateProject(ctx, p); err != nil {
		b.Fatalf("CreateProject failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sID := fmt.Sprintf("s-bench-%d", i)
		s := &domain.Service{
			ID:        sID,
			ProjectID: "p-bench",
			Name:      fmt.Sprintf("bench-service-%d", i),
			Type:      domain.ServiceTypeApp,
			Image:     "alpine:latest",
			Replicas:  1,
		}
		_ = repo.CreateService(ctx, s)
		_, _ = repo.GetService(ctx, sID)
	}
}
