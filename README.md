# OpenSource Easypanel 🚀

An open-source, ultra-low resource footprint (<30MB idle RAM), production-ready server management and PaaS control plane inspired by Easypanel.io. 

Built with a **Go backend** and an **embedded modern React/Svelte SPA frontend**, OpenSource Easypanel compiles into a single, zero-dependency static binary that communicates directly with the Docker daemon and orchestrates Traefik reverse proxy routing.

---

## ⚡ Key Highlights

- **Ultra-Low Memory Footprint**: Idles at **< 30 MB RAM** (vs ~200 MB+ for Node.js-based panels).
- **Single-Binary Zero-Dependency**: Frontend assets are compiled and embedded directly into the Go executable via `embed.FS`.
- **Direct Docker Socket Integration**: Full lifecycle control over containers, Swarm services, volumes, networks, and buildpacks via `/var/run/docker.sock`.
- **Automated Dynamic Ingress**: Generates Traefik dynamic configuration files (`/etc/easypanel/traefik/config/*.yaml`) with hot-reload and Let's Encrypt SSL.
- **Embedded Database**: Pure-Go SQLite (`modernc.org/sqlite`) with zero external database containers or CGO requirements.
- **Live Terminal & Streaming**: Real-time container log streaming and interactive PTY shell sessions over WebSockets.
- **1-Click Application Templates**: Compatible with standard Docker Compose and Easypanel template schemas.
- **100% Telemetry-Free**: Completely self-contained with zero tracking or analytics.

---

## 🏛️ Architecture

OpenSource Easypanel follows **Hexagonal / Clean Architecture** principles:

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
|  |   internal/adapter/http/ (ProjectsHandler, ServicesHandler, DomainsHandler, StreamHandler)        |  |
|  +--------------------------------------------------+-----------------------------------------------+  |
|                                                     |                                                   |
|                                                     v                                                   |
|  +----------------------------------- Core Domain & Services ---------------------------------------+  |
|  |   internal/core/domain/ (Pure Entities: Project, Service, Domain, Deployment, User)               |  |
|  |   internal/core/service/ (ProjectService, DeployService, TemplateService, DomainService)          |  |
|  +--------------------------------------------------+-----------------------------------------------+  |
|                                                     |                                                   |
|                                         Outbound Ports (Interfaces)                                     |
|               +---------------------+---------------------+---------------------+                       |
|               |                     |                     |                     |                       |
|               v                     v                     v                     v                       |
|        [ DockerPort ]         [ ProxyPort ]        [ DatabasePort ]      [ EventBusPort ]               |
+---------------+---------------------+---------------------+---------------------+-----------------------+
                |                     |                     |                     |
                v                     v                     v                     v
       Docker Go SDK Adapter    Traefik File Driver     SQLite Adapter    In-Memory EventBus
      (unix:///var/run/docker.sock) (/etc/easypanel/...) (Embedded DB)   (Async Pub/Sub)
```

---

## 🧩 Contract-First Modular Parallelism

To enable multiple developers or AI subagents to develop features concurrently without conflicts:
1. **Foundation Phase**: Defines and locks the core interfaces (`internal/core/port/`) and mock harnesses (`internal/test/mock/`).
2. **Parallel Workstreams Phase**: Modules are implemented in strict isolation:
   - `internal/adapter/docker/` $\rightarrow$ Docker Engine & Swarm lifecycle
   - `internal/adapter/proxy/traefik/` $\rightarrow$ Traefik dynamic router config
   - `internal/adapter/stream/` $\rightarrow$ WebSocket logs & PTY terminal
   - `internal/adapter/template/` $\rightarrow$ 1-click Compose & Easypanel templates
   - `frontend/src/modules/` $\rightarrow$ Modern SPA dashboard views

---

## 📋 Governance & State Persistence

The project lifecycle is tracked in `.agents/`:
- **Architecture Decisions**: [`.agents/adrs/`](.agents/adrs/)
  - [`ADR-001: Hexagonal Architecture & Go Core`](.agents/adrs/ADR-001-hexagonal-architecture-and-go-core.md)
- **Current Milestone**: [`.agents/state/phase_status.md`](.agents/state/phase_status.md)
- **Master Roadmap**: [`.agents/state/roadmap.md`](.agents/state/roadmap.md)
- **Execution Log**: [`.agents/state/task_log.jsonl`](.agents/state/task_log.jsonl)

---

## 🛠️ Project Skills

- `@orchestrator`: Milestone planning, foundation locking, parallel task dispatch.
- `@architect`: Interface contracts, ADR authoring, schema design.
- `@builder`: Go backend implementation, Docker adapter, frontend development.
- `@tester`: Table-driven tests, race detection (`-race`), memory benchmarking (<30MB).
