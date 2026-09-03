package domain

import (
	"errors"
	"time"
)

// Standard Domain Errors
var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrValidation    = errors.New("validation error")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidState  = errors.New("invalid state transition")
	ErrInternal      = errors.New("internal server error")
)

// ServiceType defines the architectural role or database engine of a service.
type ServiceType string

const (
	ServiceTypeApp      ServiceType = "app"
	ServiceTypeDatabase ServiceType = "database" // generic fallback
	ServiceTypePostgres ServiceType = "postgres"
	ServiceTypeRedis    ServiceType = "redis"
	ServiceTypeMySQL    ServiceType = "mysql"
	ServiceTypeMariaDB  ServiceType = "mariadb"
	ServiceTypeMongoDB  ServiceType = "mongodb"
	ServiceTypeTemplate ServiceType = "template"
)

// ServiceSourceType defines how a service is sourced or built.
type ServiceSourceType string

const (
	SourceTypeImage      ServiceSourceType = "image"
	SourceTypeGit        ServiceSourceType = "git"
	SourceTypeDockerfile ServiceSourceType = "dockerfile"
	SourceTypeGithub     ServiceSourceType = "github"
	SourceTypeUpload     ServiceSourceType = "upload"
)

// SourceConfig provides build/git parameters when deploying from source.
type SourceConfig struct {
	RepoURL          string            `json:"repoUrl,omitempty"`
	Branch           string            `json:"branch,omitempty"`
	DockerfilePath   string            `json:"dockerfilePath,omitempty"`
	ContextPath      string            `json:"contextPath,omitempty"`
	BuildArgs        map[string]string `json:"buildArgs,omitempty"`
	RegistryUser     string            `json:"registryUser,omitempty"`
	RegistryPassword string            `json:"registryPassword,omitempty"`
}

// RestartPolicy constants for container lifecycle.
const (
	RestartPolicyAlways        = "always"
	RestartPolicyUnlessStopped = "unless-stopped"
	RestartPolicyOnFailure     = "on-failure"
	RestartPolicyNo            = "no"
)

// HealthCheckConfig defines container liveness/readiness probes.
type HealthCheckConfig struct {
	Test               []string `json:"test"` // e.g. ["CMD", "curl", "-f", "http://localhost:80/health"]
	IntervalSeconds    int      `json:"intervalSeconds"`
	TimeoutSeconds     int      `json:"timeoutSeconds"`
	Retries            int      `json:"retries"`
	StartPeriodSeconds int      `json:"startPeriodSeconds,omitempty"`
}

// ServiceStatus tracks the runtime state of a deployed service.
type ServiceStatus string

const (
	ServiceStatusStopped   ServiceStatus = "stopped"
	ServiceStatusStarting  ServiceStatus = "starting"
	ServiceStatusRunning   ServiceStatus = "running"
	ServiceStatusFailed    ServiceStatus = "failed"
	ServiceStatusDeploying ServiceStatus = "deploying"
)

// DeploymentStatus tracks the progression of an image build or container deploy.
type DeploymentStatus string

const (
	DeploymentStatusQueued    DeploymentStatus = "queued"
	DeploymentStatusBuilding  DeploymentStatus = "building"
	DeploymentStatusDeploying DeploymentStatus = "deploying"
	DeploymentStatusRunning   DeploymentStatus = "running"
	DeploymentStatusFailed    DeploymentStatus = "failed"
	DeploymentStatusCancelled DeploymentStatus = "cancelled"
)

// PortMapping defines a network port exposed by a container.
type PortMapping struct {
	HostPort      int    `json:"hostPort"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol"` // "tcp" or "udp"
}

// VolumeMount specifies container volume persistence.
type VolumeMount struct {
	Name          string `json:"name"`
	HostPath      string `json:"hostPath,omitempty"`
	ContainerPath string `json:"containerPath"`
	ReadOnly      bool   `json:"readOnly"`
}

// EnvVar represents a runtime environment variable with optional secret masking.
type EnvVar struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret,omitempty"`
}

// MaskedValue returns masked string if marked secret, or actual value.
func (e EnvVar) MaskedValue() string {
	if e.IsSecret && len(e.Value) > 0 {
		return "••••••••"
	}
	return e.Value
}

// ResourceLimits caps the container resource consumption.
type ResourceLimits struct {
	CPULimit    float64 `json:"cpuLimit"`    // Cores (e.g. 0.5, 1.0)
	MemoryLimit int64   `json:"memoryLimit"` // In megabytes (e.g. 256, 512)
}

