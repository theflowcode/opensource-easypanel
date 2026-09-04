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

const domainColumns = `id, service_id, project_name, service_name, domain_name, port, path, https, cert_mode, middlewares, status, created_at, updated_at`

func scanDomainRow(scanner rowScanner) (*domain.Domain, error) {
	var (
		d        domain.Domain
		httpsInt int
		midJSON  string
	)
	err := scanner.Scan(
		&d.ID, &d.ServiceID, &d.ProjectName, &d.ServiceName, &d.DomainName,
		&d.Port, &d.Path, &httpsInt, &d.CertMode, &midJSON, &d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
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
		INSERT INTO domains (
			id, service_id, project_name, service_name, domain_name, port, path, https, cert_mode, middlewares, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.q.ExecContext(ctx, query,
		d.ID, d.ServiceID, d.ProjectName, d.ServiceName, d.DomainName, d.Port, d.Path, httpsInt, d.CertMode, string(midJSON), d.Status, d.CreatedAt, d.UpdatedAt,
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

	query := `SELECT ` + domainColumns + ` FROM domains WHERE id = ?`
	return scanDomainRow(r.q.QueryRowContext(ctx, query, id))
}

func (r *Repository) ListDomainsByService(ctx context.Context, serviceID string) ([]*domain.Domain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT ` + domainColumns + ` FROM domains WHERE service_id = ? ORDER BY created_at ASC`
	rows, err := r.q.QueryContext(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*domain.Domain
	for rows.Next() {
		d, err := scanDomainRow(rows)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (r *Repository) ListAllDomains(ctx context.Context) ([]*domain.Domain, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT ` + domainColumns + ` FROM domains ORDER BY created_at ASC`
	rows, err := r.q.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*domain.Domain
	for rows.Next() {
		d, err := scanDomainRow(rows)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (r *Repository) UpdateDomain(ctx context.Context, d *domain.Domain) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := d.Validate(); err != nil {
		return err
	}

	d.UpdatedAt = time.Now().UTC()
	httpsInt := 0
	if d.HTTPS {
		httpsInt = 1
	}
	midJSON, _ := json.Marshal(d.Middlewares)

	query := `
		UPDATE domains SET
			domain_name = ?,
			port = ?,
			path = ?,
			https = ?,
			cert_mode = ?,
			middlewares = ?,
			status = ?,
			updated_at = ?
		WHERE id = ?
	`
	res, err := r.q.ExecContext(ctx, query,
		d.DomainName, d.Port, d.Path, httpsInt, d.CertMode, string(midJSON), d.Status, d.UpdatedAt, d.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrAlreadyExists
		}
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
