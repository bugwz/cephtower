<div align="center">

<img src="frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

Ceph 集群 Web 管理控制台

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](backend/go.mod)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

[**简体中文**](README.md) | [繁體中文](docs/readme/README-zh-TW.md) | [English](docs/readme/README-en.md) | [日本語](docs/readme/README-ja.md) | [Français](docs/readme/README-fr.md) | [Deutsch](docs/readme/README-de.md) | [Español](docs/readme/README-es.md) | [Português](docs/readme/README-pt.md) | [Русский](docs/readme/README-ru.md) | [한국어](docs/readme/README-ko.md)

</div>

CephTower 使用 Go 后端和 React / Ant Design 前端，通过 Ceph Dashboard API 与
Ceph 命令管理一个或多个 Ceph 集群。后端提供版本化 REST API、持久化、后台采集
任务和内嵌 Web UI，前端始终通过同源 `/api` 访问后端。

## 1. 当前能力与状态

- 首次启动向导：选择 SQLite 或 MySQL，测试连接并创建管理员账户。
- 身份认证：12 小时 Bearer Token 会话、管理员/普通用户、细粒度读取与用户管理权限；
  配置 SMTP 后可使用邮件验证码重置密码。
- 多集群连接：保存 MON 地址、`client.admin` 密钥和 Dashboard 凭据，自动发现并缓存
  主机、守护进程、服务、MON、MGR、MDS、OSD、Mgr 模块与集群配置。
- 集群界面：集群连接与详情、主机、MON、MGR、OSD 和 MDS 管理；支持 Mgr 模块开关、
  守护进程操作以及 OSD in/out、reweight 和 scrub 等操作。
- 数据采集：按模块配置采集来源、周期、超时、重试与优先级，也可立即运行并查看记录。
- 后端集成：覆盖集群、Pool/RBD、CephFS/NFS/SMB、RGW、iSCSI、NVMe-oF、
  Prometheus/Grafana、Dashboard 用户/角色与配置等 API。
- 交付方式：生产构建将前端产物嵌入 Go 可执行文件，由同一 HTTP 服务提供 UI 和 API。

> [!IMPORTANT]
> 项目仍在开发中。集群管理、用户管理和采集配置已经连接真实后端；总览和系统信息
> 目前包含演示数据，块/文件/对象存储及监控页面主要展示工作流占位内容。后端已具备的
> 集成接口不代表所有前端操作都已完成。

## 2. 项目结构

```text
CephTower/
├── backend/
│   ├── cmd/main.go              # 进程入口
│   └── internal/
│       ├── api/v1/              # REST 路由与处理器
│       ├── service/             # 认证、集群、采集、设置与初始化业务
│       ├── store/               # GORM、迁移及 SQLite/MySQL 存储
│       ├── integration/ceph/    # Ceph Dashboard 与命令客户端
│       ├── task/                # 后台任务与调度
│       └── webui/               # 内嵌前端资源
├── frontend/src/                # React 控制台、路由、页面与 API 客户端
├── config/config.yaml           # 带完整注释的参考配置
├── docs/                        # 架构、Ceph API/命令资料和多语言 README
├── Makefile                     # 开发、测试与构建入口
└── README.md
```

详细分层和生命周期说明见 [docs/architecture.md](docs/architecture.md)。

## 3. 环境要求

| 工具/服务 | 最低要求 | 说明 |
|---|---:|---|
| Go | 1.26 | 后端构建和测试 |
| Node.js | 20 | 前端开发和构建 |
| npm | 10 | 前端依赖管理 |
| C 编译工具链 | 系统适配版本 | SQLite 驱动使用 CGO |
| Ceph | 启用 Dashboard API | 集群接入还需 MON 地址和具有足够权限的 keyring |
| MySQL | 可选 | 不使用默认 SQLite 时需要 |

## 4. 快速开始

在仓库根目录运行：

```bash
make run
```

