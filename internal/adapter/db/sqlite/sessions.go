package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func (r *Repository) CreateSession(ctx context.Context, s *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := s.Validate(); err != nil {
		return err
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.q.ExecContext(ctx, query, s.ID, s.UserID, s.TokenHash, s.ExpiresAt, s.CreatedAt)
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

func (r *Repository) GetSession(ctx context.Context, tokenHash string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if strings.TrimSpace(tokenHash) == "" {
		return nil, domain.ErrNotFound
	}

	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM sessions WHERE token_hash = ?
	`
	row := r.q.QueryRowContext(ctx, query, tokenHash)

	var s domain.Session
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.ExpiresAt, &s.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) DeleteSession(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, err := r.q.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id)
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

func (r *Repository) DeleteExpiredSessions(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	_, err := r.q.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < ?`, now)
	return err
}
