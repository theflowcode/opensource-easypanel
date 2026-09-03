package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

// Ensure Repository implements port.DatabasePort
var _ port.DatabasePort = (*Repository)(nil)

type queryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Repository provides a pure-Go embedded SQLite implementation of port.DatabasePort.
type Repository struct {
	db *sql.DB
	q  queryer
	mu sync.RWMutex
}

// New creates and configures a new SQLite repository instance.
func New(dsn string) (*Repository, error) {
	// Enable WAL mode, busy timeout, and foreign keys in the DSN if not already specified
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	if !strings.Contains(dsn, "_pragma=") && !strings.Contains(dsn, "mode=memory") {
		dsn = dsn + sep + "_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// SQLite operates most reliably with a single writer connection
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	repo := &Repository{
		db: db,
		q:  db,
	}

	return repo, nil
}

// Migrate executes versioned schema migrations.
func (r *Repository) Migrate(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return runMigrations(ctx, r.q)
}

// Close closes the underlying SQLite database connection.
func (r *Repository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// WithTx executes the supplied function within an atomic transaction.
func (r *Repository) WithTx(ctx context.Context, fn func(tx port.DatabasePort) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txRepo := &Repository{
		db: r.db,
		q:  tx,
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(txRepo); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
