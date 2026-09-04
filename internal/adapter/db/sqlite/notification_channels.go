package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (r *Repository) CreateNotificationChannel(ctx context.Context, ch *domain.NotificationChannel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ch.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()
	if ch.CreatedAt.IsZero() {
		ch.CreatedAt = now
	}
	if ch.UpdatedAt.IsZero() {
		ch.UpdatedAt = now
	}

	targetJSON, err := json.Marshal(ch.Target)
	if err != nil {
		return err
	}
	eventsJSON, err := json.Marshal(ch.Events)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO notification_channels (id, name, target, events, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = r.q.ExecContext(ctx, query,
		ch.ID, ch.Name, string(targetJSON), string(eventsJSON), ch.CreatedAt, ch.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *Repository) GetNotificationChannel(ctx context.Context, id string) (*domain.NotificationChannel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, name, target, events, created_at, updated_at
		FROM notification_channels WHERE id = ?
	`
	var (
		ch         domain.NotificationChannel
		targetJSON string
		eventsJSON string
	)
	err := r.q.QueryRowContext(ctx, query, id).Scan(
		&ch.ID, &ch.Name, &targetJSON, &eventsJSON, &ch.CreatedAt, &ch.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if targetJSON != "" && targetJSON != "{}" {
		_ = json.Unmarshal([]byte(targetJSON), &ch.Target)
	}
	if eventsJSON != "" && eventsJSON != "{}" {
		_ = json.Unmarshal([]byte(eventsJSON), &ch.Events)
	}
	return &ch, nil
}

func (r *Repository) ListNotificationChannels(ctx context.Context) ([]*domain.NotificationChannel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, name, target, events, created_at, updated_at
		FROM notification_channels ORDER BY name ASC
	`
	rows, err := r.q.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.NotificationChannel
	for rows.Next() {
		var (
			ch         domain.NotificationChannel
			targetJSON string
			eventsJSON string
		)
		if err := rows.Scan(
			&ch.ID, &ch.Name, &targetJSON, &eventsJSON, &ch.CreatedAt, &ch.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if targetJSON != "" && targetJSON != "{}" {
			_ = json.Unmarshal([]byte(targetJSON), &ch.Target)
		}
		if eventsJSON != "" && eventsJSON != "{}" {
			_ = json.Unmarshal([]byte(eventsJSON), &ch.Events)
		}
		list = append(list, &ch)
	}
	return list, rows.Err()
}

func (r *Repository) UpdateNotificationChannel(ctx context.Context, ch *domain.NotificationChannel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ch.Validate(); err != nil {
		return err
	}
	ch.UpdatedAt = time.Now().UTC()

	targetJSON, err := json.Marshal(ch.Target)
	if err != nil {
		return err
	}
	eventsJSON, err := json.Marshal(ch.Events)
	if err != nil {
		return err
	}

	query := `
		UPDATE notification_channels SET
			name = ?, target = ?, events = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.q.ExecContext(ctx, query,
		ch.Name, string(targetJSON), string(eventsJSON), ch.UpdatedAt, ch.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteNotificationChannel(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM notification_channels WHERE id = ?`
	res, err := r.q.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}
