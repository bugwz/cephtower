# Architecture

CephTower consists of a Go backend and a React frontend. The backend exposes the stable
`/api/v1` management API, serves the embedded frontend in packaged deployments, collects
Ceph state in the background, and communicates with Ceph through native commands, gateway
protocols, and monitoring HTTP APIs.

## Backend layout

```text
backend/
├── cmd/main.go                  process entry point
├── tools/                       backend-scoped development tools
│   ├── ceph-cmds/               Ceph CLI command reference generator
│   ├── config-init/             run configuration bootstrap helper
│   ├── dash-client/             Ceph Dashboard client generator
│   ├── dash-docs/               Ceph Dashboard API documentation generator
│   └── openapi-gen/             OpenAPI contract generator/checker
└── internal/
    ├── app/                     dependency wiring and lifecycle
    ├── api/                     HTTP server and cross-version middleware
    │   └── v1/
    │       ├── router/          v1 methods, paths and route registration
    │       └── handler/         v1 HTTP input, DTOs and response mapping
    ├── service/                 application business operations
    │   ├── auth/                users, sessions and password reset
    │   ├── cluster/             managed cluster lifecycle and discovery
    │   ├── operation/           durable queue, cluster serialization, recovery, and audit
    │   ├── reconciler/          current-state collection and stale-data handling
    │   ├── endpoint/            encrypted external endpoint configuration
    │   ├── external/            protocol-native gateway and monitoring integration
    │   └── mutation/            native Ceph/RBD/RGW/CephFS command mutations
    ├── task/                    scheduling, task submission and shutdown
    ├── store/                   GORM models, migrations and database access
    ├── integration/ceph/        Ceph command, gateway, and monitoring clients
    ├── config/                  configuration loading and validation
    ├── logging/                 logging, rotation and retention cleanup
    └── webui/                   embedded frontend assets
```

The executable remains at `backend/cmd/main.go`; development helpers live under
`backend/tools` so they can access `backend/internal` without adding extra runtime entries
below `cmd`.

## Dependency direction

```text
cmd → app
app → api, task, service, store, integration
api/server → api/v1/router → api/v1/handler
api/v1/handler → service, task
task → service
service → store, integration
```

The versioned router package only maps HTTP methods and paths to handler methods. It may
depend on `v1/handler`, while the handler package must never import the router package.
HTTP handlers do not own database transactions. Scheduled reconciliation and manual refresh
use the same reconciler; manual refresh is queued as a durable operation. Long-running
components are created and stopped by `internal/app`.
`internal/app` is also the only composition root: the API server receives constructed
services and does not open databases or construct Ceph clients.

The store package is the persistence boundary and is the only package that directly imports
GORM. It exposes a project-owned database handle, models, migration setup, and synchronized
runtime replacement through `store.Manager`. Command execution and external gateway protocol
differences remain isolated under `internal/integration/ceph`; the application does not
construct a Dashboard runtime client.

## Runtime lifecycle

At startup, the application loads configuration, installs logging, opens and migrates the
database, starts durable operation workers and reconcilers, and starts the HTTP listener.
SIGINT or SIGTERM stops HTTP acceptance first, cancels workers and reconcilers, closes the
active database, and finally closes log files.

The process liveness endpoint reports whether CephTower itself is running. Ceph cluster
health is business data and does not determine process liveness.

## Ceph access and reconciliation

Core state comes from allowlisted argv-based Ceph, RBD, RGW admin, and CephFS commands. Each
execution materializes a mode-0600 temporary config and keyring, runs without a shell, applies
timeouts and output limits, and removes the temporary directory when the command finishes.
CephX keys are decrypted only for this in-memory execution scope.

The reconciler stores each resource kind in its dedicated `ceph_<kind>` table. Cluster and host
configuration use `ceph_cluster` and `ceph_host`; internal discovery payloads are stored in
`discovered_data`, parsed into explicit API DTO fields, and never returned as raw storage fields.
Its six independently scheduled modules are `ceph_auth`, `fast`, `topology`, `storage`,
`inventory`, and `configuration`; their intervals are configured under `collection.intervals`.
A collection result explicitly records which kinds were authoritative: a successful empty result marks
disappeared resources stale, while a failed optional command preserves the last known rows.
Successful mutations immediately reconcile the owning module before the operation is marked
complete.

Dedicated resource tables contain only the latest Ceph observation. `overview` and
`health_check` additionally write five-minute samples to `ceph_observation_history`, retain
them for 90 days, and expose them through `/api/v1/resource/history`. Collection run metadata
is retained for 30 days. These policies are explicit; request bodies are never merged into
observed resource state.

Prometheus, Alertmanager, Grafana, S3 bucket configuration, iSCSI, and NVMe-oF reads use their
native HTTP, S3, or gRPC protocols. Endpoint credentials and custom CA/mTLS material are stored
in encrypted cluster credentials; TLS verification is enabled by default. No Dashboard runtime
client or arbitrary path proxy participates in either read or write flows.

## Operations and API contract

Ceph mutations return `202 Accepted` with a durable operation. Workers serialize commands per
cluster, recheck optimistic resource versions immediately before execution, run command
post-checks, reconcile successful mutations, and recover interrupted state after restart.
Only idempotent collection refreshes retry automatically; mutation commands remain
single-attempt because their outcome may be uncertain. Completed operation rows are retained
for 90 days. Accepted, started, retrying, and completed outcomes are appended to a redacted
audit hash chain.

Request contracts are shared by runtime validation and the OpenAPI generator. Unknown fields,
wrong JSON types, and action-specific invalid values are rejected before enqueue. Every JSON
response and every SSE data event has exactly `code`, `message`, and `data` at the top level.
OpenAPI binds each success response to a concrete envelope for clusters, resources, operations,
roles, endpoints, credentials, audit data, and external protocol results. Cluster-scoped role
bindings organize users without sharing Ceph credentials.

The versioned API uses standard HTTP status codes (for example 201, 400, 404 and 409) and
returns the same status in its JSON error envelope. Routes in `internal/api/v1/router` and
DTO/response mapping in `internal/api/v1/handler` jointly define the API contract; no legacy
transport or DTO compatibility package is retained.
