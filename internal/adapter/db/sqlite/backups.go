package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (r *Repository) CreateBackup(ctx context.Context, b *domain.Backup) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := b.Validate(); err != nil {
		return err
	}
	if b.StartedAt.IsZero() {
		b.StartedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO backups (id, service_id, status, file_name, size_bytes, started_at, finished_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.q.ExecContext(ctx, query,
		b.ID, b.ServiceID, string(b.Status), b.FileName, b.SizeBytes, b.StartedAt, b.FinishedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrAlreadyExists
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *Repository) GetBackup(ctx context.Context, id string) (*domain.Backup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, service_id, status, file_name, size_bytes, started_at, finished_at
		FROM backups WHERE id = ?
	`
	row := r.q.QueryRowContext(ctx, query, id)
	return r.scanBackupRow(row)
}

func (r *Repository) ListBackupsByService(ctx context.Context, serviceID string, limit, offset int) ([]*domain.Backup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, service_id, status, file_name, size_bytes, started_at, finished_at
		FROM backups
		WHERE service_id = ?
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.q.QueryContext(ctx, query, serviceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []*domain.Backup
	for rows.Next() {
		b, err := r.scanBackupRow(rows)
		if err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, rows.Err()
}

func (r *Repository) DeleteBackup(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, err := r.q.ExecContext(ctx, `DELETE FROM backups WHERE id = ?`, id)
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

func (r *Repository) scanBackupRow(scanner rowScanner) (*domain.Backup, error) {
	var (
		b          domain.Backup
		statusStr  string
		finishedAt sql.NullTime
	)

	err := scanner.Scan(
		&b.ID,
		&b.ServiceID,
		&statusStr,
		&b.FileName,
		&b.SizeBytes,
		&b.StartedAt,
		&finishedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	b.Status = domain.BackupStatus(statusStr)
	if finishedAt.Valid {
		b.FinishedAt = &finishedAt.Time
	}
	return &b, nil
}
