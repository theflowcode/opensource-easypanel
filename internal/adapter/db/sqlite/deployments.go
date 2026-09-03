package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (r *Repository) CreateDeployment(ctx context.Context, d *domain.Deployment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := d.Validate(); err != nil {
		return err
	}

	if d.StartedAt.IsZero() {
		d.StartedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO deployments (id, service_id, status, trigger, commit_hash, commit_message, logs, started_at, finished_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.q.ExecContext(ctx, query,
		d.ID, d.ServiceID, string(d.Status), d.Trigger, d.CommitHash, d.CommitMessage, d.Logs, d.StartedAt, d.FinishedAt,
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

func (r *Repository) GetDeployment(ctx context.Context, id string) (*domain.Deployment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT id, service_id, status, trigger, commit_hash, commit_message, logs, started_at, finished_at FROM deployments WHERE id = ?`
	row := r.q.QueryRowContext(ctx, query, id)

	var (
		d      domain.Deployment
		status string
	)
	err := row.Scan(&d.ID, &d.ServiceID, &status, &d.Trigger, &d.CommitHash, &d.CommitMessage, &d.Logs, &d.StartedAt, &d.FinishedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	d.Status = domain.DeploymentStatus(status)
	return &d, nil
}

func (r *Repository) ListDeploymentsByService(ctx context.Context, serviceID string, limit, offset int) ([]*domain.Deployment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, service_id, status, trigger, commit_hash, commit_message, logs, started_at, finished_at
		FROM deployments
		WHERE service_id = ?
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.q.QueryContext(ctx, query, serviceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*domain.Deployment
	for rows.Next() {
		var (
			d      domain.Deployment
			status string
		)
		if err := rows.Scan(&d.ID, &d.ServiceID, &status, &d.Trigger, &d.CommitHash, &d.CommitMessage, &d.Logs, &d.StartedAt, &d.FinishedAt); err != nil {
			return nil, err
		}
		d.Status = domain.DeploymentStatus(status)
		deployments = append(deployments, &d)
	}
	return deployments, rows.Err()
}

func (r *Repository) UpdateDeployment(ctx context.Context, d *domain.Deployment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := d.Validate(); err != nil {
		return err
	}

	query := `UPDATE deployments SET status = ?, logs = ?, finished_at = ? WHERE id = ?`
	res, err := r.q.ExecContext(ctx, query, string(d.Status), d.Logs, d.FinishedAt, d.ID)
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
