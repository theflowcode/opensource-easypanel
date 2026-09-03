package domain

import "time"

// HostMetrics captures server-wide CPU, memory, disk, and uptime telemetry.
type HostMetrics struct {
	CPUPercent       float64   `json:"cpuPercent"`
	MemoryUsedBytes  uint64    `json:"memoryUsedBytes"`
	MemoryTotalBytes uint64    `json:"memoryTotalBytes"`
	DiskUsedBytes    uint64    `json:"diskUsedBytes"`
	DiskTotalBytes   uint64    `json:"diskTotalBytes"`
	UptimeSeconds    uint64    `json:"uptimeSeconds"`
	ReadAt           time.Time `json:"readAt"`
}

// DockerInfo encapsulates Docker Engine and Swarm cluster metadata.
type DockerInfo struct {
	ServerVersion     string `json:"serverVersion"`
	SwarmActive       bool   `json:"swarmActive"`
	IsManager         bool   `json:"isManager"`
	NodeID            string `json:"nodeId,omitempty"`
	ContainersTotal   int    `json:"containersTotal"`
	ContainersRunning int    `json:"containersRunning"`
	ContainersStopped int    `json:"containersStopped"`
	ImagesTotal       int    `json:"imagesTotal"`
}
