# Architecture

CephTower consists of a Go backend and a React frontend. The backend exposes the stable
`/api/v1` management API, serves the embedded frontend in packaged deployments, collects
Ceph state in the background, and communicates with Ceph through Dashboard HTTP APIs and
CLI commands.

## Backend layout

```text
backend/
├── cmd/main.go                  process entry point
└── internal/
    ├── app/                     dependency wiring and lifecycle
    ├── api/                     HTTP server and cross-version middleware
    │   └── v1/
    │       ├── router/          v1 methods, paths and route registration
    │       └── handler/         v1 HTTP input, DTOs and response mapping
    ├── service/                 application business operations
    │   ├── auth/                users, sessions and password reset
    │   ├── cluster/             managed cluster lifecycle and discovery
    │   ├── collector/           scheduled and manual collection entry point
    │   ├── settings/            system and Ceph settings orchestration
    │   ├── setup/               first-run database and admin initialization
    │   └── cephproxy/           cache-aware Ceph Dashboard proxy facade
    ├── task/                    scheduling, task submission and shutdown
    ├── store/                   GORM models, migrations and database access
    ├── integration/ceph/        Ceph Dashboard and command clients
    ├── config/                  configuration loading and validation
    ├── logging/                 logging, rotation and retention cleanup
    └── webui/                   embedded frontend assets
```

The executable remains at `backend/cmd/main.go`; the project intentionally has no extra
single-binary subdirectory below `cmd`.

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
HTTP handlers must not start background goroutines or own database transactions. Scheduled
and manually triggered collection both enter through the task manager and use the same
collector service. Long-running components are created and stopped by `internal/app`.
`internal/app` is also the only composition root: the API server receives constructed
services and does not open databases or construct Ceph clients.

The store package is the persistence boundary and is the only package that directly imports
GORM. It exposes a project-owned database handle, models, migration setup, and synchronized
runtime replacement through `store.Manager`. Ceph Dashboard authentication, API path
compatibility, and CLI execution remain isolated under `internal/integration/ceph`.

## Runtime lifecycle

At startup, the application loads configuration, installs logging, creates the task manager,
opens the database, synchronizes Ceph runtime files, registers recurring jobs, and starts the
HTTP listener. SIGINT or SIGTERM stops HTTP acceptance first, cancels and waits for managed
tasks, closes the active database, and finally closes log files.

The process liveness endpoint reports whether CephTower itself is running. Ceph cluster
health is business data and does not determine process liveness.

The versioned API uses standard HTTP status codes (for example 201, 400, 404 and 409) and
returns the same status in its JSON error envelope. Routes in `internal/api/v1/router` and
DTO/response mapping in `internal/api/v1/handler` jointly define the API contract; no legacy
transport or DTO compatibility package is retained.
