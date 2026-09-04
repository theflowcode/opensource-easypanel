# Phase 1 & Phase 2 Remediation & Hardening Tasks

**Workstream**: Pre-Workstream 2A Hardening & Parity Remediation  
**Target Resource Footprint**: < 30MB idle RAM (Measured: **12.8 MB VmRSS**)  
**File Limit Rules**: Max 300 lines (hard limit 500 lines) per Go file (Audited: 100% compliant)  
**Hexagonal Integrity**: Inbound adapters depend exclusively on `domain` and `port` interfaces  

---

## Task Matrix

| Task ID | Focus Area | Target Files | Status | Description |
|:---|:---|:---|:---|:---|
| **TASK-REM-01** | Session Security & Entropy | `internal/adapter/http/orpc/dispatcher.go`<br>`internal/adapter/http/router_setup.go` | [x] Completed | Remove unhashed token fallback; handle CSPRNG errors on token generation |
| **TASK-REM-02** | Comprehensive RBAC Lockdown | `router_services_app.go`<br>`router_services_app_ops.go`<br>`router_services_app_config.go`<br>`router_services_db.go`<br>`router_services_common.go`<br>`router_server_users.go`<br>`router_actions_extra.go`<br>`router_server_infra.go`<br>`router_server_storage.go`<br>`router_settings.go`<br>`router_projects.go` | [x] Completed | Enforce `RequireAdmin()` on all mutating endpoints; enforce `RequireAuth()` on inspection |
| **TASK-REM-03** | CORS Origin Whitelisting | `internal/adapter/http/server.go` | [x] Completed | Remove arbitrary origin reflection with credentials; restrict to Host / configured domains |
| **TASK-REM-04** | Domain Update Port Contract | `internal/core/port/db.go`<br>`internal/adapter/db/sqlite/domains.go`<br>`internal/test/mock/mock_db_domains.go`<br>`internal/adapter/http/router_domains.go` | [x] Completed | Add `UpdateDomain` to `DatabasePort`, SQLite, Mocks, and wire genuine handler in router |
| **TASK-REM-05** | User Update & API Token Storage | `internal/adapter/http/router_server_users.go` | [x] Completed | Implement `users/updateUser`; persist generated tokens to SQLite `sessions` table |
| **TASK-REM-06** | Storage Providers & Notifications Parity | `internal/adapter/http/router_server_storage.go`<br>`internal/adapter/http/router_actions_extra.go` | [x] Completed | Register `storageProviders/common/list`, provider-specific routes, `notifications/sendTestNotification`, mask secrets |
| **TASK-REM-07** | RFC 6455 WebSocket Stubs | `internal/adapter/http/server.go` | [x] Completed | Implement RFC 6455 HTTP 101 WebSocket upgrade to eliminate SPA infinite reconnect storm |
| **TASK-REM-08** | Multi-Port Server DI Struct | `internal/adapter/http/server.go`<br>`cmd/openpanel/main.go`<br>`internal/adapter/http/http_test.go`<br>`internal/adapter/http/http_parity_test.go` | [x] Completed | Refactor `NewServer` to accept `ServerDependencies` holding all 7 ports with `noop.*` fallbacks |
| **TASK-REM-09** | Verification & Memory Audit | `internal/adapter/http/http_remediation_test.go`<br>`internal/adapter/http/http_remediation_extra_test.go`<br>`.agents/state/*` | [x] Completed | Full race test (`go test -race ./...`), memory check (<30MB), update roadmap & phase status |

---

## Detailed Task Breakdown

### TASK-REM-01: Session Token Security & Entropy
- [x] Remove the fallback query `d.db.GetSession(r.Context(), token)` in `internal/adapter/http/orpc/dispatcher.go`.
- [x] Ensure all tokens extracted from headers/query are strictly hashed via `HashToken()` prior to database lookup.
- [x] Update `generateSecureToken()` in `router_setup.go` to handle error from `rand.Read(b)` and fail fast.
- [x] Update test fixtures so sessions are seeded with hashed tokens.

