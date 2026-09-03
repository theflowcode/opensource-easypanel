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
