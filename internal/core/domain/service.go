package domain

import (
	"strings"
	"time"
)

// Service represents a workload managed by Easypanel (container, app, or database).
type Service struct {
	ID            string             `json:"id"`
	ProjectID     string             `json:"projectId"`
	ProjectName   string             `json:"projectName,omitempty"`
	Name          string             `json:"name"`
	Type          ServiceType        `json:"type"`
	DeployToken   string             `json:"deployToken,omitempty"`
	DeployScript  string             `json:"deployScript,omitempty"`
	SourceType    ServiceSourceType  `json:"sourceType"`
	SourceConfig  *SourceConfig      `json:"sourceConfig,omitempty"`
	Image         string             `json:"image"`
	Command       string             `json:"command,omitempty"`
	Args          []string           `json:"args,omitempty"`
	EnvVars       []EnvVar           `json:"envVars,omitempty"`
	Ports         []PortMapping      `json:"ports,omitempty"`
	Volumes       []VolumeMount      `json:"volumes,omitempty"`
	Domains       []string           `json:"domains,omitempty"`
	Replicas      int                `json:"replicas"`
	Resources     ResourceLimits     `json:"resources"`
	RestartPolicy  string             `json:"restartPolicy,omitempty"`
	HealthCheck    *HealthCheckConfig `json:"healthCheck,omitempty"`
	CronJobs       []CronJobSpec      `json:"cronJobs,omitempty"`
	DatabaseConfig  *DatabaseConfig    `json:"databaseConfig,omitempty"`
	Redirects       []RedirectRule     `json:"redirects,omitempty"`
	PrimaryDomainID string             `json:"primaryDomainId,omitempty"`
	ZeroDowntime    bool               `json:"zeroDowntime,omitempty"`
	Labels          map[string]string  `json:"labels,omitempty"`
	Status          ServiceStatus      `json:"status"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
}

// Validate checks that the service is well-formed.
func (s *Service) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(s.ProjectID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(s.Name) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(s.DeployToken) == "" {
		s.DeployToken = NewID()
	}
	if s.Type == "" {
		s.Type = ServiceTypeApp
	}
	if s.Type.IsDatabase() && strings.TrimSpace(s.Image) == "" {
		switch s.Type {
		case ServiceTypePostgres:
			s.Image = "postgres:16-alpine"
		case ServiceTypeRedis:
			s.Image = "redis:7-alpine"
		case ServiceTypeMySQL:
			s.Image = "mysql:8.0"
		case ServiceTypeMariaDB:
			s.Image = "mariadb:11"
		case ServiceTypeMongoDB:
			s.Image = "mongo:7"
		}
	}
	if s.SourceType == "" {
		s.SourceType = SourceTypeImage
	}
	if s.SourceType == SourceTypeImage && strings.TrimSpace(s.Image) == "" {
		return ErrValidation
	}
	if s.SourceType == SourceTypeGit || s.SourceType == SourceTypeGithub {
		if s.SourceConfig == nil || strings.TrimSpace(s.SourceConfig.RepoURL) == "" {
			return ErrValidation
		}
	}
	if s.Replicas < 0 {
		return ErrValidation
	}
	return nil
}

// ToSpec converts domain Service to deployment ServiceSpec.
func (s *Service) ToSpec() ServiceSpec {
	replicas := s.Replicas
	if replicas <= 0 {
		replicas = 1
	}

	labels := make(map[string]string)
	for k, v := range s.Labels {
		labels[k] = v
	}
	labels["easypanel.project"] = s.ProjectID
	if s.ProjectName != "" {
		labels["easypanel.projectName"] = s.ProjectName
	}
	labels["easypanel.service"] = s.ID
	labels["easypanel.name"] = s.Name

	restartPolicy := s.RestartPolicy
	if restartPolicy == "" {
		restartPolicy = RestartPolicyUnlessStopped
	}

	return ServiceSpec{
		ID:             s.ID,
		ProjectID:      s.ProjectID,
		ProjectName:    s.ProjectName,
		Name:           s.Name,
		Type:           s.Type,
		DeployToken:    s.DeployToken,
		DeployScript:   s.DeployScript,
		SourceType:     s.SourceType,
		SourceConfig:   s.SourceConfig,
		Image:          s.Image,
		Command:        s.Command,
		Args:           s.Args,
		EnvVars:        s.EnvVars,
		Ports:          s.Ports,
		Volumes:        s.Volumes,
		NetworkAliases: []string{s.Name},
		Replicas:       replicas,
		Resources:      s.Resources,
		RestartPolicy:  restartPolicy,
		HealthCheck:    s.HealthCheck,
		CronJobs:        s.CronJobs,
		DatabaseConfig:  s.DatabaseConfig,
		Redirects:       s.Redirects,
		PrimaryDomainID: s.PrimaryDomainID,
		ZeroDowntime:    s.ZeroDowntime,
		Labels:          labels,
	}
}

// CanTransitionTo validates if the service can transition to the given status.
func (s *Service) CanTransitionTo(next ServiceStatus) bool {
	if s.Status == next {
		return true
	}
	switch s.Status {
	case ServiceStatusStopped:
		return next == ServiceStatusStarting || next == ServiceStatusDeploying
	case ServiceStatusStarting:
		return next == ServiceStatusRunning || next == ServiceStatusFailed || next == ServiceStatusStopped
	case ServiceStatusRunning:
		return next == ServiceStatusStopped || next == ServiceStatusDeploying || next == ServiceStatusFailed
	case ServiceStatusDeploying:
		return next == ServiceStatusRunning || next == ServiceStatusFailed || next == ServiceStatusStopped
	case ServiceStatusFailed:
		return next == ServiceStatusStarting || next == ServiceStatusDeploying || next == ServiceStatusStopped
	default:
		return true
	}
}

// IsRunning returns true if the service is currently running.
func (s *Service) IsRunning() bool {
	return s.Status == ServiceStatusRunning
}
