package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (r *Repository) CreateAction(ctx context.Context, a *domain.Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := a.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	if a.UpdatedAt.IsZero() {
		a.UpdatedAt = now
	}

	metaJSON, _ := json.Marshal(a.Meta)
	noKill := 0
	if a.NoKill {
		noKill = 1
	}
	noLogs := 0
	if a.NoLogs {
		noLogs = 1
	}
	isAPI := 0
	if a.IsAPIAction {
		isAPI = 1
	}
	isSys := 0
	if a.IsSystemAction {
		isSys = 1
	}

	query := `
		INSERT INTO actions (
			id, project_name, service_name, type, status, description,
			no_kill, no_logs, created_at, updated_at, user_id, is_api_action, is_system_action, meta
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.q.ExecContext(ctx, query,
		a.ID, a.ProjectName, a.ServiceName, a.Type, a.Status, a.Description,
		noKill, noLogs, a.CreatedAt, a.UpdatedAt, a.UserID, isAPI, isSys, string(metaJSON),
	)
	return err
}

func (r *Repository) GetAction(ctx context.Context, id string) (*domain.Action, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_name, service_name, type, status, description,
		       no_kill, no_logs, created_at, updated_at, user_id, is_api_action, is_system_action, meta
		FROM actions WHERE id = ?
	`
	var (
		a        domain.Action
		noKill   int
		noLogs   int
		isAPI    int
		isSys    int
		metaJSON string
	)
	err := r.q.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.ProjectName, &a.ServiceName, &a.Type, &a.Status, &a.Description,
		&noKill, &noLogs, &a.CreatedAt, &a.UpdatedAt, &a.UserID, &isAPI, &isSys, &metaJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	a.NoKill = noKill == 1
	a.NoLogs = noLogs == 1
	a.IsAPIAction = isAPI == 1
	a.IsSystemAction = isSys == 1
	_ = json.Unmarshal([]byte(metaJSON), &a.Meta)
	return &a, nil
}

func (r *Repository) UpdateAction(ctx context.Context, a *domain.Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := a.Validate(); err != nil {
		return err
	}
	a.UpdatedAt = time.Now().UTC()

	metaJSON, _ := json.Marshal(a.Meta)
	noKill := 0
	if a.NoKill {
		noKill = 1
	}
	noLogs := 0
	if a.NoLogs {
		noLogs = 1
	}

	query := `
		UPDATE actions SET
			status = ?, description = ?, no_kill = ?, no_logs = ?, updated_at = ?, meta = ?
		WHERE id = ?
	`
	res, err := r.q.ExecContext(ctx, query,
		a.Status, a.Description, noKill, noLogs, a.UpdatedAt, string(metaJSON), a.ID,
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

func (r *Repository) ListActions(ctx context.Context, projectName, serviceName string, limit, offset int) ([]*domain.Action, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, project_name, service_name, type, status, description,
		       no_kill, no_logs, created_at, updated_at, user_id, is_api_action, is_system_action, meta
		FROM actions
		WHERE (project_name = ? OR ? = '')
		  AND (service_name = ? OR ? = '')
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.q.QueryContext(ctx, query, projectName, projectName, serviceName, serviceName, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []*domain.Action
	for rows.Next() {
		var (
			a        domain.Action
			noKill   int
			noLogs   int
			isAPI    int
			isSys    int
			metaJSON string
		)
		if err := rows.Scan(
			&a.ID, &a.ProjectName, &a.ServiceName, &a.Type, &a.Status, &a.Description,
			&noKill, &noLogs, &a.CreatedAt, &a.UpdatedAt, &a.UserID, &isAPI, &isSys, &metaJSON,
		); err != nil {
			return nil, err
		}
		a.NoKill = noKill == 1
		a.NoLogs = noLogs == 1
		a.IsAPIAction = isAPI == 1
		a.IsSystemAction = isSys == 1
		_ = json.Unmarshal([]byte(metaJSON), &a.Meta)
		actions = append(actions, &a)
	}
	return actions, rows.Err()
}
