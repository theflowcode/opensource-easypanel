package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (r *Repository) CreateStorageProvider(ctx context.Context, sp *domain.StorageProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := sp.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()
	if sp.CreatedAt.IsZero() {
		sp.CreatedAt = now
	}
	if sp.UpdatedAt.IsZero() {
		sp.UpdatedAt = now
	}

	query := `
		INSERT INTO storage_providers (
			id, name, type, path, endpoint, bucket, region, access_key, secret_key, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.q.ExecContext(ctx, query,
		sp.ID, sp.Name, sp.Type, sp.Path, sp.Endpoint, sp.Bucket, sp.Region, sp.AccessKey, sp.SecretKey,
		sp.CreatedAt, sp.UpdatedAt,
	)
	return err
}

func (r *Repository) GetStorageProvider(ctx context.Context, id string) (*domain.StorageProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, name, type, path, endpoint, bucket, region, access_key, secret_key, created_at, updated_at
		FROM storage_providers WHERE id = ?
	`
	var sp domain.StorageProvider
	err := r.q.QueryRowContext(ctx, query, id).Scan(
		&sp.ID, &sp.Name, &sp.Type, &sp.Path, &sp.Endpoint, &sp.Bucket, &sp.Region, &sp.AccessKey, &sp.SecretKey,
		&sp.CreatedAt, &sp.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

func (r *Repository) ListStorageProviders(ctx context.Context) ([]*domain.StorageProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, name, type, path, endpoint, bucket, region, access_key, secret_key, created_at, updated_at
		FROM storage_providers ORDER BY name ASC
	`
	rows, err := r.q.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.StorageProvider
	for rows.Next() {
		var sp domain.StorageProvider
		if err := rows.Scan(
			&sp.ID, &sp.Name, &sp.Type, &sp.Path, &sp.Endpoint, &sp.Bucket, &sp.Region, &sp.AccessKey, &sp.SecretKey,
			&sp.CreatedAt, &sp.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &sp)
	}
	return list, rows.Err()
}

func (r *Repository) DeleteStorageProvider(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM storage_providers WHERE id = ?`
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
