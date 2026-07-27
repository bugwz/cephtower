<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

Consola web de administración para clústeres Ceph

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Native%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [**Español**](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower combina un backend Go y un frontend React / Ant Design para administrar uno o varios
clústeres Ceph mediante comandos Ceph y protocolos nativos. El backend ofrece una API REST
versionada, persistencia, recolección en segundo plano y una interfaz web integrada. El frontend
siempre accede al backend mediante la ruta del mismo origen `/api`.

## 1. Capacidades y estado actuales

- Asistente inicial para elegir SQLite o MySQL, probar la conexión y crear el administrador.
- Autenticación con sesiones Bearer Token de 12 horas, roles de administrador/usuario, permisos granulares de lectura y gestión de usuarios, y restablecimiento por código de correo cuando SMTP está configurado.
- Conexiones multiclúster con direcciones MON, clave `client.admin` y credenciales CephX cifradas; descubrimiento y caché automáticos de hosts, daemons, servicios, MON, MGR, MDS, OSD, módulos Mgr y configuración.
- Interfaz de clúster para conexiones, detalles, hosts, MON, MGR, OSD y MDS; incluye módulos Mgr, acciones de daemon y operaciones OSD in/out, reweight y scrub.
- Recolección configurable por módulo: fuente, intervalo, timeout, reintentos y prioridad, con ejecución manual e historial.
- Integraciones backend para clúster, Pool/RBD, CephFS/NFS/SMB, RGW, iSCSI, NVMe-oF, Prometheus/Grafana y usuarios y roles de CephTower e integraciones nativas.
- El build de producción integra el frontend en el ejecutable Go; un servicio HTTP entrega UI y API.

> [!IMPORTANT]
> El proyecto está en desarrollo activo. La gestión de clústeres y usuarios y la configuración de
> recolección usan el backend real. Las páginas de resumen e información del sistema aún contienen
> datos de demostración; las páginas de bloques, archivos, objetos y monitorización son principalmente
> marcadores de flujo. Una integración backend no implica que toda acción frontend esté terminada.

## 2. Estructura del proyecto

```text
CephTower/
├── backend/
│   ├── cmd/main.go              # entrada del proceso
│   └── internal/
│       ├── api/v1/              # rutas y handlers REST
│       ├── service/             # autenticación, clúster, recolección, ajustes e inicio
│       ├── store/               # GORM, migraciones y SQLite/MySQL
│       ├── integration/ceph/    # clientes de comandos, gateway y monitorizacion Ceph
│       ├── task/                # tareas en segundo plano y planificación
│       └── webui/               # recursos frontend integrados
├── frontend/src/                # consola React, rutas, páginas y clientes API
├── config/config.yaml           # configuración de referencia comentada
├── docs/                        # arquitectura, referencias Ceph y README traducidos
├── Makefile                     # desarrollo, pruebas y build
└── README.md
```

Consulta [docs/architecture.md](../architecture.md) para las capas y el ciclo de vida.

## 3. Requisitos

| Herramienta/servicio | Mínimo | Uso |
|---|---:|---|
| Go | 1.26 | build y pruebas del backend |
| Node.js | 20 | desarrollo y build del frontend |
| npm | 10 | dependencias del frontend |
| Toolchain C | apropiada para el SO | necesaria para el driver SQLite CGO |
| Ceph | Ceph 20.2.2+ | también requiere direcciones MON y un CephX client keys con permisos suficientes |
| MySQL | opcional | solo si no se usa SQLite por defecto |

## 4. Inicio rápido

Desde la raíz del repositorio:

```bash
make run
```

El comando comprueba el entorno, instala dependencias si es necesario y crea
`app/config/config.yaml` desde `config/config.yaml` si falta (directorio de ejecución `./app`). Inicia:

- Backend y entrada web de producción: <http://localhost:36900>
- Servidor Vite: <http://localhost:36901> (`/api` se redirige al backend)

La primera visita va a `/initialize`. Configura la base y el administrador y añade una conexión Ceph.
Para iniciar por separado, ejecuta primero `make ensure-run-config` y después en dos terminales:

```bash
make run-backend
make run-frontend
```

### Build de producción

```bash
make build
./bin/cephtower -config /path/to/config.yaml
```

El ejecutable queda en `bin/cephtower`. Sin `-config` lee
`/opt/cephtower/config/config.yaml`, que debe existir antes del inicio.

## 5. Configuración

Consulta [config/config.yaml](../../config/config.yaml) para todas las opciones y valores por defecto.

| Sección | Uso |
|---|---|
| `server` | dirección, puerto y directorio de ejecución (por defecto `0.0.0.0:36900`, `/opt/cephtower`) |
| `log` | salida, nivel, formato, rotación y retención |
| `runtime` | configuración Ceph, CephX client keys y otros archivos de ejecución |
| `database` | SQLite o conexión/TLS MySQL; migraciones automáticas al iniciar |
| `smtp` | correo opcional para restablecer contraseñas |

Las credenciales Ceph no se guardan en este YAML: tras la inicialización se almacenan en la base
desde la gestión de clústeres. Restringe el acceso a configuración, base y archivos de ejecución, y
usa validación TLS adecuada en producción.

## 6. Comandos habituales

| Comando | Uso |
|---|---|
| `make check-env` | comprobar versiones de Go, Node.js y npm |
| `make run` | iniciar juntos backend y frontend de desarrollo |
| `make run-backend` | compilar/iniciar backend; `CONFIG` elige la configuración |
| `make run-frontend` | iniciar Vite en el puerto `36901` |
| `make build` | compilar frontend y crear `bin/cephtower` con UI integrada |
| `make build-frontend` | comprobar tipos, compilar y sincronizar recursos integrados |
| `make test` | ejecutar pruebas backend y validar el build frontend |
| `make test-backend` | ejecutar `go test ./...` |
| `make test-frontend` | comprobar tipos y compilar el frontend |

Usa `CONFIG=/path/to/config.yaml` para la configuración backend o `FRONTEND_PORT=puerto` para
el puerto frontend de `make run`.

## 7. API y documentación

El prefijo es `/api/v1`. Endpoints básicos sin autenticación:

| Método | Ruta | Uso |
|---|---|---|
| `GET` | `/api/v1/healthz` | estado vital del proceso |
| `GET` | `/api/v1/readyz` | preparación tras inicialización |
| `GET` | `/api/v1/setup/status` | estado del primer inicio |
| `POST` | `/api/v1/auth/login` | iniciar sesión y obtener Token |

Salvo inicio, login y restablecimiento, se requiere `Authorization: Bearer <token>`.
Las rutas están en `backend/internal/api/v1/router/`. Alcance y compatibilidad Ceph:
[docs/ceph/apis/index.md](../ceph/apis/index.md).

## 8. Desarrollo y contribución

- Ejecuta `make test-backend` para backend y `make test-frontend` para frontend.
- No confirmes datos, bases, logs ni claves de clúster locales de `app/`.
- Sigue [docs/commit-convention.md](../commit-convention.md) para los commits.
- Issues y Pull Requests son bienvenidos; distingue funciones verificadas y marcadores.

## 9. Licencia

CephTower se distribuye bajo la [licencia MIT](../../LICENSE).
