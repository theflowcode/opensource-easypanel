package mock_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/adapter/noop"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/test/mock"
)

func TestMockDatabasePort_NotificationChannels(t *testing.T) {
	ctx := context.Background()
	m := mock.NewMockDatabasePort()

	ch := &domain.NotificationChannel{
		ID:   "ch-mock-1",
		Name: "Discord Channel",
		Target: domain.NotificationTarget{
			Type: domain.NotificationTargetDiscord,
			URL:  "https://discord.com/webhook",
		},
	}

	// 1. Create
	if err := m.CreateNotificationChannel(ctx, ch); err != nil {
		t.Fatalf("CreateNotificationChannel failed: %v", err)
	}

	// 2. Duplicate ID
	if err := m.CreateNotificationChannel(ctx, ch); err != domain.ErrAlreadyExists {
		t.Fatalf("expected ErrAlreadyExists on duplicate ID, got %v", err)
	}

	// 3. Duplicate Name
	chDupName := &domain.NotificationChannel{
		ID:   "ch-mock-2",
		Name: "Discord Channel",
		Target: domain.NotificationTarget{
			Type: domain.NotificationTargetSlack,
			URL:  "https://slack.com/webhook",
		},
	}
	if err := m.CreateNotificationChannel(ctx, chDupName); err != domain.ErrAlreadyExists {
		t.Fatalf("expected ErrAlreadyExists on duplicate name, got %v", err)
	}

	// 4. Get
	got, err := m.GetNotificationChannel(ctx, "ch-mock-1")
	if err != nil {
		t.Fatalf("GetNotificationChannel failed: %v", err)
	}
	if got.Name != ch.Name {
		t.Fatalf("expected name %s, got %s", ch.Name, got.Name)
	}

	// 5. List
	list, err := m.ListNotificationChannels(ctx)
	if err != nil {
		t.Fatalf("ListNotificationChannels failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(list))
	}

	// 6. Update
	ch.Name = "Updated Channel"
	if err := m.UpdateNotificationChannel(ctx, ch); err != nil {
		t.Fatalf("UpdateNotificationChannel failed: %v", err)
	}

	// 7. Delete
	if err := m.DeleteNotificationChannel(ctx, "ch-mock-1"); err != nil {
		t.Fatalf("DeleteNotificationChannel failed: %v", err)
	}
	if _, err := m.GetNotificationChannel(ctx, "ch-mock-1"); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// 8. Reset
	_ = m.CreateNotificationChannel(ctx, &domain.NotificationChannel{
		ID:   "ch-reset",
		Name: "To Reset",
		Target: domain.NotificationTarget{
			Type: domain.NotificationTargetSlack,
			URL:  "https://slack.com/hook",
		},
	})
	m.Reset()
	if len(m.NotificationChannels) != 0 || len(m.Calls) != 0 {
		t.Fatalf("expected reset to clear channels and calls")
	}
}

func TestMockNotifierPort(t *testing.T) {
	ctx := context.Background()
	notifier := mock.NewMockNotifierPort()

	target := domain.NotificationTarget{
		Type: domain.NotificationTargetDiscord,
		URL:  "https://discord.com/hook",
	}
	payload := domain.NotificationPayload{
		Title:       "Deploy Success",
		Message:     "Service web deployed successfully",
		Level:       "success",
		ProjectName: "myproj",
		ServiceName: "web",
		Timestamp:   time.Now().UTC(),
	}

	// SendNotification
	if err := notifier.SendNotification(ctx, target, payload); err != nil {
		t.Fatalf("SendNotification failed: %v", err)
	}
	if len(notifier.Sent) != 1 {
		t.Fatalf("expected 1 sent notification, got %d", len(notifier.Sent))
	}
	if notifier.Sent[0].Payload.Title != "Deploy Success" {
		t.Errorf("unexpected payload title: %s", notifier.Sent[0].Payload.Title)
	}

	// SendTestNotification
	if err := notifier.SendTestNotification(ctx, target); err != nil {
		t.Fatalf("SendTestNotification failed: %v", err)
	}
	if len(notifier.Tests) != 1 {
		t.Fatalf("expected 1 test notification, got %d", len(notifier.Tests))
	}

	// Custom hook
	customErr := errors.New("custom network failure")
	notifier.SendNotificationFunc = func(ctx context.Context, target domain.NotificationTarget, payload domain.NotificationPayload) error {
		return customErr
	}
	if err := notifier.SendNotification(ctx, target, payload); !errors.Is(err, customErr) {
		t.Fatalf("expected customErr, got %v", err)
	}

	// Reset
	notifier.Reset()
	if len(notifier.Sent) != 0 || len(notifier.Tests) != 0 || len(notifier.Calls) != 0 {
		t.Fatalf("expected Reset to clear all recordings")
	}
}

func TestNoOpNotifier(t *testing.T) {
	ctx := context.Background()
	noopNotifier := noop.NewNoOpNotifier()

	target := domain.NotificationTarget{
		Type: domain.NotificationTargetDiscord,
		URL:  "https://discord.com/hook",
	}
	payload := domain.NotificationPayload{
		Title:   "Ping",
		Message: "Hello",
	}

	if err := noopNotifier.SendNotification(ctx, target, payload); err != nil {
		t.Errorf("NoOpNotifier.SendNotification returned unexpected error: %v", err)
	}
	if err := noopNotifier.SendTestNotification(ctx, target); err != nil {
		t.Errorf("NoOpNotifier.SendTestNotification returned unexpected error: %v", err)
	}
}
