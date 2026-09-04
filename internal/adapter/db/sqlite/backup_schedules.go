package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (r *Repository) CreateBackupSchedule(ctx context.Context, bs *domain.BackupSchedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := bs.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()
	if bs.CreatedAt.IsZero() {
		bs.CreatedAt = now
	}
	if bs.UpdatedAt.IsZero() {
		bs.UpdatedAt = now
	}

	query := `
		INSERT INTO backup_schedules (
			id, project_name, service_name, type, target_name, schedule,
			enabled, storage_provider_id, storage_provider_path, retention,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	enabledInt := 0
	if bs.Enabled {
		enabledInt = 1
	}

	_, err := r.q.ExecContext(ctx, query,
		bs.ID, bs.ProjectName, bs.ServiceName, bs.Type, bs.TargetName, bs.Schedule,
		enabledInt, bs.StorageProviderID, bs.StorageProviderPath, bs.Retention,
		bs.CreatedAt, bs.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *Repository) GetBackupSchedule(ctx context.Context, id string) (*domain.BackupSchedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_name, service_name, type, target_name, schedule,
		       enabled, storage_provider_id, storage_provider_path, retention,
		       created_at, updated_at
		FROM backup_schedules WHERE id = ?
	`
	var (
		bs      domain.BackupSchedule
		enabled int
	)
	err := r.q.QueryRowContext(ctx, query, id).Scan(
		&bs.ID, &bs.ProjectName, &bs.ServiceName, &bs.Type, &bs.TargetName, &bs.Schedule,
		&enabled, &bs.StorageProviderID, &bs.StorageProviderPath, &bs.Retention,
		&bs.CreatedAt, &bs.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	bs.Enabled = enabled == 1
	return &bs, nil
}

func (r *Repository) ListBackupSchedulesByService(ctx context.Context, projectName, serviceName string) ([]*domain.BackupSchedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_name, service_name, type, target_name, schedule,
		       enabled, storage_provider_id, storage_provider_path, retention,
		       created_at, updated_at
		FROM backup_schedules
		WHERE project_name = ? AND service_name = ?
		ORDER BY created_at ASC
	`
	rows, err := r.q.QueryContext(ctx, query, projectName, serviceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.BackupSchedule
	for rows.Next() {
		var (
			bs      domain.BackupSchedule
			enabled int
		)
		if err := rows.Scan(
			&bs.ID, &bs.ProjectName, &bs.ServiceName, &bs.Type, &bs.TargetName, &bs.Schedule,
			&enabled, &bs.StorageProviderID, &bs.StorageProviderPath, &bs.Retention,
			&bs.CreatedAt, &bs.UpdatedAt,
		); err != nil {
			return nil, err
		}
		bs.Enabled = enabled == 1
		list = append(list, &bs)
	}
	return list, rows.Err()
}

func (r *Repository) UpdateBackupSchedule(ctx context.Context, bs *domain.BackupSchedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := bs.Validate(); err != nil {
		return err
	}
	bs.UpdatedAt = time.Now().UTC()

	enabledInt := 0
	if bs.Enabled {
		enabledInt = 1
	}

	query := `
		UPDATE backup_schedules SET
			type = ?, target_name = ?, schedule = ?, enabled = ?,
			storage_provider_id = ?, storage_provider_path = ?, retention = ?,
			updated_at = ?
		WHERE id = ?
	`
	res, err := r.q.ExecContext(ctx, query,
		bs.Type, bs.TargetName, bs.Schedule, enabledInt,
		bs.StorageProviderID, bs.StorageProviderPath, bs.Retention,
		bs.UpdatedAt, bs.ID,
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

func (r *Repository) DeleteBackupSchedule(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM backup_schedules WHERE id = ?`
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
