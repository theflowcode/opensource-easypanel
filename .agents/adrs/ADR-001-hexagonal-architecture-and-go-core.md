# ADR-001: Hexagonal Architecture, Go Backend & Embedded Single-Binary Design

- **Status**: Accepted
- **Author**: OpenSource Easypanel Architect
- **Date**: 2026-09-03
- **Target Component**: Core & Distribution

---

## 1. Context & Problem Statement
Commercial Easypanel is implemented in Node.js (Fastify + React), requiring ~200MB+ of idle memory, multiple Node dependencies, and a containerized Node runtime. 

To create a superior, truly open-source alternative (**OpenSource Easypanel**), we require an architecture that:
1. Operates with an ultra-low resource footprint (<30MB idle RAM, <100ms cold start).
2. Distributes as a **single static binary** with embedded frontend assets (zero external dependencies).
3. Directly communicates with the Docker socket (`unix:///var/run/docker.sock`) using the official Docker Go SDK.
4. Hot-reloads Traefik routing dynamically via file provider YAMLs (`/etc/easypanel/traefik/config/*.yaml`).
5. Persists projects, services, env vars, and domains in an embedded, pure-Go SQLite database.

---

## 2. Decision Drivers
- **Memory & CPU Efficiency**: Target <30MB RAM idle on small 1GB/2GB VPS servers.
- **Portability & Simplicity**: Single executable with no external Node.js runtime needed at deployment.
- **Decoupled Hexagonal Boundaries**: Inbound ports (HTTP/WS), Outbound ports (Docker, Traefik, SQLite, Events) isolated from core business logic.
- **Traefik Compatibility**: 100% interoperable with standard Traefik 3.x dynamic routing and automatic Let's Encrypt certificates.
- **Zero Telemetry**: Completely self-contained without external phone-home calls.

---

## 3. Considered Options
1. **Option 1 (Node.js/TypeScript)**: Follow commercial Easypanel's stack. Rejected due to high memory footprint (>200MB) and multi-gigabyte container base.
2. **Option 2 (Go Backend + Embedded Modern React/Svelte SPA)**: High-performance Go compiled binary embedding Vite-built frontend assets via `embed.FS`. Communicates with Docker API directly. **Selected.**

---

## 4. Decision Outcome & Architecture Contract
Chosen Option: **Option 2 — Hexagonal Go Core with Embedded Frontend**.

```text
                                  +---------------------------------------+
                                  |         Embedded Web Dashboard        |
                                  |     (React/Svelte + Vite + Tailwind)  |
                                  +-------------------+-------------------+
                                                      |
                                           REST API / WebSockets
                                                      |
                                                      v
+---------------------------------------------------------------------------------------------------------+
|                                              GO BACKEND CORE                                            |
|                                                                                                         |
|  +-------------------------------- Inbound HTTP & WS Adapters --------------------------------------+  |
|  |  /api/v1/projects, /api/v1/services, /api/v1/templates, /api/v1/domains, /ws/logs, /ws/terminal   |  |
|  +--------------------------------------------------+-----------------------------------------------+  |
|                                                     |                                                   |
|                                                     v                                                   |
|  +------------------------------------------- Core Domain ------------------------------------------+  |
|  |  ProjectService, ServiceOrchestrator, EnvironmentManager, DomainManager, TemplateEngine          |  |
|  +--------------------------------------------------+-----------------------------------------------+  |
|                                                     |                                                   |
|                                    Abstract Outbound Interfaces (Ports)                                 |
|                +--------------------+--------------------+--------------------+                         |
|                |                    |                    |                    |                         |
|                v                    v                    v                    v                         |
|         [ DockerPort ]      [ ProxyPort ]       [ DatabasePort ]      [ EventBusPort ]                  |
+----------------+--------------------+--------------------+--------------------+-------------------------+
                 |                    |                    |                    |
                 v                    v                    v                    v
      Docker Go SDK Adapter    Traefik File Driver     SQLite Adapter    In-Memory EventBus
     (unix:///var/run/docker.sock) (/etc/easypanel/...)  (Embedded DB)   (Async Pub/Sub)
```

### Core Interface Contracts (Outbound Ports)

```go
package port

import (
	"context"
	"io"
)

// DockerPort abstracts container and swarm lifecycle management
type DockerPort interface {
	// Container & Service lifecycle
	DeployService(ctx context.Context, spec ServiceSpec) (*DeployResult, error)
	StopService(ctx context.Context, serviceID string) error
	RestartService(ctx context.Context, serviceID string) error
	DeleteService(ctx context.Context, serviceID string) error

	// Inspection & Monitoring
	GetServiceStatus(ctx context.Context, serviceID string) (*ServiceStatus, error)
	StreamServiceLogs(ctx context.Context, serviceID string, stdout, stderr io.Writer) error
	ExecServiceTerminal(ctx context.Context, serviceID string, stdin io.Reader, stdout, stderr io.Writer, resizeChan <-chan TerminalSize) error

	// Infrastructure
	EnsureNetwork(ctx context.Context, networkName string) error
	EnsureVolume(ctx context.Context, volumeName string) error
}

// ProxyDriverPort abstracts dynamic reverse proxy routing
type ProxyDriverPort interface {
	ApplyRoute(ctx context.Context, route RouteConfig) error
	RemoveRoute(ctx context.Context, serviceID string) error
	SyncAllRoutes(ctx context.Context, routes []RouteConfig) error
}

// DatabasePort abstracts metadata persistence
type DatabasePort interface {
	// Projects
	CreateProject(ctx context.Context, p *Project) error
	GetProject(ctx context.Context, id string) (*Project, error)
	ListProjects(ctx context.Context) ([]*Project, error)
	DeleteProject(ctx context.Context, id string) error

	// Services
	SaveService(ctx context.Context, s *Service) error
	GetService(ctx context.Context, id string) (*Service, error)
	ListServicesByProject(ctx context.Context, projectID string) ([]*Service, error)
	DeleteService(ctx context.Context, id string) error

	// Settings & Auth
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, val string) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	SaveUser(ctx context.Context, u *User) error
}
```

---

## 5. Positive Consequences
- **Ultra-low Resource Footprint**: Idle memory ~15-25MB, binary size ~20MB.
- **Trivial Deployment**: Can run directly on the host or inside a single minimal Docker container.
- **Fast Startup**: Starts in <50ms.
- **Zero Configuration DB**: Embedded SQLite requires no separate PostgreSQL or Redis container.

---

## 6. Verification Strategy
- **Tester Skill**: Table-driven tests validating domain services with mock ports.
- **Concurrency**: `go test -race ./...` across all packages.
- **Memory Profiling**: `pprof` heap and alloc audits verifying <30MB idle RAM constraint.
