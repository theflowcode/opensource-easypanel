package domain

import (
	"strings"
	"time"
)

// ActionType categorizes administrative and deployment operations.
const (
	ActionTypeDeployment    = "deployment"
	ActionTypeAuth          = "auth"
	ActionTypeServiceStart  = "service_start"
	ActionTypeServiceStop   = "service_stop"
	ActionTypeServiceRestart = "service_restart"
	ActionTypeServiceDelete = "service_delete"
	ActionTypeBackup        = "backup"
	ActionTypeSystemCleanup = "system_cleanup"
)

// ActionStatus tracks execution lifecycle of an action.
const (
	ActionStatusPending   = "pending"
	ActionStatusRunning   = "running"
	ActionStatusDone      = "done"
	ActionStatusFailed    = "failed"
	ActionStatusCancelled = "cancelled"
)

// Action models an asynchronous task, deployment step, or security audit log entry.
type Action struct {
	ID             string                 `json:"id"`
	ProjectName    string                 `json:"projectName,omitempty"`
	ServiceName    string                 `json:"serviceName,omitempty"`
	Type           string                 `json:"type"`
	Status         string                 `json:"status"`
	Description    string                 `json:"description"`
	NoKill         bool                   `json:"noKill,omitempty"`
	NoLogs         bool                   `json:"noLogs,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	UserID         string                 `json:"userId,omitempty"`
	IsAPIAction    bool                   `json:"isApiAction,omitempty"`
	IsSystemAction bool                   `json:"isSystemAction,omitempty"`
	Meta           map[string]interface{} `json:"meta,omitempty"`
}

// Validate ensures required action fields are provided.
func (a *Action) Validate() error {
	if strings.TrimSpace(a.ID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(a.Type) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(a.Status) == "" {
		a.Status = ActionStatusPending
	}
	return nil
}
