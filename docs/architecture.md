# Architecture

Cephtower is split into two applications:

- `backend`: exposes a stable management API for the web console and talks to the Ceph Manager Dashboard API.
- `frontend`: provides the React and Ant Design based cluster operations console.

The backend keeps Ceph Dashboard authentication and API compatibility concerns behind `internal/integrations/ceph`. The frontend should call only Cephtower backend routes under `/api`.

## Initial Backend Routes

- `GET /healthz`: process health check.
- `GET /api/v1/cluster/summary`: cluster overview endpoint.

## Ceph Dashboard Integration

The Ceph Dashboard API base URL, credentials, and future frontend-facing settings are configured through the single YAML configuration file under the root `config/` directory. Production deployments should protect configuration files that contain secrets and keep TLS verification enabled.
