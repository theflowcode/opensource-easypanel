package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (r *Repository) CreateProject(ctx context.Context, p *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := p.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}

	query := `INSERT INTO projects (id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.q.ExecContext(ctx, query, p.ID, p.Name, p.Description, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *Repository) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT id, name, description, created_at, updated_at FROM projects WHERE id = ?`
	row := r.q.QueryRowContext(ctx, query, id)

	var p domain.Project
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetProjectByName(ctx context.Context, name string) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT id, name, description, created_at, updated_at FROM projects WHERE name = ?`
	row := r.q.QueryRowContext(ctx, query, name)

	var p domain.Project
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT id, name, description, created_at, updated_at FROM projects ORDER BY created_at ASC`
	rows, err := r.q.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}
	return projects, rows.Err()
}

func (r *Repository) UpdateProject(ctx context.Context, p *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := p.Validate(); err != nil {
		return err
	}

	p.UpdatedAt = time.Now().UTC()
	query := `UPDATE projects SET name = ?, description = ?, updated_at = ? WHERE id = ?`
	res, err := r.q.ExecContext(ctx, query, p.Name, p.Description, p.UpdatedAt, p.ID)
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

func (r *Repository) DeleteProject(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM projects WHERE id = ?`
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
