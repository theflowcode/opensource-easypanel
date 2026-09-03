# OpenSource Easypanel Agent Governance & Architectural Rules

Welcome to **OpenSource Easypanel**! This document defines the architectural boundaries, engineering constraints, and parallel execution rules for all AI agents and developers working in this workspace.

---

## 1. Project Overview

- **Mission**: Build an open-source, ultra-low resource footprint (<30MB idle RAM), production-ready PaaS and server management control plane inspired by Easypanel.io.
- **Tech Stack**:
  - **Backend**: Go 1.22+ (net/http + Chi or Fiber, pure-Go SQLite `modernc.org/sqlite`, official Docker Go SDK).
  - **Frontend**: Modern SPA (React or Svelte + Vite + TypeScript + Tailwind CSS).
  - **Packaging**: Single static binary with embedded frontend assets (`//go:embed all:frontend/dist`).
  - **Orchestration**: Direct `/var/run/docker.sock` integration (Docker Engine & Docker Swarm).
  - **Reverse Proxy**: Dynamic Traefik 3.x file provider (`/etc/easypanel/traefik/config/*.yaml`) with automatic Let's Encrypt TLS.

---

## 2. Hard Non-Negotiable Constraints

1. **Memory Footprint**: Total idle RAM usage of the running server must remain strictly **< 30MB**.
2. **Zero Telemetry**: Absolute zero tracking, analytics, phone-home, or unsolicited outbound network requests.
3. **Single Binary Distribution**: The entire system (API, static frontend, migrations, templates) must compile into a self-contained executable with zero external runtime dependencies.
4. **Contract-First Modular Parallelism**:
   - The Foundation (`internal/core/port/`, `internal/core/domain/`, `internal/test/mock/`) is established and locked first.
   - Once locked, each subsystem module resides in its own isolated package (`internal/adapter/<name>/`).
   - Adapters **never** import other adapters.
   - A task or agent working on Module X must **never** modify files in Module Y.

---

## 3. Directory Layout Conventions

```text
opensource-easypanel/
├── .agents/
│   ├── adrs/             # Architectural Decision Records
│   ├── skills/           # Project skills (orchestrator, architect, builder, tester)
│   └── state/            # roadmap.md, phase_status.md, task_log.jsonl
├── cmd/
│   └── openpanel/        # Composition root (main.go with Dependency Injection)
├── internal/
│   ├── core/
│   │   ├── domain/       # Pure domain models (Project, Service, Domain, User)
│   │   ├── port/         # Interface contracts (DockerPort, ProxyPort, DBPort, etc.)
│   │   └── service/      # Orchestration services implementing business logic
│   ├── adapter/
│   │   ├── db/sqlite/    # Pure-Go SQLite repository
│   │   ├── docker/       # Docker socket & Swarm adapter
│   │   ├── proxy/traefik/# Dynamic Traefik config generator
│   │   ├── stream/       # WebSocket log streaming & PTY terminal
│   │   ├── template/     # 1-click Compose & Easypanel template engine
│   │   └── http/         # REST API routes & middleware
│   └── test/mock/        # Mock implementations of all ports for isolated testing
└── frontend/             # Modern React/Svelte SPA (Vite + TypeScript + Tailwind)
```

---

## 4. Skill Roles & Execution Sequence

- **Orchestrator** (`@orchestrator`): Coordinates roadmap progression, enforces foundation locking, dispatches parallel workstreams, and maintains `.agents/state/`.
- **Architect** (`@architect`): Designs interface contracts, authors ADRs in `.agents/adrs/`, and preserves hexagonal boundaries.
- **Builder** (`@builder`): Implements Go packages and frontend components strictly within isolated package boundaries.
- **Tester** (`@tester`): Runs table-driven unit tests, race audits (`go test -race ./...`), mock validation, and memory benchmarks (<30MB).
