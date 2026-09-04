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

func (r *Repository) CreateMiddleware(ctx context.Context, mw *domain.Middleware) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := mw.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()
	if mw.CreatedAt.IsZero() {
		mw.CreatedAt = now
	}
	if mw.UpdatedAt.IsZero() {
		mw.UpdatedAt = now
	}

	configJSON, err := json.Marshal(mw.Config)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO middlewares (id, name, type, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = r.q.ExecContext(ctx, query,
		mw.ID, mw.Name, mw.Type, string(configJSON), mw.CreatedAt, mw.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *Repository) GetMiddleware(ctx context.Context, id string) (*domain.Middleware, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, name, type, config, created_at, updated_at
		FROM middlewares WHERE id = ?
	`
	var (
		mw         domain.Middleware
		configJSON string
	)
	err := r.q.QueryRowContext(ctx, query, id).Scan(
		&mw.ID, &mw.Name, &mw.Type, &configJSON, &mw.CreatedAt, &mw.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if configJSON != "" && configJSON != "{}" {
		_ = json.Unmarshal([]byte(configJSON), &mw.Config)
	}
	return &mw, nil
}

func (r *Repository) ListMiddlewares(ctx context.Context) ([]*domain.Middleware, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, name, type, config, created_at, updated_at
		FROM middlewares ORDER BY name ASC
	`
	rows, err := r.q.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Middleware
	for rows.Next() {
		var (
			mw         domain.Middleware
			configJSON string
		)
		if err := rows.Scan(
			&mw.ID, &mw.Name, &mw.Type, &configJSON, &mw.CreatedAt, &mw.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if configJSON != "" && configJSON != "{}" {
			_ = json.Unmarshal([]byte(configJSON), &mw.Config)
		}
		list = append(list, &mw)
	}
	return list, rows.Err()
}

func (r *Repository) UpdateMiddleware(ctx context.Context, mw *domain.Middleware) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := mw.Validate(); err != nil {
		return err
	}
	mw.UpdatedAt = time.Now().UTC()

	configJSON, err := json.Marshal(mw.Config)
	if err != nil {
		return err
	}

	query := `
		UPDATE middlewares SET
			name = ?, type = ?, config = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.q.ExecContext(ctx, query,
		mw.Name, mw.Type, string(configJSON), mw.UpdatedAt, mw.ID,
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

func (r *Repository) DeleteMiddleware(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM middlewares WHERE id = ?`
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
