<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-Backend-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-Frontend-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)
[![Multilingue](https://img.shields.io/badge/Multilingue-yellow)](../../README.md)

</div>

<div align="center">

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [**Français**](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower est un projet composé d'un backend Go et d'un frontend React avec Ant Design pour gérer des clusters Ceph via l'API Ceph Manager Dashboard.

## 1. Fonctionnalités

- Service HTTP Go avec endpoints de santé et de résumé de cluster.
- Console d'administration basée sur React, Ant Design, Vite et TypeScript.
- Couche client dédiée à l'authentification et aux appels Ceph Dashboard API.
- Licence MIT.

## 2. Démarrage rapide

```bash
make backend-dev
```

```bash
cd frontend
npm install
npm run dev
```

## 3. Structure du projet

```text
backend/     Service API Go
frontend/    Console Web React
  public/ceph-tower-logo.svg    Logo partagé entre README et frontend
docs/        Notes d'architecture, README multilingues et références locales
```

## 4. Licence

MIT License. Voir [LICENSE](../../LICENSE).
