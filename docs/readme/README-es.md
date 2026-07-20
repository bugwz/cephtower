<div align="center">

# CephTower

</div>

<div align="center">

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [**Español**](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower es un proyecto con backend en Go y frontend en Vue para administrar clústeres Ceph mediante la API de Ceph Manager Dashboard.

## 1. Funciones

- Servicio HTTP en Go con endpoints de salud y resumen del clúster.
- Consola de administración basada en Vue 3, Vite y TypeScript.
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
frontend/    Consola Web Vue
docs/        Notas de arquitectura, README multilingües y referencias locales
```

## 4. Licencia

MIT License. Consulta [LICENSE](../../LICENSE).
