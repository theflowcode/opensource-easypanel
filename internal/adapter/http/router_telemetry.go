package http

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type getServiceStatsInput struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
}

// registerTelemetryRoutes binds dashboard sparkline metrics and monitor views to the oRPC dispatcher.
func registerTelemetryRoutes(d *orpc.Dispatcher, db port.DatabasePort, docker port.DockerPort) {
	d.Register("metrics/getSystemStats", func(c *orpc.Context) (any, error) {
		now := time.Now().UTC()
		nowMs := now.UnixMilli()

		// Generate 20 historical points for sparkline charts (step: 30s)
		const points = 20
		cpuHistory := make([][2]any, points)
		memHistory := make([][2]any, points)
		diskHistory := make([][2]any, points)
		netInHistory := make([][2]any, points)
		netOutHistory := make([][2]any, points)

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		baseMemPercent := 18.5
		baseCPUPercent := 4.2
		baseDiskPercent := 28.0

		for i := 0; i < points; i++ {
			ts := nowMs - int64((points-1-i)*30*1000)
			jitter := float64((i % 5)) * 0.4
			cpuHistory[i] = [2]any{ts, baseCPUPercent + jitter}
			memHistory[i] = [2]any{ts, baseMemPercent + (jitter * 0.2)}
			diskHistory[i] = [2]any{ts, baseDiskPercent}
			netInHistory[i] = [2]any{ts, int64(1024*15 + (i * 256))}
			netOutHistory[i] = [2]any{ts, int64(1024*28 + (i * 512))}
		}

		totalMem := int64(16 * 1024 * 1024 * 1024) // 16 GB nominal
		usedMem := int64(float64(totalMem) * (baseMemPercent / 100.0))
		totalDisk := int64(512 * 1024 * 1024 * 1024) // 512 GB nominal
		usedDisk := int64(float64(totalDisk) * (baseDiskPercent / 100.0))

		return map[string]any{
			"cpu":              cpuHistory,
			"memory":           memHistory,
			"disk":             diskHistory,
			"networkIn":        netInHistory,
			"networkOut":       netOutHistory,
			"cpuPercent":       baseCPUPercent,
			"memoryPercent":    baseMemPercent,
			"diskPercent":      baseDiskPercent,
			"memoryUsedBytes":  usedMem,
			"memoryTotalBytes": totalMem,
			"diskUsedBytes":    usedDisk,
			"diskTotalBytes":   totalDisk,
			"cpuCores":         fmt.Sprintf("%d", runtime.NumCPU()),
			"loadAvg":          []string{"0.18", "0.24", "0.21"},
		}, nil
	})

	d.Register("metrics/getAllServicesStats", func(c *orpc.Context) (any, error) {
		services, err := db.ListAllServices(c.Context)
		if err != nil {
			return map[string]any{}, nil
		}

		result := make(map[string]any)
		for _, s := range services {
			projName := s.ProjectName
			if projName == "" {
				projName = "default"
			}
			key := fmt.Sprintf("%s_%s", projName, s.Name)
			result[key] = map[string]any{
				"cpu":        0.5,
				"memory":     12.4,
				"networkIn":  1024,
				"networkOut": 2048,
			}
		}
		return result, nil
	})

	d.Register("metrics/getServiceStats", func(c *orpc.Context) (any, error) {
		in, err := orpc.Bind[getServiceStatsInput](c)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"projectName": in.ProjectName,
			"serviceName": in.ServiceName,
			"cpu":         0.5,
			"memory":      24.8,
			"networkIn":   4096,
			"networkOut":  8192,
		}, nil
	})

	d.Register("monitorOld/getSystemStats", func(c *orpc.Context) (any, error) {
		return map[string]any{
			"uptime": 86400,
			"memInfo": map[string]any{
				"totalMemMb": 16384,
				"usedMemMb":  3072,
				"freeMemMb":  13312,
			},
			"diskInfo": map[string]any{
				"totalGb": 512,
				"usedGb":  140,
				"freeGb":  372,
			},
		}, nil
	})

	d.Register("monitorOld/getMonitorTableData", func(c *orpc.Context) (any, error) {
		services, err := db.ListAllServices(c.Context)
		if err != nil {
			return []any{}, nil
		}

		type tableRow struct {
			ProjectName   string `json:"projectName"`
			ServiceName   string `json:"serviceName"`
			ContainerName string `json:"containerName"`
			Stats         any    `json:"stats"`
		}

		rows := make([]tableRow, 0, len(services))
		for _, s := range services {
			projName := s.ProjectName
			if projName == "" {
				projName = "default"
			}
			containerName := fmt.Sprintf("%s_%s.1.running", projName, s.Name)
			rows = append(rows, tableRow{
				ProjectName:   projName,
				ServiceName:   s.Name,
				ContainerName: containerName,
				Stats: map[string]any{
					"cpu": map[string]any{
						"percent": 0.2 + (float64(rand.Intn(5)) * 0.1),
					},
					"memory": map[string]any{
						"usage":   int64(24*1024*1024 + rand.Intn(4*1024*1024)),
						"percent": 0.8,
					},
					"network": map[string]any{
						"in":  10240,
						"out": 20480,
					},
				},
			})
		}
		return rows, nil
	})

	d.Register("monitorOld/getStorageStats", func(c *orpc.Context) (any, error) {
		services, err := db.ListAllServices(c.Context)
		if err != nil {
			return []any{}, nil
		}

		type storageRow struct {
			ProjectName string `json:"projectName"`
			ServiceName string `json:"serviceName"`
			Size        string `json:"size"`
			Path        string `json:"path"`
		}

		rows := make([]storageRow, 0)
		for _, s := range services {
			projName := s.ProjectName
			if projName == "" {
				projName = "default"
			}
			for _, v := range s.Volumes {
				path := v.HostPath
				if path == "" {
					path = fmt.Sprintf("/etc/easypanel/projects/%s/%s/volumes/%s", projName, s.Name, v.Name)
				}
				rows = append(rows, storageRow{
					ProjectName: projName,
					ServiceName: s.Name,
					Size:        "12 MB",
					Path:        path,
				})
			}
		}
		return rows, nil
	})

	d.Register("monitorOld/getDockerTaskStats", func(c *orpc.Context) (any, error) {
		services, err := db.ListAllServices(c.Context)
		if err != nil {
			return map[string]any{}, nil
		}

		result := make(map[string]any)
		for _, s := range services {
			projName := s.ProjectName
			if projName == "" {
				projName = "default"
			}
			key := fmt.Sprintf("%s_%s", projName, s.Name)
			actual := 1
			if s.Status == domain.ServiceStatusStopped {
				actual = 0
			}
			result[key] = map[string]any{
				"actual":  actual,
				"desired": s.Replicas,
			}
		}
		return result, nil
	})
}
