# Workstream 2A: Docker Engine & Swarm Adapter Tasks

**Workstream**: 2A — Docker Engine & Swarm Adapter  
**Target Resource Footprint**: < 30MB idle RAM (Current measured: **12.8 MB VmRSS**)  
**File Limit Rules**: Max 300 lines (hard limit 500 lines) per Go file  
**Hexagonal Integrity**: `internal/adapter/docker/` depends exclusively on `domain` and `port` interfaces; zero cross-adapter imports  

---

## Task Matrix

| Task ID | Focus Area | Target Files | Status | Description |
|:---|:---|:---|:---|:---|
| **TASK-DOCKER-01** | Docker SDK & Adapter Core | `go.mod`<br>`internal/adapter/docker/adapter.go` | [x] Completed | Add official Docker SDK, define `DockerAdapter` struct, options, constructor, client ping and interface guard |
| **TASK-DOCKER-02** | Swarm Spec & Volume Translation | `internal/adapter/docker/services_spec.go` | [x] Completed | Translate `domain.ServiceSpec` to `swarm.ServiceSpec`, format env vars, bind/volume mounts, and file writing |
| **TASK-DOCKER-03** | Service Lifecycle Management | `internal/adapter/docker/services.go` | [x] Completed | Implement `DeployService`, `StopService`, `RestartService`, `DeleteService` with Swarm support and rolling updates |
| **TASK-DOCKER-04** | Image Building & Pulling | `internal/adapter/docker/images.go` | [x] Completed | Implement `BuildImage` (in-memory tar generator + docker build stream) and `PullImage` with registry auth |
| **TASK-DOCKER-05** | Host Metrics & Container Telemetry | `internal/adapter/docker/metrics.go` | [x] Completed | Implement `GetDockerInfo`, `GetHostMetrics` (/proc + statfs), `GetServiceStatus`, and `GetServiceStats` |
| **TASK-DOCKER-06** | Streaming Logs, Events & PTY Terminal | `internal/adapter/docker/stream.go` | [x] Completed | Implement `StreamServiceLogs`, `StreamDockerEvents`, and `ExecServiceTerminal` with PTY resize handling |
| **TASK-DOCKER-07** | Storage, Networks, Volumes & Prune | `internal/adapter/docker/storage.go` | [x] Completed | Implement `GetServiceStorage`, `ListStorageUsage`, `EnsureNetwork`, `EnsureVolume`, `ListContainers`, `PruneSystem` |
| **TASK-DOCKER-08** | Table-Driven Unit Tests | `internal/adapter/docker/*_test.go` | [x] Completed | Implement comprehensive isolated unit tests for spec translation, metrics parsing, and adapter lifecycle |
| **TASK-DOCKER-09** | Live Integration & Memory Verification | `cmd/openpanel/main.go`<br>`.agents/state/*` | [x] Completed | Wire real `DockerAdapter` into composition root (with fallback to `noop`), run full race audit and memory benchmark |

---

## Detailed Task Breakdown

### TASK-DOCKER-01: Docker SDK & Adapter Core
- Add `github.com/docker/docker` and `github.com/docker/go-connections` to `go.mod`.
- Create `internal/adapter/docker/adapter.go`.
- Implement `DockerAdapter` struct with `cli *client.Client`, `projectsDir string`, `defaultNetwork string`.
- Implement functional options (`WithHost`, `WithProjectsDir`, `WithDefaultNetwork`, `WithClient`).
- Ensure compile-time interface assertion: `var _ port.DockerPort = (*DockerAdapter)(nil)`.
- Enforce strict file size limit (< 300 lines).

### TASK-DOCKER-02: Swarm Spec & Volume Translation
- Create `internal/adapter/docker/services_spec.go`.
- Map `domain.ServiceSpec` into `swarm.ServiceSpec`:
  - `TaskTemplate.ContainerSpec`: Image, Env, Mounts, Command, Args, HealthCheck, Labels.
  - `TaskTemplate.Resources`: NanoCPUs, MemoryBytes.
  - `TaskTemplate.RestartPolicy`: RestartPolicyCondition (any, on-failure, none).
  - `UpdateConfig`: Parallelism, Order (start-first for zero downtime).
  - `Networks`: Attach to `easypanel` and project network.
- Implement volume mount handling: create host path `/etc/easypanel/projects/<project>/<service>/<volume>` for bind mounts, and write inline file contents for `type == "file"`.
- Enforce strict file size limit (< 300 lines).

### TASK-DOCKER-03: Service Lifecycle Management
- Create `internal/adapter/docker/services.go`.
- Implement `DeployService`:
  - Ensure overlay network exists.
  - Ensure volumes exist and write config files.
  - Inspect service: if exists, call `ServiceUpdate`; if not, call `ServiceCreate`.
  - Fetch task list to obtain active container IDs.
  - Return `domain.DeployResult`.
