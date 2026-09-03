package mock

import (
	"context"
	"io"
	"sync"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.StreamPort = (*MockStreamPort)(nil)

// MockStreamPort is a thread-safe mock implementing port.StreamPort.
type MockStreamPort struct {
	mu sync.RWMutex

	// Customizable function hooks
	SubscribeLogsFunc        func(ctx context.Context, serviceID string, opts domain.LogStreamOptions, w io.Writer) error
	HandleTerminalStreamFunc func(ctx context.Context, serviceID string, stdin io.Reader, stdout io.Writer, resize <-chan domain.TerminalSize) error

	// Call tracking
	Calls []string
}

func NewMockStreamPort() *MockStreamPort {
	return &MockStreamPort{}
}

// Reset clears all in-memory mock state and recorded calls.
func (m *MockStreamPort) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = nil
	m.SubscribeLogsFunc = nil
	m.HandleTerminalStreamFunc = nil
}

func (m *MockStreamPort) SubscribeLogs(ctx context.Context, serviceID string, opts domain.LogStreamOptions, w io.Writer) error {
	m.mu.Lock()
	m.Calls = append(m.Calls, "SubscribeLogs")
	m.mu.Unlock()

	if m.SubscribeLogsFunc != nil {
		return m.SubscribeLogsFunc(ctx, serviceID, opts, w)
	}
	_, err := w.Write([]byte("mock stream logs for service: " + serviceID + "\n"))
	return err
}

func (m *MockStreamPort) HandleTerminalStream(ctx context.Context, serviceID string, stdin io.Reader, stdout io.Writer, resize <-chan domain.TerminalSize) error {
	m.mu.Lock()
	m.Calls = append(m.Calls, "HandleTerminalStream")
	m.mu.Unlock()

	if m.HandleTerminalStreamFunc != nil {
		return m.HandleTerminalStreamFunc(ctx, serviceID, stdin, stdout, resize)
	}
	_, err := stdout.Write([]byte("mock terminal stream connected: " + serviceID + "\n"))
	return err
}
