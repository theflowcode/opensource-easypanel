package domain

import "time"

// HostMetrics captures server-wide CPU, memory, disk, and uptime telemetry.
type HostMetrics struct {
	CPUPercent            float64    `json:"cpuPercent"`
	CPUCores              int        `json:"cpuCores,omitempty"`
	LoadAvg               [3]float64 `json:"loadAvg,omitempty"`
	MemoryUsedBytes       uint64     `json:"memoryUsedBytes"`
	MemoryTotalBytes      uint64     `json:"memoryTotalBytes"`
	DiskUsedBytes         uint64     `json:"diskUsedBytes"`
	DiskTotalBytes        uint64     `json:"diskTotalBytes"`
	NetworkInBytesPerSec  uint64     `json:"networkInBytesPerSec,omitempty"`
	NetworkOutBytesPerSec uint64     `json:"networkOutBytesPerSec,omitempty"`
	UptimeSeconds         uint64     `json:"uptimeSeconds"`
	ReadAt                time.Time  `json:"readAt"`
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

// ServiceStorage represents disk consumption of a service data directory on host.
type ServiceStorage struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
	SizeBytes   int64  `json:"sizeBytes"`
	Path        string `json:"path"`
}

// DockerEvent encapsulates a real-time container or swarm daemon event.
type DockerEvent struct {
	Type   string    `json:"type"`   // "container", "service", "network", "volume"
	Action string    `json:"action"` // "create", "start", "die", "destroy", "restart"
	Actor  string    `json:"actor"`  // container ID or service name
	Time   time.Time `json:"time"`
}
