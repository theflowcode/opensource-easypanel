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

func (r *Repository) CreateDomain(ctx context.Context, d *domain.Domain) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := d.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = now
	}

	httpsInt := 0
	if d.HTTPS {
		httpsInt = 1
	}

	midJSON, _ := json.Marshal(d.Middlewares)

	query := `
		INSERT INTO domains (id, service_id, domain_name, port, path, https, cert_mode, middlewares, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.q.ExecContext(ctx, query,
		d.ID, d.ServiceID, d.DomainName, d.Port, d.Path, httpsInt, d.CertMode, string(midJSON), d.Status, d.CreatedAt, d.UpdatedAt,
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

func (r *Repository) GetDomain(ctx context.Context, id string) (*domain.Domain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT id, service_id, domain_name, port, path, https, cert_mode, middlewares, status, created_at, updated_at FROM domains WHERE id = ?`
	row := r.q.QueryRowContext(ctx, query, id)

	var (
		d        domain.Domain
		httpsInt int
		midJSON  string
	)
	err := row.Scan(&d.ID, &d.ServiceID, &d.DomainName, &d.Port, &d.Path, &httpsInt, &d.CertMode, &midJSON, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	d.HTTPS = (httpsInt == 1)
	_ = json.Unmarshal([]byte(midJSON), &d.Middlewares)
	return &d, nil
}

func (r *Repository) ListDomainsByService(ctx context.Context, serviceID string) ([]*domain.Domain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT id, service_id, domain_name, port, path, https, cert_mode, middlewares, status, created_at, updated_at FROM domains WHERE service_id = ? ORDER BY created_at ASC`
	rows, err := r.q.QueryContext(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*domain.Domain
	for rows.Next() {
		var (
			d        domain.Domain
			httpsInt int
			midJSON  string
		)
		if err := rows.Scan(&d.ID, &d.ServiceID, &d.DomainName, &d.Port, &d.Path, &httpsInt, &d.CertMode, &midJSON, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.HTTPS = (httpsInt == 1)
		_ = json.Unmarshal([]byte(midJSON), &d.Middlewares)
		domains = append(domains, &d)
	}
	return domains, rows.Err()
}

func (r *Repository) ListAllDomains(ctx context.Context) ([]*domain.Domain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT id, service_id, domain_name, port, path, https, cert_mode, middlewares, status, created_at, updated_at FROM domains ORDER BY created_at ASC`
	rows, err := r.q.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*domain.Domain
	for rows.Next() {
		var (
			d        domain.Domain
			httpsInt int
			midJSON  string
		)
		if err := rows.Scan(&d.ID, &d.ServiceID, &d.DomainName, &d.Port, &d.Path, &httpsInt, &d.CertMode, &midJSON, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.HTTPS = (httpsInt == 1)
		_ = json.Unmarshal([]byte(midJSON), &d.Middlewares)
		domains = append(domains, &d)
	}
	return domains, rows.Err()
}

func (r *Repository) DeleteDomain(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM domains WHERE id = ?`
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
