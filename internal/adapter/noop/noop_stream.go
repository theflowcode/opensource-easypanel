package noop

import (
	"context"
	"io"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.StreamPort = (*NoOpStreamer)(nil)

// NoOpStreamer is a Null Object implementation of port.StreamPort when streaming is disabled.
type NoOpStreamer struct{}

func NewNoOpStreamer() *NoOpStreamer {
	return &NoOpStreamer{}
}

func (n *NoOpStreamer) SubscribeLogs(ctx context.Context, serviceID string, opts domain.LogStreamOptions, w io.Writer) error {
	return nil
}

func (n *NoOpStreamer) HandleTerminalStream(ctx context.Context, serviceID string, stdin io.Reader, stdout io.Writer, resize <-chan domain.TerminalSize) error {
	return nil
}
