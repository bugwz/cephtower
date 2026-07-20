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

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [**Español**](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower es un proyecto con backend en Go y frontend en React con Ant Design para administrar clústeres Ceph mediante la API de Ceph Manager Dashboard.

## 1. Funciones

- Servicio HTTP en Go con endpoints de salud y resumen del clúster.
- Consola de administración basada en React, Ant Design, Vite y TypeScript.
- Capa cliente para autenticación y llamadas a la API de Ceph Dashboard.
- Licencia MIT.

## 2. Inicio rápido

```bash
make backend-dev
```

```bash
cd frontend
npm install
npm run dev
```

## 3. Estructura del proyecto

```text
backend/     Servicio API Go
frontend/    Consola Web React
  public/ceph-tower-logo.svg    Logo compartido entre README y frontend
docs/        Notas de arquitectura, README multilingües y referencias locales
```

## 4. Licencia

MIT License. Consulta [LICENSE](../../LICENSE).
