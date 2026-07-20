<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-Backend-00ADD8?logo=go)](../../backend/go.mod)
[![Vue](https://img.shields.io/badge/Vue-Frontend-42B883?logo=vue.js)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)
[![Multilingual](https://img.shields.io/badge/Multilingual-yellow)](../../README.md)

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
  public/ceph-tower-logo.svg    Shared README and frontend logo
docs/        Architecture notes, multilingual README files, and local references
```

## 4. License

MIT License. See [LICENSE](../../LICENSE).
