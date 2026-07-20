<div align="center">

# CephTower

</div>

<div align="center">

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [**Deutsch**](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower ist ein Projekt mit Go-Backend und Vue-Frontend zur Verwaltung von Ceph-Clustern über die Ceph Manager Dashboard API.

## 1. Funktionen

- Go HTTP API mit Health-Check und Cluster-Zusammenfassung.
- Verwaltungskonsole mit Vue 3, Vite und TypeScript.
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
frontend/    Vue Web-Konsole
docs/        Architekturhinweise, mehrsprachige README-Dateien und lokale Referenzen
```

## 4. Lizenz

MIT License. Siehe [LICENSE](../../LICENSE).
