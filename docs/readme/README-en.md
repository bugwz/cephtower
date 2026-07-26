<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

Web management console for Ceph clusters

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [**English**](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower combines a Go backend with a React/Ant Design frontend to manage one or more
Ceph clusters through the Ceph Dashboard API and Ceph commands. The backend provides a
versioned REST API, persistence, background collection, and an embedded web UI. The
frontend always accesses the backend through the same-origin `/api` path.

## 1. Current capabilities and status

- First-run wizard for selecting SQLite or MySQL, testing the connection, and creating an administrator.
- Authentication with 12-hour Bearer Token sessions, administrator/user roles, granular read and user-management permissions, and email-code password reset when SMTP is configured.
- Multi-cluster connections storing MON addresses, `client.admin` keys, and Dashboard credentials; automatic discovery and caching of hosts, daemons, services, MONs, MGRs, MDSs, OSDs, Mgr modules, and cluster configuration.
- Cluster UI for connections and details, hosts, MON, MGR, OSD, and MDS, including Mgr module toggles, daemon actions, and OSD in/out, reweight, and scrub operations.
- Per-module data collection settings for source, interval, timeout, retry, and priority, with manual runs and run history.
- Backend integrations for clusters, Pool/RBD, CephFS/NFS/SMB, RGW, iSCSI, NVMe-oF, Prometheus/Grafana, and Dashboard users, roles, and configuration.
- Production builds embed the frontend in the Go executable so one HTTP service delivers both UI and API.

> [!IMPORTANT]
> The project is under active development. Cluster management, user management, and data
> collection settings use the real backend. The overview and system-information pages still
> contain demo data, while block, file, object, and monitoring pages mainly show workflow
> placeholders. An existing backend integration does not mean every frontend action is complete.

## 2. Project layout

```text
CephTower/
├── backend/
│   ├── cmd/main.go              # process entry point
│   └── internal/
│       ├── api/v1/              # REST routes and handlers
│       ├── service/             # auth, cluster, collector, settings, and setup logic
│       ├── store/               # GORM, migrations, and SQLite/MySQL storage
│       ├── integration/ceph/    # Ceph Dashboard and command clients
│       ├── task/                # background jobs and scheduling
│       └── webui/               # embedded frontend assets
├── frontend/src/                # React console, routes, pages, and API clients
├── config/config.yaml           # fully commented reference configuration
├── docs/                        # architecture, Ceph references, and translated READMEs
├── Makefile                     # development, test, and build entry points
└── README.md
```

See [docs/architecture.md](../architecture.md) for layering and lifecycle details.

## 3. Requirements

| Tool/service | Minimum | Purpose |
|---|---:|---|
| Go | 1.26 | backend builds and tests |
| Node.js | 20 | frontend development and builds |
| npm | 10 | frontend dependency management |
| C toolchain | OS-appropriate | required by the CGO SQLite driver |
| Ceph | Dashboard API enabled | connection also requires MON addresses and a sufficiently privileged keyring |
| MySQL | optional | only when not using the default SQLite database |

## 4. Quick start

From the repository root, run:

```bash
make run
```

This checks the environment, installs frontend dependencies when needed, creates
`app/config/config.yaml` from `config/config.yaml` if missing (using `./app` as the development
runtime directory), and starts:

- Backend and production web entry point: <http://localhost:36900>
- Vite development server: <http://localhost:36901> (`/api` proxies to the backend)

The first visit redirects to `/initialize`. Configure the database and administrator, then add
a Ceph connection under cluster management. To start both services separately, run
`make ensure-run-config` first and then use two terminals:

```bash
make run-backend
make run-frontend
```

### Production build

```bash
make build
./bin/cephtower -config /path/to/config.yaml
```

The executable is written to `bin/cephtower`. Without `-config`, it reads
`/opt/cephtower/config/config.yaml`; that file must exist before startup.

## 5. Configuration

See [config/config.yaml](../../config/config.yaml) for all options and defaults.

| Section | Purpose |
|---|---|
| `server` | listen address, port, and runtime directory (defaults: `0.0.0.0:36900`, `/opt/cephtower`) |
| `log` | output, level, format, rotation, and retention |
| `runtime` | directory for Ceph configuration, keyrings, and other runtime files |
| `database` | SQLite file or MySQL connection/TLS settings; migrations run at startup |
| `smtp` | optional mail service for password resets |

Ceph cluster credentials are not stored in this YAML. They are saved to the database through
cluster management after initialization. Restrict access to the configuration, database, and
runtime directories, and use appropriate TLS validation in production.

## 6. Common commands

| Command | Purpose |
|---|---|
| `make check-env` | validate Go, Node.js, and npm versions |
| `make run` | start the development backend and frontend together |
| `make run-backend` | build and start the backend; use `CONFIG` to select a config file |
| `make run-frontend` | start Vite on port `36901` |
| `make build` | build the frontend and create `bin/cephtower` with the embedded UI |
| `make build-frontend` | type-check/build the frontend and synchronize embedded assets |
| `make test` | run the backend tests and frontend build validation |
| `make test-backend` | run `go test ./...` |
| `make test-frontend` | type-check and build the frontend for validation |

Override the backend configuration with `CONFIG=/path/to/config.yaml`, or the frontend port
used by `make run` with `FRONTEND_PORT=port`.

## 7. API and documentation

The API prefix is `/api/v1`. Basic unauthenticated endpoints include:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/api/v1/healthz` | process liveness |
| `GET` | `/api/v1/readyz` | initialization readiness |
| `GET` | `/api/v1/setup/status` | first-run status |
| `POST` | `/api/v1/auth/login` | sign in and obtain a token |

Except for setup, login, and password-reset endpoints, requests require
`Authorization: Bearer <token>`. Route definitions live in `backend/internal/api/v1/router/`.
See [docs/ceph/apis/index.md](../ceph/apis/index.md) for Ceph integration scope and compatibility.

## 8. Development and contribution

- Run `make test-backend` for backend changes and `make test-frontend` for frontend changes.
- Do not commit local runtime data, databases, logs, or cluster keys from `app/`.
- Follow [docs/commit-convention.md](../commit-convention.md) for commit messages.
- Issues and pull requests are welcome; clearly identify verified and placeholder functionality.

## 9. License

CephTower is available under the [MIT License](../../LICENSE).
