package port

import (
	"context"
	"io"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// StreamPort defines contracts for real-time container log streaming and PTY terminal sessions.
type StreamPort interface {
	// SubscribeLogs streams live container logs for a service to the provided writer with configurable options.
	SubscribeLogs(ctx context.Context, serviceID string, opts domain.LogStreamOptions, w io.Writer) error

	// HandleTerminalStream pipes raw PTY stdin/stdout/stderr for an interactive shell session.
	HandleTerminalStream(ctx context.Context, serviceID string, stdin io.Reader, stdout io.Writer, resize <-chan domain.TerminalSize) error
}
