<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

Веб-консоль управления кластерами Ceph

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [**Русский**](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower объединяет backend на Go и frontend на React / Ant Design для управления одним или
несколькими кластерами Ceph через Ceph Dashboard API и команды Ceph. Backend предоставляет
версионированный REST API, хранение данных, фоновый сбор и встроенный Web UI. Frontend всегда
обращается к backend по same-origin пути `/api`.

## 1. Текущие возможности и состояние

- Мастер первого запуска: выбор SQLite/MySQL, проверка соединения и создание администратора.
- Аутентификация: 12-часовые сессии Bearer Token, роли администратора/пользователя, раздельные права чтения и управления пользователями; сброс пароля по коду e-mail при настройке SMTP.
- Подключение нескольких кластеров: хранение адресов MON, ключа `client.admin` и данных Dashboard; автоматическое обнаружение и кэширование хостов, демонов, сервисов, MON, MGR, MDS, OSD, модулей Mgr и конфигурации.
- Интерфейс кластера: подключения и сведения, хосты, MON, MGR, OSD и MDS; переключение модулей Mgr, действия с демонами и операции OSD in/out, reweight и scrub.
- Сбор данных: источник, интервал, тайм-аут, повторы и приоритет по модулям, ручной запуск и история.
- Интеграции backend: кластер, Pool/RBD, CephFS/NFS/SMB, RGW, iSCSI, NVMe-oF, Prometheus/Grafana, а также пользователи, роли и настройки Dashboard.
- Production-сборка встраивает frontend в исполняемый файл Go; один HTTP-сервис отдает UI и API.

> [!IMPORTANT]
> Проект активно разрабатывается. Управление кластерами и пользователями и настройки сбора работают
> с реальным backend. Обзор и системная информация пока содержат демонстрационные данные; страницы
> блочного, файлового, объектного хранилища и мониторинга в основном являются заготовками workflow.
> Наличие интеграции backend не означает готовность всех действий frontend.

## 2. Структура проекта

```text
CephTower/
├── backend/
│   ├── cmd/main.go              # точка входа процесса
│   └── internal/
│       ├── api/v1/              # REST-маршруты и обработчики
│       ├── service/             # авторизация, кластеры, сбор, настройки и инициализация
│       ├── store/               # GORM, миграции и SQLite/MySQL
│       ├── integration/ceph/    # клиенты Dashboard и команд Ceph
│       ├── task/                # фоновые задачи и планирование
│       └── webui/               # встроенные ресурсы frontend
├── frontend/src/                # React-консоль, маршруты, страницы и API-клиенты
├── config/config.yaml           # комментированная эталонная конфигурация
├── docs/                        # архитектура, материалы Ceph и переводы README
├── Makefile                     # разработка, тесты и сборка
└── README.md
```

Слои и жизненный цикл описаны в [docs/architecture.md](../architecture.md).

## 3. Требования

| Инструмент/сервис | Минимум | Назначение |
|---|---:|---|
| Go | 1.26 | сборка и тесты backend |
| Node.js | 20 | разработка и сборка frontend |
| npm | 10 | зависимости frontend |
| C toolchain | подходящий для ОС | требуется драйвером SQLite CGO |
| Ceph | Dashboard API включен | также нужны адреса MON и keyring с достаточными правами |
| MySQL | необязательно | только вместо SQLite по умолчанию |

## 4. Быстрый старт

В корне репозитория:

```bash
make run
```

Команда проверяет окружение, при необходимости устанавливает зависимости и создает
`app/config/config.yaml` из `config/config.yaml`, если файл отсутствует (рабочий каталог `./app`). Запускаются:

- Backend и production Web: <http://localhost:36900>
- Сервер Vite: <http://localhost:36901> (`/api` проксируется к backend)

Первое посещение перенаправляет на `/initialize`. Настройте БД и администратора, затем добавьте
Ceph-подключение. Для раздельного запуска сначала выполните `make ensure-run-config`, затем в двух терминалах:

```bash
make run-backend
make run-frontend
```

### Production-сборка

```bash
make build
./bin/cephtower -config /path/to/config.yaml
```

Файл создается как `bin/cephtower`. Без `-config` читается
`/opt/cephtower/config/config.yaml`, который должен существовать до запуска.

## 5. Конфигурация

Все параметры и значения по умолчанию: [config/config.yaml](../../config/config.yaml).

| Раздел | Назначение |
|---|---|
| `server` | адрес, порт и рабочий каталог (по умолчанию `0.0.0.0:36900`, `/opt/cephtower`) |
| `log` | вывод, уровень, формат, ротация и срок хранения |
| `runtime` | конфигурация Ceph, keyrings и прочие runtime-файлы |
| `database` | SQLite или соединение/TLS MySQL; миграции выполняются при запуске |
| `smtp` | необязательная почта для сброса пароля |

Учетные данные Ceph не хранятся в YAML: после инициализации они записываются в БД через управление
кластерами. Ограничьте доступ к конфигурации, БД и runtime-файлам и используйте корректную проверку TLS.

## 6. Основные команды

| Команда | Назначение |
|---|---|
| `make check-env` | проверить версии Go, Node.js и npm |
| `make run` | вместе запустить backend и frontend для разработки |
| `make run-backend` | собрать/запустить backend; файл выбирается через `CONFIG` |
| `make run-frontend` | запустить Vite на порту `36901` |
| `make build` | собрать frontend и создать `bin/cephtower` со встроенным UI |
| `make build-frontend` | проверить типы, собрать и синхронизировать встроенные ресурсы |
| `make test` | запустить тесты backend и проверку сборки frontend |
| `make test-backend` | выполнить `go test ./...` |
| `make test-frontend` | проверить типы и собрать frontend |

Используйте `CONFIG=/path/to/config.yaml` для конфигурации backend или `FRONTEND_PORT=порт`
для порта frontend команды `make run`.

## 7. API и документация

Префикс API — `/api/v1`. Основные endpoints без аутентификации:

| Метод | Путь | Назначение |
|---|---|---|
| `GET` | `/api/v1/healthz` | проверка процесса |
| `GET` | `/api/v1/readyz` | готовность после инициализации |
| `GET` | `/api/v1/setup/status` | состояние первого запуска |
| `POST` | `/api/v1/auth/login` | вход и получение Token |

Кроме инициализации, входа и сброса пароля требуется `Authorization: Bearer <token>`.
Маршруты находятся в `backend/internal/api/v1/router/`. Область и совместимость Ceph:
[docs/ceph/apis/index.md](../ceph/apis/index.md).

## 8. Разработка и участие

- Для backend запускайте `make test-backend`, для frontend — `make test-frontend`.
- Не коммитьте локальные данные, БД, журналы и ключи кластеров из `app/`.
- Соблюдайте [docs/commit-convention.md](../commit-convention.md).
- Issues и Pull Requests приветствуются; четко отмечайте проверенные функции и заготовки.

## 9. Лицензия

CephTower распространяется по [лицензии MIT](../../LICENSE).
