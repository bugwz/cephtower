<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-Backend-00ADD8?logo=go)](../../backend/go.mod)
[![Vue](https://img.shields.io/badge/Vue-Frontend-42B883?logo=vue.js)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)
[![多語言](https://img.shields.io/badge/多語言-yellow)](../../README.md)

</div>

<div align="center">

[简体中文](../../README.md) | [**繁體中文**](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower 是一個使用 Go 後端與 Vue 前端建構的 Ceph 叢集管理專案，透過 Ceph Manager Dashboard API 管理整個 Ceph 叢集。

## 1. 功能特性

- Go HTTP API 服務，提供健康檢查與叢集摘要介面。
- Vue 3、Vite、TypeScript 管理控制台。
- 封裝 Ceph Dashboard API 認證與後續叢集操作邊界。
- 採用 MIT 開源協議。

## 2. 快速開始

```bash
make backend-dev
```

```bash
cd frontend
npm install
npm run dev
```

## 3. 專案結構

```text
backend/     Go API 服務
frontend/    Vue Web 控制台
  public/ceph-tower-logo.svg    README 與前端共用 logo
docs/        架構說明、多語言 README 文件與本地參考資料
```

## 4. 開源協議

MIT License。詳見 [LICENSE](../../LICENSE)。
