package domain

import (
	"strings"
	"time"
)

// Supported Notification Target types.
const (
	NotificationTargetDiscord  = "discord"
	NotificationTargetSlack    = "slack"
	NotificationTargetTelegram = "telegram"
	NotificationTargetSMTP     = "smtp"
	NotificationTargetWebhook  = "webhook"
)

// EventToggle controls whether a specific event triggers notifications.
type EventToggle struct {
	Enabled bool `json:"enabled"`
}

// DiskLoadConfig defines the threshold and recurring schedule for disk usage alerts.
type DiskLoadConfig struct {
	Enabled  bool    `json:"enabled"`
	Min      float64 `json:"min"`      // Percentage threshold e.g. 80.0
	Schedule string  `json:"schedule"` // Cron expression e.g. "0 2 * * *"
}

// NotificationEvents maps event types to their delivery trigger settings.
type NotificationEvents struct {
	AppDeploy       EventToggle    `json:"appDeploy"`
	DatabaseBackup  EventToggle    `json:"databaseBackup"`
	DiskLoad        DiskLoadConfig `json:"diskLoad"`
	DockerCleanup   EventToggle    `json:"dockerCleanup"`
	UpdateAvailable EventToggle    `json:"updateAvailable"`
}

// NotificationTarget contains the delivery credentials and destination endpoint.
type NotificationTarget struct {
	Type        string   `json:"type"` // "discord", "slack", "telegram", "smtp", "webhook"
	URL         string   `json:"url,omitempty"`         // discord, slack, webhook
	AccessToken string   `json:"accessToken,omitempty"` // telegram bot token
	ChatID      string   `json:"chatId,omitempty"`      // telegram chat id
	Host        string   `json:"host,omitempty"`        // smtp host
	Port        int      `json:"port,omitempty"`        // smtp port
	Username    string   `json:"username,omitempty"`    // smtp username
	Password    string   `json:"password,omitempty"`    // smtp password
	Recipients  []string `json:"recipients,omitempty"`  // smtp recipient email addresses
	Secret      string   `json:"secret,omitempty"`      // webhook authorization secret
}

// NotificationChannel models an alert routing channel configured in Server Settings.
type NotificationChannel struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Target    NotificationTarget `json:"target"`
	Events    NotificationEvents `json:"events"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

// NotificationPayload represents the alert message dispatched to a notification target.
type NotificationPayload struct {
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Level       string    `json:"level,omitempty"` // "info", "warning", "error", "success"
	ProjectName string    `json:"projectName,omitempty"`
	ServiceName string    `json:"serviceName,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// Validate ensures required fields and target-specific credentials are valid.
func (ch *NotificationChannel) Validate() error {
	if strings.TrimSpace(ch.ID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(ch.Name) == "" {
		return ErrValidation
	}
	return ch.Target.Validate()
}

// Validate ensures the target destination and credentials are well-formed.
func (t *NotificationTarget) Validate() error {
	switch t.Type {
	case NotificationTargetDiscord, NotificationTargetSlack, NotificationTargetWebhook:
		if strings.TrimSpace(t.URL) == "" {
			return ErrValidation
		}
	case NotificationTargetTelegram:
		if strings.TrimSpace(t.AccessToken) == "" || strings.TrimSpace(t.ChatID) == "" {
			return ErrValidation
		}
	case NotificationTargetSMTP:
		if strings.TrimSpace(t.Host) == "" || t.Port <= 0 || len(t.Recipients) == 0 {
			return ErrValidation
		}
	default:
		return ErrValidation
	}
	return nil
}
