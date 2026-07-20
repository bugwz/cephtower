<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-Backend-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-Frontend-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)
[![Mehrsprachig](https://img.shields.io/badge/Mehrsprachig-yellow)](../../README.md)

</div>

<div align="center">

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [**Deutsch**](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower ist ein Projekt mit Go-Backend und React-/Ant-Design-Frontend zur Verwaltung von Ceph-Clustern über die Ceph Manager Dashboard API.

## 1. Funktionen

- Go HTTP API mit Health-Check und Cluster-Zusammenfassung.
- Verwaltungskonsole mit React, Ant Design, Vite und TypeScript.
- Gekapselte Client-Schicht für Ceph Dashboard Authentifizierung und API-Aufrufe.
- MIT-Lizenz.

## 2. Schnellstart

```bash
make backend-dev
```

```bash
cd frontend
npm install
npm run dev
```

## 3. Projektstruktur

```text
backend/     Go API Service
frontend/    React Web-Konsole
  public/ceph-tower-logo.svg    Gemeinsames Logo für README und Frontend
docs/        Architekturhinweise, mehrsprachige README-Dateien und lokale Referenzen
```

## 4. Lizenz

MIT License. Siehe [LICENSE](../../LICENSE).
