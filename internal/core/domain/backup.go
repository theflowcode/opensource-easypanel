package domain

import (
	"strings"
	"time"
)

// BackupStatus defines the lifecycle state of a database snapshot.
type BackupStatus string

const (
	BackupStatusPending   BackupStatus = "pending"
	BackupStatusRunning   BackupStatus = "running"
	BackupStatusCompleted BackupStatus = "completed"
	BackupStatusFailed    BackupStatus = "failed"
)

// Backup represents a managed database dump snapshot.
type Backup struct {
	ID         string       `json:"id"`
	ServiceID  string       `json:"serviceId"`
	Status     BackupStatus `json:"status"`
	FileName   string       `json:"fileName"`
	SizeBytes  int64        `json:"sizeBytes"`
	StartedAt  time.Time    `json:"startedAt"`
	FinishedAt *time.Time   `json:"finishedAt,omitempty"`
}

// Validate ensures required backup fields are present.
func (b *Backup) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(b.ServiceID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(b.FileName) == "" {
		return ErrValidation
	}
	return nil
}

// BackupSchedule represents a recurring cron backup policy for a database or volume.
type BackupSchedule struct {
	ID                  string    `json:"id"`
	ProjectName         string    `json:"projectName"`
	ServiceName         string    `json:"serviceName"`
	Type                string    `json:"type"` // "database" or "volume"
	TargetName          string    `json:"targetName"` // database name or volume name
	Schedule            string    `json:"schedule"` // cron expression e.g. "0 2 * * *"
	Enabled             bool      `json:"enabled"`
	StorageProviderID   string    `json:"storageProviderId"`
	StorageProviderPath string    `json:"storageProviderPath"`
	Retention           int       `json:"retention,omitempty"` // retention count (for databases)
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// Validate ensures required backup schedule fields are present.
func (bs *BackupSchedule) Validate() error {
	if strings.TrimSpace(bs.ID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(bs.ProjectName) == "" || strings.TrimSpace(bs.ServiceName) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(bs.Schedule) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(bs.StorageProviderID) == "" {
		return ErrValidation
	}
	if bs.Type == "" {
		bs.Type = "database"
	}
	return nil
}
