<div align="center">

# CephTower

</div>

<div align="center">

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [**English**](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower is a Go backend and Vue frontend project for managing Ceph clusters through the Ceph Manager Dashboard API.

## 1. Features

- Go HTTP API service with health check and cluster summary endpoints.
- Vue 3, Vite, and TypeScript based management console.
- Ceph Dashboard API client boundary for authentication and future cluster operations.
- MIT licensed.

## 2. Quick Start

```bash
make backend-dev
```

```bash
cd frontend
npm install
npm run dev
```

## 3. Project Layout

```text
backend/     Go API service
frontend/    Vue web console
docs/        Architecture notes, multilingual README files, and local references
```

## 4. License

MIT License. See [LICENSE](../../LICENSE).