### TASK-REM-02: Comprehensive RBAC Lockdown
- [x] In `router_services_app.go` & `router_services_app_ops.go`: Guard `deployService`, `restartService`, `stopService`, `destroyService`, `updateEnv`, `updateDeploy`, `updateResources`, `updateSourceImage` with `c.RequireAdmin()`.
- [x] In `router_services_app_config.go`: Guard `startService`, `refreshDeployToken`, `updateSourceGit`, `updateSourceDockerfile`, and all config stubs with `c.RequireAdmin()`.
- [x] In `router_services_db.go`: Guard `enableService`, `disableService`, `updateCredentials`, `updateResources`, `destroyService`, and tool toggles with `c.RequireAdmin()`.
- [x] In `router_services_common.go`: Guard `setNotes`, `rename`, and compose lifecycle actions with `c.RequireAdmin()`.
- [x] In `router_server_users.go`: Guard `users/listUsers` with `c.RequireAdmin()`.
- [x] In `router_server_storage.go`: Guard `storageProviders/common/listStorageProviders`, backup execution, and certificates with `c.RequireAdmin()`.
- [x] In `router_projects.go`: Require authentication on `projects/inspectProject`, `listProjectsAndServices`, and `listProjects`.

### TASK-REM-03: CORS Origin Whitelisting
- [x] In `internal/adapter/http/server.go`: Validate `Origin` against request host, loopback (`localhost`, `127.0.0.1`), or configured panel domains.
- [x] Stop setting `Access-Control-Allow-Credentials: true` when origin is wildcarded or untrusted.

### TASK-REM-04: Domain Update Port Contract
- [x] In `internal/core/port/db.go`: Add `UpdateDomain(ctx context.Context, dom *domain.Domain) error`.
- [x] In `internal/adapter/db/sqlite/domains.go`: Implement `UpdateDomain` SQL update statement.
- [x] In `internal/test/mock/mock_db_domains.go`: Implement thread-safe `UpdateDomain`.
- [x] In `internal/adapter/http/router_domains.go`: Replace dummy NO-OP in `domains/updateDomain` with complete input binding, DB update, and DTO response.

### TASK-REM-05: User Update & API Token Storage
- [x] In `router_server_users.go`: Implement `users/updateUser` procedure (allows updating email, password hash, role).
- [x] In `users/generateApiToken`: Persist created API token into `sessions` table (`domain.Session{ID, UserID: c.User.ID, TokenHash: HashToken(raw), ExpiresAt: time.Now().AddDate(1, 0, 0)}`).
- [x] In `users/revokeApiToken`: Delete session record from `sessions` table.

### TASK-REM-06: Storage Providers & Notifications Parity
- [x] In `router_server_storage.go`:
  - Register procedure aliases matching frontend calls: `storageProviders/common/list`, `storageProviders/common/listOptions`.
  - Register provider-specific create/update/delete endpoints (`storageProviders/s3/*`, `storageProviders/local/*`, `storageProviders/sftp/*`).
  - Redact `secretKey`, `password`, `refreshToken` in storage provider DTOs.
- [x] In `router_actions_extra.go`:
  - Register `notifications/sendTestNotification` and `notifications/updateNotificationChannel`.
  - Redact sensitive credentials in notification channel targets.

### TASK-REM-07: RFC 6455 WebSocket Stubs
- [x] In `internal/adapter/http/server.go`: Replace broken `200 OK` response with standard RFC 6455 handshake (`101 Switching Protocols`), sending `Sec-WebSocket-Accept` and maintaining connection or closing cleanly to stop SPA reconnection storm.

### TASK-REM-08: Multi-Port Server Dependency Injection
- [x] Refactor `http.NewServer` to accept `ServerDependencies` struct.
- [x] Provide Null Object (`noop.*`) fallbacks for all nil port fields.
- [x] Pass `ProxyDriverPort` into `registerDomainsRoutes` and wire route application on create/update/delete.
- [x] Update `cmd/openpanel/main.go` and test suites to construct `ServerDependencies`.

### TASK-REM-09: Verification & Memory Audit
- [x] Add unit tests in `http_test.go`, `http_parity_test.go`, `http_remediation_test.go`, and `http_remediation_extra_test.go` verifying that unauthenticated requests to mutating routes return 401/403.
- [x] Verify `go test -race ./...` passes 100%.
- [x] Confirm idle memory footprint remains `< 30 MB` (Actual: **12.8 MB VmRSS**).
- [x] Update `.agents/state/roadmap.md`, `phase_status.md`, and `task_log.jsonl`.
