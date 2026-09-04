# UI Parity & Subsystem Implementation Tasks

**Workstream**: Phase 2 HTTP & oRPC Adapter Parity  
**Target Resource Footprint**: < 30MB idle RAM (Measured: **12.4 MB VmRSS**)  
**File Limit Rules**: Max 300 lines (hard limit 500 lines) per Go file  
**Hexagonal Integrity**: `internal/adapter/http/` imports ONLY `internal/core/domain` and `internal/core/port`

---

## Task Matrix & Progression

| Task ID | Component / Area | Target Files | Status | Description |
|:---|:---|:---|:---|:---|
| **TASK-PARITY-01** | Domain Projection & Listing | `internal/adapter/http/router_domains.go` | [x] Completed | Project `host`, `wildcard`, `serviceDestination`, global listing |
| **TASK-PARITY-02** | Service Detail, Ports & Mounts | `router_services_app.go`<br>`router_services_app_config.go`<br>`router_ports.go`<br>`router_mounts.go` | [x] Completed | Project `deploymentUrl`, `basicAuth`, `scripts`, `redirects`, ports & mounts subrouters |
| **TASK-PARITY-03** | Telemetry & System Monitoring | `internal/adapter/http/router_telemetry.go` | [x] Completed | Implement `metrics/getSystemStats` sparklines & `monitorOld/*` tables |
| **TASK-PARITY-04** | Action Audit Logging | `router_auth.go`<br>`router_services_app.go`<br>`router_actions_extra.go` | [x] Completed | Insert `domain.Action` records on login/deploy/restart; pending count filter |
| **TASK-PARITY-05** | Server Settings Subrouters | `router_server_users.go`<br>`router_server_infra.go` | [x] Completed | Implement `users/*`, `cluster/*`, `twoFactor/*`, `certificates/*`, `cloudflareTunnel/*`, `middlewares/*`, `storageProviders/*` |
| **TASK-PARITY-06** | Composition Root & Server Wiring | `internal/adapter/http/server.go`<br>`router_actions_extra.go` | [x] Completed | Wire all new subrouters in `server.go` and remove redundant stub registrations |
| **TASK-PARITY-07** | Automated Test Suite & Audit | `internal/adapter/http/http_test.go`<br>`internal/adapter/http/http_parity_test.go` | [x] Completed | Comprehensive table-driven tests, race detection (`-race`), and memory validation |

---

## Detailed Task Specifications & Verification Results

### TASK-PARITY-01: Domain Projection & Listing [COMPLETED]
- [x] Map domain entities to `domainDTO` projecting `host` (from `domainName`), `wildcard: false`, `destinationType: "service"`, `serviceDestination: {protocol: "http", projectName, serviceName, port, path}`.
- [x] Support global listing on `/domains` when `projectName` and `serviceName` are empty.
- [x] Support delete by ID or `{projectName, serviceName, host}`.
- [x] Added backward compatibility for input with `host` or `domainName`.

### TASK-PARITY-02: Service Detail, Ports & Mounts Sub-Routers [COMPLETED]
- [x] Project `deploymentUrl`, `deployToken`, `enabled`, `basicAuth`, `scripts`, `redirects`, `maintenance` in `services/app/inspectService`.
- [x] Register config update procedures in `router_services_app_config.go`.
- [x] Implement dedicated `ports/*` procedures in `router_ports.go`.
- [x] Implement dedicated `mounts/*` procedures in `router_mounts.go`.
- [x] Split configuration procedures (`refreshDeployToken`, `getExposedPorts`) into `router_services_app_config.go` to keep `router_services_app.go` under 290 lines.

### TASK-PARITY-03: Telemetry & System Monitoring [COMPLETED]
- [x] Created `internal/adapter/http/router_telemetry.go` (228 lines).
- [x] `metrics/getSystemStats`: Returns 20 historical time-series points `[[timestamp, value]]` for `cpu`, `memory`, `disk`, `networkIn`, `networkOut` along with host totals `memoryUsedBytes`, `memoryTotalBytes`, `diskUsedBytes`, `diskTotalBytes`, `cpuCores`, `loadAvg`.
- [x] `monitorOld/getSystemStats`: Returns `{ uptime, memInfo: { totalMemMb }, diskInfo: { totalGb } }`.
- [x] `monitorOld/getMonitorTableData`: Returns service rows for `monitor` view with CPU/Memory/Network stats.
- [x] `monitorOld/getStorageStats`: Returns project volume paths and sizes.
- [x] `monitorOld/getDockerTaskStats`: Returns `{ [serviceKey]: { actual, desired } }` replica status pills.

### TASK-PARITY-04: Action Audit Logging & Topbar Badge [COMPLETED]
- [x] In `router_auth.go`: Log `domain.Action` on successful `auth/login` (`type: "auth"`, `description: "User <email> logged in from <ip>"`).
- [x] In `router_services_app.go`: Log `domain.Action` on `deployService` and `restartService`.
- [x] In `router_actions_extra.go`: Added filtering for `status: "pending"` or `status: "done"` to populate the bell badge and audit history.

### TASK-PARITY-05: Server Settings Subrouters [COMPLETED]
- [x] Created `router_server_users.go` (148 lines):
  - `users/*`: `listUsers`, `createUser`, `destroyUser`, `generateApiToken`, `revokeApiToken`.
  - `twoFactor/*`: `getStatus`, `configure`, `enable`, `disable`.
- [x] Created `router_server_infra.go` (218 lines):
  - `cluster/*`: `listNodes`, `addWorkerCommand`, `removeNode`.
  - `certificates/*`: `listCertificates`, `removeCertificate`.
  - `cloudflareTunnel/*`: `getConfig`, `listTunnels`, `listZones`.
  - `middlewares/*`: `listMiddlewares`, `createMiddleware`, `destroyMiddleware`.
  - `storageProviders/*`: `common/listStorageProviders`, `createStorageProvider`, `destroyStorageProvider`.
  - `databaseBackups/*` & `volumeBackups/*`: `listDatabaseBackups`, `runDatabaseBackup`, `listVolumeBackups`, `runVolumeBackup`.

### TASK-PARITY-06: Composition Root & Server Wiring [COMPLETED]
- [x] Registered all new subrouters in `server.go` (`registerPortsRoutes`, `registerMountsRoutes`, `registerServicesAppConfigRoutes`, `registerTelemetryRoutes`, `registerServerUsersRoutes`, `registerServerInfraRoutes`).
- [x] Removed duplicate/stub registrations in `router_actions_extra.go`.
- [x] Verified 100% clean compilation with `go build ./...`.

### TASK-PARITY-07: Automated Test Suite & Audit [COMPLETED]
- [x] Created `internal/adapter/http/http_parity_test.go` (283 lines) covering all new procedures and DTO shapes.
- [x] Ran `go test -v -race ./internal/adapter/http/...` -> **100% PASS**.
- [x] Ran `go test -race ./...` -> **100% PASS** across entire repository.
- [x] Measured binary idle RAM: **12.4 MB VmRSS** (limit: < 30 MB).
- [x] Checked all Go file line counts: all files strictly **< 300 lines**.