该命令会检查环境、按需安装前端依赖，并在缺少时从 `config/config.yaml` 创建
`app/config/config.yaml`（开发运行目录改为 `./app`），随后启动：

- 后端与生产 Web 入口：<http://localhost:36900>
- Vite 开发服务器：<http://localhost:36901>（`/api` 代理到后端）

首次访问会跳转至 `/initialize`。完成数据库和管理员初始化后，在集群管理中添加 Ceph
连接。若要分别启动服务，先运行 `make ensure-run-config`，再在两个终端中运行：

```bash
make run-backend
make run-frontend
```

### 生产构建

```bash
make build
./bin/cephtower -config /path/to/config.yaml
```

可执行文件位于 `bin/cephtower`。未传入 `-config` 时默认读取
`/opt/cephtower/config/config.yaml`；配置文件必须在进程启动前存在。

## 5. 配置

完整选项和默认值以 [config/config.yaml](config/config.yaml) 为准：

| 配置段 | 用途 |
|---|---|
| `server` | 监听地址、端口和运行目录（默认 `0.0.0.0:36900`、`/opt/cephtower`） |
| `log` | 输出目标、级别、格式、轮转与保留时间 |
| `runtime` | Ceph 配置和 keyring 等运行时文件目录 |
| `database` | SQLite 文件或 MySQL 连接与 TLS 选项；启动时自动迁移 |
| `smtp` | 可选的密码重置邮件服务 |

Ceph 集群凭据不写在该 YAML 中，而是在初始化完成后通过集群管理保存到数据库。请限制
配置、数据库和运行时目录的访问权限，并在生产环境中启用适当的 TLS 校验。

## 6. 常用命令

| 命令 | 作用 |
|---|---|
| `make check-env` | 检查 Go、Node.js 和 npm 版本 |
| `make run` | 同时启动开发后端和前端 |
| `make run-backend` | 构建并启动后端，使用 `CONFIG` 指定配置路径 |
| `make run-frontend` | 在 `36901` 端口启动 Vite |
| `make build` | 构建前端并生成内嵌 UI 的 `bin/cephtower` |
| `make build-frontend` | 类型检查、构建前端并同步内嵌资源 |
| `make test` | 运行后端测试和前端构建校验 |
| `make test-backend` | 运行 `go test ./...` |
| `make test-frontend` | 执行前端类型检查和 Vite 构建校验 |
| `ruby tools/generate_ceph_dashboard_client.rb` | 从本地资料重新生成 Ceph Dashboard 客户端代码 |

可通过 `CONFIG=/path/to/config.yaml` 覆盖后端配置，通过 `FRONTEND_PORT=端口` 覆盖
`make run` 使用的前端端口。

## 7. API 与文档

API 前缀为 `/api/v1`。无需认证的基础端点包括：

| 方法 | 路径 | 用途 |
|---|---|---|
| `GET` | `/api/v1/healthz` | 进程存活检查 |
| `GET` | `/api/v1/readyz` | 初始化就绪检查 |
| `GET` | `/api/v1/setup/status` | 首次启动状态 |
| `POST` | `/api/v1/auth/login` | 登录并获取 Token |

除初始化、登录和密码重置端点外，API 请求需要
`Authorization: Bearer <token>`。路由源码位于 `backend/internal/api/v1/router/`；Ceph
集成范围与兼容性说明见 [docs/ceph/apis/index.md](docs/ceph/apis/index.md)。

## 8. 开发与贡献

- 后端改动运行 `make test-backend`；前端改动运行 `make test-frontend`。
- 不要提交 `app/` 中的本地运行数据、数据库、日志或集群密钥。
- 提交信息遵循 [docs/commit-convention.md](docs/commit-convention.md)。
- 欢迎提交 Issue 和 Pull Request；请明确说明已验证和仍为占位的功能。

## 9. 开源协议

CephTower 使用 [MIT License](LICENSE)。
