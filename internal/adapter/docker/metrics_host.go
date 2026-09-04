package docker

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// GetHostMetrics captures real-time server telemetry from Linux /proc and filesystem statfs.
func (a *DockerAdapter) GetHostMetrics(ctx context.Context) (*domain.HostMetrics, error) {
	metrics := &domain.HostMetrics{
		ReadAt: time.Now().UTC(),
	}

	readHostMemory(metrics)
	readHostLoadAndUptime(metrics)
	readHostCPU(metrics)
	readHostDisk(metrics, "/")

	return metrics, nil
}

func readHostMemory(m *domain.HostMetrics) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var totalKB, availableKB, freeKB uint64
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(parts[1], 10, 64)
		switch parts[0] {
		case "MemTotal:":
			totalKB = val
		case "MemAvailable:":
			availableKB = val
		case "MemFree:":
			freeKB = val
		}
	}

	m.MemoryTotalBytes = totalKB * 1024
	if availableKB > 0 && totalKB >= availableKB {
		m.MemoryUsedBytes = (totalKB - availableKB) * 1024
	} else if freeKB > 0 && totalKB >= freeKB {
		m.MemoryUsedBytes = (totalKB - freeKB) * 1024
	}
}

func readHostLoadAndUptime(m *domain.HostMetrics) {
	if data, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			l1, _ := strconv.ParseFloat(fields[0], 64)
			l5, _ := strconv.ParseFloat(fields[1], 64)
			l15, _ := strconv.ParseFloat(fields[2], 64)
			m.LoadAvg = [3]float64{l1, l5, l15}
		}
	}

	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 1 {
			if upSec, err := strconv.ParseFloat(fields[0], 64); err == nil {
				m.UptimeSeconds = uint64(upSec)
			}
		}
	}
}

func readHostCPU(m *domain.HostMetrics) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	cores := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu") && len(line) > 3 && line[3] >= '0' && line[3] <= '9' {
			cores++
		}
	}
	if cores > 0 {
		m.CPUCores = cores
		if m.LoadAvg[0] > 0 {
			pct := (m.LoadAvg[0] / float64(cores)) * 100.0
			if pct > 100.0 {
				pct = 100.0
			}
			m.CPUPercent = pct
		}
	}
}

func readHostDisk(m *domain.HostMetrics, path string) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err == nil {
		bsize := uint64(stat.Bsize)
		m.DiskTotalBytes = stat.Blocks * bsize
		freeBytes := stat.Bavail * bsize
		if m.DiskTotalBytes >= freeBytes {
			m.DiskUsedBytes = m.DiskTotalBytes - freeBytes
		}
	}
}
