# OpenSource Easypanel Roadmap

## Mission
Build an open-source, production-ready, ultra-low resource footprint PaaS control plane inspired by Easypanel.io. Packaged as a single static binary in Go embedding a modern React/Svelte SPA frontend, directly interfacing with the Docker daemon and Traefik reverse proxy.

---

## Architecture Execution Model: Foundation Lock $\rightarrow$ Parallel Workstreams

```text
[ Phase 0: Ecosystem Bootstrap ] (COMPLETED)
              |
              v
[ Phase 1: Foundation Lock ] (SEQUENTIAL - Establishes contracts, ports, mocks, DB)
              |
              +-----------------------------------+-----------------------------------+
              |                                   |                                   |
              v                                   v                                   v
   [ Workstream A: Docker ]             [ Workstream B: Traefik ]           [ Workstream C: Frontend ]
  (internal/adapter/docker/)        (internal/adapter/proxy/traefik/)          (frontend/src/modules/)
              |                                   |                                   |
              v                                   v                                   v
   [ Workstream D: Stream/PTY ]         [ Workstream E: Templates ]                     |
  (internal/adapter/stream/)           (internal/adapter/template/)                   |
              |                                   |                                   |
              +-----------------------------------+-----------------------------------+
              |
              v
[ Phase 3: Integration, Wiring & Release ] (Wire dependencies, memory audit <30MB, packaging)
```

---

## Phase Breakdown

### Phase 0: Bootstrap Governance & Skill Ecosystem [COMPLETED]
- [x] Run Master Skill to analyze mission and design ecosystem
- [x] Create project skills: `orchestrator`, `architect`, `builder`, `tester`
- [x] Author ADR-001 (Hexagonal Go Core & Embedded Single-Binary Architecture)
- [x] Establish state persistence framework (`phase_status.md`, `roadmap.md`, `task_log.jsonl`)
- [x] Configure skills for **Contract-First Modular Parallelism**

---

### Phase 1: Foundation Lock (Prerequisite for Parallelism) [NEXT]
*Goal: Lock down all core domain models, interfaces, and mock harnesses so parallel tasks can proceed without interruption.*
- [ ] Initialize Go module (`go.mod`) with zero-bloat dependencies
- [ ] Implement pure domain entities in `internal/core/domain/` (`Project`, `Service`, `Domain`, `Deployment`, `User`)
- [ ] Define immutable outbound port interfaces in `internal/core/port/`:
  - `DockerPort` (container/swarm lifecycle, volume, network)
  - `ProxyDriverPort` (dynamic reverse proxy routing)
  - `DatabasePort` (metadata persistence)
  - `StreamPort` (container log streaming & PTY terminal)
  - `TemplatePort` (1-click app/db templates)
  - `EventBusPort` (in-memory async event dispatcher)
- [ ] Implement pure-Go embedded SQLite repository in `internal/adapter/db/sqlite/`
- [ ] Implement thread-safe Mock harnesses in `internal/test/mock/` for every port
- [ ] Validate Foundation Lock: All mocks and entity unit tests pass cleanly

---

### Phase 2: Parallel Workstreams Execution (Concurrent & Independent)
*Once Phase 1 is locked, all workstreams below can be worked on in parallel by independent subagents/developers without merge conflicts or cross-blocking.*

#### Workstream 2A: Docker Engine & Swarm Adapter
- [ ] Location: `internal/adapter/docker/`
- [ ] Implement `port.DockerPort` using official `github.com/docker/docker/client`
- [ ] Implement service deployment, replica scaling, stop, restart, teardown
- [ ] Implement network bridging (`easypanel` overlay/bridge) and volume mount management
- [ ] Isolated table-driven unit tests against mock daemon / unit harness

#### Workstream 2B: Traefik Dynamic Reverse Proxy Driver
- [ ] Location: `internal/adapter/proxy/traefik/`
- [ ] Implement `port.ProxyDriverPort` writing dynamic YAML to `/etc/easypanel/traefik/config/*.yaml`
- [ ] Implement automatic HTTPS redirect, router middleware, and Let's Encrypt TLS binding
- [ ] Isolated unit tests verifying YAML formatting and router rule correctness

#### Workstream 2C: Live Container Logs & Interactive PTY Terminal
- [ ] Location: `internal/adapter/stream/`
- [ ] Implement `port.StreamPort` using WebSockets
- [ ] Stream real-time container logs with backpressure
- [ ] Implement raw PTY terminal session execution via Docker container exec API
- [ ] Isolated WebSocket protocol unit tests

#### Workstream 2D: 1-Click App & Database Template Engine
- [ ] Location: `internal/adapter/template/`
- [ ] Implement `port.TemplatePort` parsing Compose / Easypanel template schemas
- [ ] Add 1-click database templates (PostgreSQL, MySQL, Redis, MongoDB)
- [ ] Add 1-click app templates (Node.js, Python, Go, PHP, WordPress)
- [ ] Isolated template parser unit tests

#### Workstream 2E: Modern React/Svelte SPA Frontend
- [ ] Location: `frontend/`
- [ ] Initialize Vite SPA with TypeScript and Tailwind CSS
- [ ] Build Projects & Services dashboard views
- [ ] Build interactive xterm.js terminal view and live log viewer
- [ ] Build environment variable, domain, and volume management forms
- [ ] Configure Vite build pipeline to output static assets to `frontend/dist/`

---

### Phase 3: Integration, Wiring & Single-Binary Release
*Goal: Wire all parallel modules together in the composition root and verify performance.*
- [ ] Wire dependencies in `cmd/openpanel/main.go` via Dependency Injection
- [ ] Embed compiled frontend static assets via `//go:embed all:frontend/dist`
- [ ] Run end-to-end integration tests against real Docker daemon and Traefik router
- [ ] Execute `go test -race ./...` across entire codebase
- [ ] Audit idle memory consumption to verify < 30MB RAM constraint
- [ ] Build standalone static executable (`dist/openpanel`)
