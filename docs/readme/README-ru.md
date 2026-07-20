<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-Backend-00ADD8?logo=go)](../../backend/go.mod)
[![Vue](https://img.shields.io/badge/Vue-Frontend-42B883?logo=vue.js)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)
[![Многоязычный](https://img.shields.io/badge/Многоязычный-yellow)](../../README.md)

</div>

<div align="center">

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [**Русский**](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower — это проект с backend на Go и frontend на Vue для управления кластерами Ceph через Ceph Manager Dashboard API.

## 1. Возможности

- HTTP API на Go с проверкой состояния и сводкой кластера.
- Консоль управления на Vue 3, Vite и TypeScript.
- Клиентский слой для аутентификации и вызовов Ceph Dashboard API.
- Лицензия MIT.

## 2. Быстрый старт

```bash
make backend-dev
```

```bash
cd frontend
npm install
npm run dev
```

## 3. Структура проекта

```text
backend/     Go API сервис
frontend/    Vue Web консоль
  public/ceph-tower-logo.svg    Общий логотип для README и frontend
docs/        Архитектура, многоязычные README файлы и локальные справочные материалы
```

## 4. Лицензия

MIT License. См. [LICENSE](../../LICENSE).
