package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (r *Repository) GetSetting(ctx context.Context, key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT val FROM settings WHERE key = ?`
	row := r.q.QueryRowContext(ctx, query, key)

	var val string
	err := row.Scan(&val)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *Repository) SetSetting(ctx context.Context, key, val string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `INSERT INTO settings (key, val) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET val = excluded.val`
	_, err := r.q.ExecContext(ctx, query, key, val)
	return err
}

func (r *Repository) ListSettings(ctx context.Context) (map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT key, val FROM settings`
	rows, err := r.q.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		settings[k] = v
	}
	return settings, rows.Err()
}
