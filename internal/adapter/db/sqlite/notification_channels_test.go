package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func TestNotificationChannelsCRUD(t *testing.T) {
	ctx := context.Background()
	repo := setupTestDB(t)

	now := time.Now().UTC().Truncate(time.Second)
	ch := &domain.NotificationChannel{
		ID:   "ch-discord-1",
		Name: "Production Discord Alerts",
		Target: domain.NotificationTarget{
			Type: domain.NotificationTargetDiscord,
			URL:  "https://discord.com/api/webhooks/123/xyz",
		},
		Events: domain.NotificationEvents{
			AppDeploy:      domain.EventToggle{Enabled: true},
			DatabaseBackup: domain.EventToggle{Enabled: true},
			DiskLoad: domain.DiskLoadConfig{
				Enabled:  true,
				Min:      80.0,
				Schedule: "0 4 * * *",
			},
			DockerCleanup:   domain.EventToggle{Enabled: false},
			UpdateAvailable: domain.EventToggle{Enabled: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 1. Create
	if err := repo.CreateNotificationChannel(ctx, ch); err != nil {
		t.Fatalf("CreateNotificationChannel failed: %v", err)
	}

	// 2. Duplicate ID returns ErrAlreadyExists
	if err := repo.CreateNotificationChannel(ctx, ch); err != domain.ErrAlreadyExists {
		t.Fatalf("expected ErrAlreadyExists on duplicate create, got: %v", err)
	}

	// 3. Validation failure
	invalidCh := &domain.NotificationChannel{
		ID:   "bad-ch",
		Name: "",
	}
	if err := repo.CreateNotificationChannel(ctx, invalidCh); err != domain.ErrValidation {
		t.Fatalf("expected ErrValidation on empty name, got: %v", err)
	}

	// 4. Get and verify unmarshaled payload
	got, err := repo.GetNotificationChannel(ctx, "ch-discord-1")
	if err != nil {
		t.Fatalf("GetNotificationChannel failed: %v", err)
	}
	if got.Name != "Production Discord Alerts" {
		t.Errorf("expected name 'Production Discord Alerts', got '%s'", got.Name)
	}
	if got.Target.Type != domain.NotificationTargetDiscord || got.Target.URL != "https://discord.com/api/webhooks/123/xyz" {
		t.Errorf("unexpected target data: %+v", got.Target)
	}
	if !got.Events.AppDeploy.Enabled || !got.Events.DatabaseBackup.Enabled || got.Events.DockerCleanup.Enabled {
		t.Errorf("unexpected events toggle: %+v", got.Events)
	}
	if got.Events.DiskLoad.Min != 80.0 || got.Events.DiskLoad.Schedule != "0 4 * * *" {
		t.Errorf("unexpected disk load config: %+v", got.Events.DiskLoad)
	}

	// 5. Get non-existent
	if _, err := repo.GetNotificationChannel(ctx, "non-existent"); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}

	// 6. Update
	ch.Name = "Updated Discord Alerts"
	ch.Target.URL = "https://discord.com/api/webhooks/999/updated"
	ch.Events.DockerCleanup.Enabled = true
	if err := repo.UpdateNotificationChannel(ctx, ch); err != nil {
		t.Fatalf("UpdateNotificationChannel failed: %v", err)
	}

	gotUpdated, err := repo.GetNotificationChannel(ctx, "ch-discord-1")
	if err != nil {
		t.Fatalf("GetNotificationChannel after update failed: %v", err)
	}
	if gotUpdated.Name != "Updated Discord Alerts" || gotUpdated.Target.URL != "https://discord.com/api/webhooks/999/updated" {
		t.Errorf("expected updated fields, got: %+v", gotUpdated)
	}
	if !gotUpdated.Events.DockerCleanup.Enabled {
		t.Errorf("expected DockerCleanup to be enabled after update")
	}

	// 7. Update non-existent
	notExistCh := &domain.NotificationChannel{
		ID:   "ghost-ch",
		Name: "Ghost",
		Target: domain.NotificationTarget{
			Type: domain.NotificationTargetSlack,
			URL:  "https://hooks.slack.com/services/1",
		},
	}
	if err := repo.UpdateNotificationChannel(ctx, notExistCh); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound updating non-existent, got: %v", err)
	}

	// 8. Create second channel for listing test
	chTelegram := &domain.NotificationChannel{
		ID:   "ch-telegram-2",
		Name: "Alpha Telegram Alerts",
		Target: domain.NotificationTarget{
			Type:        domain.NotificationTargetTelegram,
			AccessToken: "bot123:token",
			ChatID:      "12345",
		},
	}
	if err := repo.CreateNotificationChannel(ctx, chTelegram); err != nil {
		t.Fatalf("CreateNotificationChannel for telegram failed: %v", err)
	}

	// 9. List (ordered by name ASC: "Alpha Telegram Alerts" then "Updated Discord Alerts")
	list, err := repo.ListNotificationChannels(ctx)
	if err != nil {
		t.Fatalf("ListNotificationChannels failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(list))
	}
	if list[0].Name != "Alpha Telegram Alerts" || list[1].Name != "Updated Discord Alerts" {
		t.Errorf("unexpected ordering in list: %s, %s", list[0].Name, list[1].Name)
	}

	// 10. Delete
	if err := repo.DeleteNotificationChannel(ctx, "ch-discord-1"); err != nil {
		t.Fatalf("DeleteNotificationChannel failed: %v", err)
	}
	if _, err := repo.GetNotificationChannel(ctx, "ch-discord-1"); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got: %v", err)
	}

	// 11. Delete non-existent
	if err := repo.DeleteNotificationChannel(ctx, "ch-discord-1"); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound deleting already deleted, got: %v", err)
	}
}
