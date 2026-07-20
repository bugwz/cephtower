<div align="center">

# CephTower

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
docs/        架構說明、多語言 README 文件與本地參考資料
```

## 4. 開源協議

MIT License。詳見 [LICENSE](../../LICENSE)。
