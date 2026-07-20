<div align="center">

# CephTower

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
docs/        Архитектура, многоязычные README файлы и локальные справочные материалы
```

## 4. Лицензия

MIT License. См. [LICENSE](../../LICENSE).
