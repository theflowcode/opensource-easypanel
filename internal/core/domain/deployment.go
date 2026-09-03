package domain

import (
	"strings"
	"time"
)

// Deployment records a deployment history entry for a service.
type Deployment struct {
	ID            string           `json:"id"`
	ServiceID     string           `json:"serviceId"`
	Status        DeploymentStatus `json:"status"`
	Trigger       string           `json:"trigger"` // "manual", "webhook", "template", "auto"
	CommitHash    string           `json:"commitHash,omitempty"`
	CommitMessage string           `json:"commitMessage,omitempty"`
	Logs          string           `json:"logs,omitempty"`
	StartedAt     time.Time        `json:"startedAt"`
	FinishedAt    *time.Time       `json:"finishedAt,omitempty"`
}

// Validate verifies deployment attributes.
func (d *Deployment) Validate() error {
	if strings.TrimSpace(d.ID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(d.ServiceID) == "" {
		return ErrValidation
	}
	return nil
}

// Complete marks the deployment as finished with the provided status and logs.
func (d *Deployment) Complete(status DeploymentStatus, logs string) {
	now := time.Now().UTC()
	d.Status = status
	d.Logs = logs
	d.FinishedAt = &now
}

// Fail marks the deployment as failed with the provided logs.
func (d *Deployment) Fail(logs string) {
	d.Complete(DeploymentStatusFailed, logs)
}
