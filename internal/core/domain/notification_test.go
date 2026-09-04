package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func TestNotificationChannel_Validate(t *testing.T) {
	tests := []struct {
		name    string
		channel domain.NotificationChannel
		wantErr bool
	}{
		{
			name: "valid discord channel",
			channel: domain.NotificationChannel{
				ID:   "ch-1",
				Name: "DevOps Discord",
				Target: domain.NotificationTarget{
					Type: domain.NotificationTargetDiscord,
					URL:  "https://discord.com/api/webhooks/123/xyz",
				},
				Events: domain.NotificationEvents{
					AppDeploy:     domain.EventToggle{Enabled: true},
					DockerCleanup: domain.EventToggle{Enabled: true},
					DiskLoad: domain.DiskLoadConfig{
						Enabled:  true,
						Min:      85.0,
						Schedule: "0 2 * * *",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid slack channel",
			channel: domain.NotificationChannel{
				ID:   "ch-2",
				Name: "Ops Slack",
				Target: domain.NotificationTarget{
					Type: domain.NotificationTargetSlack,
					URL:  "https://hooks.slack.com/services/T00/B00/X00",
				},
			},
			wantErr: false,
		},
		{
			name: "valid telegram channel",
			channel: domain.NotificationChannel{
				ID:   "ch-3",
				Name: "Alert Telegram",
				Target: domain.NotificationTarget{
					Type:        domain.NotificationTargetTelegram,
					AccessToken: "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
					ChatID:      "-1001234567890",
				},
			},
			wantErr: false,
		},
		{
			name: "valid smtp channel",
			channel: domain.NotificationChannel{
				ID:   "ch-4",
				Name: "Alert Email",
				Target: domain.NotificationTarget{
					Type:       domain.NotificationTargetSMTP,
					Host:       "smtp.mailgun.org",
					Port:       587,
					Username:   "postmaster@domain.com",
					Password:   "secret",
					Recipients: []string{"admin@example.com", "ops@example.com"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid webhook channel",
			channel: domain.NotificationChannel{
				ID:   "ch-5",
				Name: "Custom Webhook",
				Target: domain.NotificationTarget{
					Type:   domain.NotificationTargetWebhook,
					URL:    "https://api.example.com/alerts",
					Secret: "whsec_123",
				},
			},
			wantErr: false,
		},
		{
			name: "missing channel id",
			channel: domain.NotificationChannel{
				ID:   "",
				Name: "No ID",
				Target: domain.NotificationTarget{
					Type: domain.NotificationTargetDiscord,
					URL:  "https://discord.com/api/webhooks/123",
				},
			},
			wantErr: true,
		},
		{
			name: "missing channel name",
			channel: domain.NotificationChannel{
				ID:   "ch-bad",
				Name: "   ",
				Target: domain.NotificationTarget{
					Type: domain.NotificationTargetDiscord,
					URL:  "https://discord.com/api/webhooks/123",
				},
			},
			wantErr: true,
		},
		{
			name: "missing discord url",
			channel: domain.NotificationChannel{
				ID:   "ch-bad",
				Name: "Bad Discord",
				Target: domain.NotificationTarget{
					Type: domain.NotificationTargetDiscord,
					URL:  "",
				},
			},
			wantErr: true,
		},
		{
			name: "telegram missing chat id",
			channel: domain.NotificationChannel{
				ID:   "ch-bad",
				Name: "Bad Telegram",
				Target: domain.NotificationTarget{
					Type:        domain.NotificationTargetTelegram,
					AccessToken: "token123",
					ChatID:      "",
				},
			},
			wantErr: true,
		},
		{
			name: "smtp invalid port",
			channel: domain.NotificationChannel{
				ID:   "ch-bad",
				Name: "Bad SMTP",
				Target: domain.NotificationTarget{
					Type:       domain.NotificationTargetSMTP,
					Host:       "smtp.example.com",
					Port:       0,
					Recipients: []string{"admin@example.com"},
				},
			},
			wantErr: true,
		},
		{
			name: "smtp missing recipients",
			channel: domain.NotificationChannel{
				ID:   "ch-bad",
				Name: "Bad SMTP",
				Target: domain.NotificationTarget{
					Type:       domain.NotificationTargetSMTP,
					Host:       "smtp.example.com",
					Port:       587,
					Recipients: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "unknown target type",
			channel: domain.NotificationChannel{
				ID:   "ch-bad",
				Name: "Bad Unknown",
				Target: domain.NotificationTarget{
					Type: "pigeon_post",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.channel.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNotificationChannel_JSONSerialization(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	ch := domain.NotificationChannel{
		ID:   "ch-json",
		Name: "JSON Test",
		Target: domain.NotificationTarget{
			Type:   domain.NotificationTargetWebhook,
			URL:    "https://example.com/hook",
			Secret: "sec123",
		},
		Events: domain.NotificationEvents{
			AppDeploy:      domain.EventToggle{Enabled: true},
			DatabaseBackup: domain.EventToggle{Enabled: false},
			DiskLoad: domain.DiskLoadConfig{
				Enabled:  true,
				Min:      80.5,
				Schedule: "0 1 * * *",
			},
			DockerCleanup:   domain.EventToggle{Enabled: true},
			UpdateAvailable: domain.EventToggle{Enabled: false},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded domain.NotificationChannel
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ID != ch.ID || decoded.Name != ch.Name {
		t.Errorf("ID or Name mismatch: got %+v, want %+v", decoded, ch)
	}
	if decoded.Target.Type != ch.Target.Type || decoded.Target.URL != ch.Target.URL {
		t.Errorf("Target mismatch: got %+v, want %+v", decoded.Target, ch.Target)
	}
	if decoded.Events.DiskLoad.Min != 80.5 || decoded.Events.DiskLoad.Schedule != "0 1 * * *" {
		t.Errorf("Events.DiskLoad mismatch: got %+v", decoded.Events.DiskLoad)
	}
}
