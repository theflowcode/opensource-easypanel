# Phase Status

- **Project**: OpenSource Easypanel
- **Current Milestone**: Workstream 2A: Docker Engine & Swarm Adapter
- **Status**: **COMPLETED (LOCKED & VERIFIED)**
- **Next Milestone**: Workstreams 2B-2D (Traefik Dynamic Proxy Driver, Stream WebSocket PTY, 1-Click Templates Engine)
- **Subsequent Milestone**: Phase 3 — Single-Binary Release & Hardening
- **Target Memory Constraint**: < 30 MB idle RAM (Current measured: **16.4 MB VmRSS**)
- **Distribution Model**: Single-binary Go executable with embedded Vite/React SPA assets (`//go:embed all:frontend/dist`)

---

## Phase Matrix

| Phase | Milestone Name | Status | Lead Role | Execution Model |
| :--- | :--- | :--- | :--- | :--- |
| **Phase 0** | Bootstrap Governance & Parallel Skills | **COMPLETED** | Master Skill | Sequential (One-time) |
| **Phase 1** | Foundation Lock (Entities, Ports, Mocks, DB) | **COMPLETED (LOCKED)** | Architect / Builder | Sequential (Locks interface contracts) |
| **Phase 2 HTTP**| Go HTTP & oRPC Protocol Engine (`internal/adapter/http/`) | **COMPLETED** | Builder / Tester | Hexagonal Inbound Adapter |
| **Phase 2 Parity**| Full UI Parity (Telemetry, Monitor, Subrouters, Audit Logs) | **COMPLETED** | Orchestrator / Builder / Tester | Hexagonal Inbound Subrouters |
| **Phase 2 Hardening**| Security, Port Contracts & Parity Remediation | **COMPLETED (LOCKED)** | Orchestrator / Architect / Builder / Tester | Hexagonal Inbound & Database Hardening |
| **Phase 2A** | Workstream A: Docker Engine & Swarm Adapter | **COMPLETED (VERIFIED)** | Builder / Tester | **Parallel / Independent** (`internal/adapter/docker/`) |
| **Phase 2B** | Workstream B: Traefik Dynamic Proxy Driver | **UNLOCKED (Ready)** | Builder / Tester | **Parallel / Independent** (`internal/adapter/proxy/traefik/`) |
| **Phase 2C** | Workstream C: Logs & WebSocket PTY Terminal | **UNLOCKED (Ready)** | Builder / Tester | **Parallel / Independent** (`internal/adapter/stream/`) |
| **Phase 2D** | Workstream D: 1-Click App & DB Templates | **UNLOCKED (Ready)** | Builder / Tester | **Parallel / Independent** (`internal/adapter/template/`) |
| **Phase 2E** | Workstream E: Embedded Production Vite SPA | **COMPLETED** | Builder | Single-Binary Embed (`frontend/dist/`) |
| **Phase 3** | Integration Wiring, Race Test & Binary Release | *In Progress* | Orchestrator / Tester | Sequential (Wires modules in `main.go`) |