- Implement `StopService`: scale replicas to 0.
- Implement `RestartService`: force rolling restart via `ForceUpdate` counter.
- Implement `DeleteService`: remove Swarm service via `ServiceRemove`.
- Enforce strict file size limit (< 300 lines).

### TASK-DOCKER-04: Image Building & Pulling
- Create `internal/adapter/docker/images.go`.
- Implement `BuildImage`:
  - Create tarball archive from `build.ContextPath`.
  - Invoke `cli.ImageBuild` with context stream and build arguments.
  - Parse JSON build output line-by-line and format readable text to `logWriter`.
- Implement `PullImage`:
  - Encode `domain.RegistryAuth` to base64 JSON if provided.
  - Invoke `cli.ImagePull`.
  - Parse progress output and stream to `logWriter`.
- Enforce strict file size limit (< 300 lines).

### TASK-DOCKER-05: Host Metrics & Container Telemetry
- Create `internal/adapter/docker/metrics.go`.
- Implement `GetDockerInfo`: query `cli.Info(ctx)` for versions, swarm state, container and image counts.
- Implement `GetHostMetrics`: inspect Linux host statistics:
  - CPU usage from `/proc/stat`.
  - Memory usage from `/proc/meminfo` or `client.Info`.
  - Load average from `/proc/loadavg`.
  - Uptime from `/proc/uptime`.
  - Disk usage from `syscall.Statfs` on root `/` or docker root dir.
- Implement `GetServiceStatus`: inspect service and its active tasks to return `running`, `starting`, `stopped`, or `failed`.
- Implement `GetServiceStats`: query one-shot container stats via `cli.ContainerStatsOneShot`, decode JSON payload, and compute CPU percentage and memory/network usage.
- Enforce strict file size limit (< 300 lines).

### TASK-DOCKER-06: Streaming Logs, Events & PTY Terminal
- Create `internal/adapter/docker/stream.go`.
- Implement `StreamServiceLogs`: query `cli.ServiceLogs` or container logs with follow and tail options, demuxing via `stdcopy.StdCopy`.
- Implement `StreamDockerEvents`: listen to Docker daemon events via `cli.Events` and stream normalized `domain.DockerEvent` objects to the provided channel.
- Implement `ExecServiceTerminal`:
  - Find active container for service.
  - Call `cli.ContainerExecCreate` (TTY, Stdin, Stdout, Stderr).
  - Attach via `cli.ContainerExecAttach`.
  - Handle resize channel calling `cli.ContainerExecResize`.
  - Stream I/O bidirectionally until session closes.
- Enforce strict file size limit (< 300 lines).

### TASK-DOCKER-07: Storage, Networks, Volumes & Prune
- Create `internal/adapter/docker/storage.go`.
- Implement `GetServiceStorage`: compute host directory size for `/etc/easypanel/projects/<project>/<service>`.
- Implement `ListStorageUsage`: walk project directories and report all service sizes.
- Implement `EnsureNetwork`: check network existence; create overlay network with attachable flag if in Swarm mode, bridge otherwise.
- Implement `EnsureVolume`: check volume existence; create via `cli.VolumeCreate` if missing.
- Implement `ListContainers`: query `cli.ContainerList` and map to `[]domain.ContainerSummary`.
- Implement `PruneSystem`: trigger prune for containers, images, volumes, and networks.
- Enforce strict file size limit (< 300 lines).

### TASK-DOCKER-08: Table-Driven Unit Tests
- Create `internal/adapter/docker/adapter_test.go`: test configuration options and constructor.
- Create `internal/adapter/docker/services_spec_test.go`: test conversion of `domain.ServiceSpec` to Swarm spec, env var formatting, mounts, ports, and resource limits.
- Create `internal/adapter/docker/metrics_test.go`: test metrics extraction, CPU calculation algorithms, and host stats parsing.
- Create `internal/adapter/docker/storage_test.go`: test storage calculation logic and mock directory walking.
- Verify 100% race-free execution with `go test -race ./internal/adapter/docker/...`.

### TASK-DOCKER-09: Live Integration & Memory Verification
- Update `cmd/openpanel/main.go` to attempt initializing `docker.New()` and fall back to `noop.NewNoOpDocker()` if `/var/run/docker.sock` is unavailable.
- Run live test against `/var/run/docker.sock`.
- Measure idle RAM usage to verify the < 30MB constraint remains firmly satisfied.
- Update `.agents/state/phase_status.md`, `roadmap.md`, and `task_log.jsonl`.
