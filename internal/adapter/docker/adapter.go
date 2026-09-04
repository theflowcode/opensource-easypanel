package docker

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/docker/docker/client"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

var _ port.DockerPort = (*DockerAdapter)(nil)

// Option defines a functional configuration option for DockerAdapter.
type Option func(*DockerAdapter)

// DockerAdapter implements port.DockerPort against the local or remote Docker daemon.
type DockerAdapter struct {
	cli            *client.Client
	projectsDir    string
	defaultNetwork string
	isSwarm        bool
	mu             sync.RWMutex
}

// WithHost sets a custom Docker daemon host/socket path.
func WithHost(host string) Option {
	return func(a *DockerAdapter) {
		if host != "" {
			_ = client.WithHost(host)(a.cli)
		}
	}
}

// WithProjectsDir overrides the base storage directory for project persistent volumes.
func WithProjectsDir(dir string) Option {
	return func(a *DockerAdapter) {
		if dir != "" {
			a.projectsDir = dir
		}
	}
}

// WithDefaultNetwork overrides the default overlay/bridge network name.
func WithDefaultNetwork(networkName string) Option {
	return func(a *DockerAdapter) {
		if networkName != "" {
			a.defaultNetwork = networkName
		}
	}
}

// WithClient injects a pre-configured Docker SDK client instance (useful for testing).
func WithClient(cli *client.Client) Option {
	return func(a *DockerAdapter) {
		a.cli = cli
	}
}

// New creates and initializes a production DockerAdapter communicating with Docker daemon.
func New(opts ...Option) (*DockerAdapter, error) {
	projDir := os.Getenv("OPENPANEL_PROJECTS_DIR")
	if projDir == "" {
		projDir = "/etc/easypanel/projects"
	}

	defNet := os.Getenv("OPENPANEL_DEFAULT_NETWORK")
	if defNet == "" {
		defNet = "easypanel"
	}

	adapter := &DockerAdapter{
		projectsDir:    projDir,
		defaultNetwork: defNet,
	}

	for _, opt := range opts {
		opt(adapter)
	}

	if adapter.cli == nil {
		host := os.Getenv("DOCKER_HOST")
		clientOpts := []client.Opt{
			client.WithAPIVersionNegotiation(),
		}
		if host != "" {
			clientOpts = append(clientOpts, client.WithHost(host))
		} else {
			clientOpts = append(clientOpts, client.FromEnv)
		}

		cli, err := client.NewClientWithOpts(clientOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create docker client: %w", err)
		}
		adapter.cli = cli
	}

	ctx := context.Background()
	ping, err := adapter.cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("docker daemon ping failed: %w", err)
	}

	info, err := adapter.cli.Info(ctx)
	if err == nil {
		adapter.isSwarm = info.Swarm.LocalNodeState == "active"
	}

	_ = ping
	return adapter, nil
}

// Close closes the underlying Docker API client.
func (a *DockerAdapter) Close() error {
	if a.cli != nil {
		return a.cli.Close()
	}
	return nil
}

// Client returns the underlying Docker SDK client.
func (a *DockerAdapter) Client() *client.Client {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cli
}

// ProjectsDir returns the base filesystem path for project storage volumes.
func (a *DockerAdapter) ProjectsDir() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.projectsDir
}

// DefaultNetwork returns the configured default overlay/bridge network.
func (a *DockerAdapter) DefaultNetwork() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.defaultNetwork
}

// IsSwarm returns whether Docker Swarm orchestration is active on this node.
func (a *DockerAdapter) IsSwarm() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isSwarm
}
