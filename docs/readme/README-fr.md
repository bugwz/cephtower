<div align="center">

# CephTower

</div>

<div align="center">

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [**Français**](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower est un projet composé d'un backend Go et d'un frontend Vue pour gérer des clusters Ceph via l'API Ceph Manager Dashboard.

## 1. Fonctionnalités

- Service HTTP Go avec endpoints de santé et de résumé de cluster.
- Console d'administration basée sur Vue 3, Vite et TypeScript.
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
frontend/    Console Web Vue
docs/        Notes d'architecture, README multilingues et références locales
```

## 4. Licence

MIT License. Voir [LICENSE](../../LICENSE).
