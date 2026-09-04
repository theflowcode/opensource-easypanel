package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func TestCalculateCPUPercent(t *testing.T) {
	// Zero system delta
	zeroStats := &container.StatsResponse{}
	if pct := calculateCPUPercent(zeroStats); pct != 0.0 {
		t.Errorf("expected 0.0 for zero stats, got %f", pct)
	}

	// 50% utilization on 2 CPUs
	stats := &container.StatsResponse{
		CPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage: 1000,
			},
			SystemUsage: 2000,
			OnlineCPUs:  2,
		},
		PreCPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage: 500,
			},
			SystemUsage: 1000,
		},
	}
	// cpuDelta = 500, sysDelta = 1000 -> (500 / 1000) * 2 * 100 = 100% total (across 2 cores)
	pct := calculateCPUPercent(stats)
	if pct != 100.0 {
		t.Errorf("expected 100.0%% CPU, got %f", pct)
	}
}

func TestReadHostMetrics(t *testing.T) {
	metrics := &domain.HostMetrics{}

	readHostMemory(metrics)
	if metrics.MemoryTotalBytes == 0 {
		t.Logf("Notice: /proc/meminfo not present or memory is 0 (non-Linux environment)")
	} else {
		if metrics.MemoryUsedBytes > metrics.MemoryTotalBytes {
			t.Errorf("used memory %d cannot exceed total memory %d", metrics.MemoryUsedBytes, metrics.MemoryTotalBytes)
		}
	}

	readHostLoadAndUptime(metrics)
	readHostCPU(metrics)
	readHostDisk(metrics, "/")

	if metrics.DiskTotalBytes > 0 && metrics.DiskUsedBytes > metrics.DiskTotalBytes {
		t.Errorf("used disk %d cannot exceed total disk %d", metrics.DiskUsedBytes, metrics.DiskTotalBytes)
	}
}