// CronJobSpec defines a scheduled background worker task inside a service.
type CronJobSpec struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Schedule string `json:"schedule"` // Standard cron expression e.g. "0 0 * * *"
	Command  string `json:"command"`
}

// ServiceSpec is the deployment specification passed to DockerPort.
type ServiceSpec struct {
	ID             string             `json:"id"`
	ProjectID      string             `json:"projectId"`
	ProjectName    string             `json:"projectName,omitempty"`
	Name           string             `json:"name"`
	Type           ServiceType        `json:"type"`
	DeployToken    string             `json:"deployToken,omitempty"`
	DeployScript   string             `json:"deployScript,omitempty"`
	SourceType     ServiceSourceType  `json:"sourceType"`
	SourceConfig   *SourceConfig      `json:"sourceConfig,omitempty"`
	Image          string             `json:"image"`
	Command        string             `json:"command,omitempty"`
	Args           []string           `json:"args,omitempty"`
	EnvVars        []EnvVar           `json:"envVars,omitempty"`
	Ports          []PortMapping      `json:"ports,omitempty"`
	Volumes        []VolumeMount      `json:"volumes,omitempty"`
	NetworkAliases []string           `json:"networkAliases,omitempty"`
	Replicas       int                `json:"replicas"`
	Resources      ResourceLimits     `json:"resources"`
	RestartPolicy  string             `json:"restartPolicy,omitempty"`
	HealthCheck    *HealthCheckConfig `json:"healthCheck,omitempty"`
	CronJobs       []CronJobSpec      `json:"cronJobs,omitempty"`
	DatabaseConfig *DatabaseConfig    `json:"databaseConfig,omitempty"`
	Redirects      []RedirectRule     `json:"redirects,omitempty"`
	Labels         map[string]string  `json:"labels,omitempty"`
}

// DeployResult returns the outcome of deploying a service.
type DeployResult struct {
	ServiceID    string    `json:"serviceId"`
	ContainerIDs []string  `json:"containerIds"`
	Status       string    `json:"status"`
	DeployedAt   time.Time `json:"deployedAt"`
}

// BuildConfig defines inputs for DockerPort.BuildImage.
type BuildConfig struct {
	ServiceID      string            `json:"serviceId"`
	ImageTag       string            `json:"imageTag"`
	ContextPath    string            `json:"contextPath"`
	DockerfilePath string            `json:"dockerfilePath"`
	BuildArgs      map[string]string `json:"buildArgs,omitempty"`
}

// RegistryAuth provides authentication credentials for private container registries.
type RegistryAuth struct {
	ServerAddress string `json:"serverAddress"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

// ServiceStats snapshot of real-time container resource telemetry.
type ServiceStats struct {
	ServiceID          string    `json:"serviceId"`
	CPUPercentage      float64   `json:"cpuPercentage"`
	MemoryUsageBytes   uint64    `json:"memoryUsageBytes"`
	MemoryLimitBytes   uint64    `json:"memoryLimitBytes"`
	NetworkRxBytes     uint64    `json:"networkRxBytes"`
	NetworkTxBytes     uint64    `json:"networkTxBytes"`
	NetworkInputBytes  uint64    `json:"networkInputBytes"`
	NetworkOutputBytes uint64    `json:"networkOutputBytes"`
	ReadAt             time.Time `json:"readAt"`
}

// LogStreamOptions configures log streaming parameters.
type LogStreamOptions struct {
	TailLines  int  `json:"tailLines"`
	Follow     bool `json:"follow"`
	Timestamps bool `json:"timestamps"`
}

// TerminalSize communicates terminal window resize events.
type TerminalSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

// ContainerSummary provides a lightweight container status snapshot.
type ContainerSummary struct {
	ID        string    `json:"id"`
	Names     []string  `json:"names"`
	Image     string    `json:"image"`
	Status    string    `json:"status"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"createdAt"`
}

// Event represents an in-memory asynchronous system event.
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

type EventType string

const (
	EventServiceCreated  EventType = "service.created"
	EventServiceUpdated  EventType = "service.updated"
	EventServiceDeleted  EventType = "service.deleted"
	EventServiceDeployed EventType = "service.deployed"
	EventServiceFailed   EventType = "service.failed"
	EventDomainAdded     EventType = "domain.added"
	EventDomainRemoved   EventType = "domain.removed"
	EventRouteApplied    EventType = "route.applied"
)

// EventHandler callback for event subscriptions.
type EventHandler func(event Event)

// Subscription allows unsubscribing from event bus.
type Subscription interface {
	Unsubscribe()
}
