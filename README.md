<div align="center">

# CephTower

</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-Backend-00ADD8?logo=go)](backend/go.mod)
[![Vue](https://img.shields.io/badge/Vue-Frontend-42B883?logo=vue.js)](frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![多语言](https://img.shields.io/badge/多语言-yellow)](README.md)

</div>

<div align="center">

[**简体中文**](README.md) | [繁體中文](docs/readme/README-zh-TW.md) | [English](docs/readme/README-en.md) | [日本語](docs/readme/README-ja.md) | [Français](docs/readme/README-fr.md) | [Deutsch](docs/readme/README-de.md) | [Español](docs/readme/README-es.md) | [Português](docs/readme/README-pt.md) | [Русский](docs/readme/README-ru.md) | [한국어](docs/readme/README-ko.md)

</div>

<div align="center">

CephTower 是一个用于支持管理 Ceph 集群的 Web 控制台项目，后端基于 Go，前端基于 Vue，通过 Ceph Manager Dashboard API 管理整个 Ceph 集群。

</div>

## 1. 功能特性

### 1.1 集群管理

- 通过 Ceph Manager Dashboard API 获取集群摘要和运行状态
- 为后续接入 OSD、Pool、Monitor、MDS、RGW 等管理能力预留清晰边界
- 后端统一封装 Ceph Dashboard 认证、请求、错误处理和兼容性逻辑

### 1.2 Go 后端

- 使用 Go 标准库初始化 HTTP 服务，基础依赖少，便于长期维护
- 提供 `/healthz` 健康检查接口
- 提供 `/api/v1/cluster/summary` 集群摘要接口
- 使用配置文件管理服务监听地址和 Ceph Dashboard 连接信息

### 1.3 Vue 前端

- 使用 Vue 3、Vite、TypeScript 初始化管理控制台
- 首屏为集群概览工作台，不做营销型落地页
- 前端统一通过 CephTower 后端 `/api` 路由访问集群能力

### 1.4 多语言文档

本项目 README 支持以下语言：

| 语言 | 文件 | 状态 |
|------|------|------|
| 简体中文 | [README.md](README.md) | 已支持 |
| 繁體中文 | [docs/readme/README-zh-TW.md](docs/readme/README-zh-TW.md) | 已支持 |
| English | [docs/readme/README-en.md](docs/readme/README-en.md) | 已支持 |
| 日本語 | [docs/readme/README-ja.md](docs/readme/README-ja.md) | 已支持 |
| Français | [docs/readme/README-fr.md](docs/readme/README-fr.md) | 已支持 |
| Deutsch | [docs/readme/README-de.md](docs/readme/README-de.md) | 已支持 |
| Español | [docs/readme/README-es.md](docs/readme/README-es.md) | 已支持 |
| Português | [docs/readme/README-pt.md](docs/readme/README-pt.md) | 已支持 |
| Русский | [docs/readme/README-ru.md](docs/readme/README-ru.md) | 已支持 |
| 한국어 | [docs/readme/README-ko.md](docs/readme/README-ko.md) | 已支持 |

## 2. 项目结构

```text
CephTower/
├── config/                       # 运行配置目录
│   └── config.yaml               # 唯一运行配置
├── backend/                      # Go HTTP API 服务
│   ├── cmd/
│   │   └── server/               # 后端服务入口
│   ├── internal/
│   │   ├── ceph/                 # Ceph Dashboard API 客户端封装
│   │   ├── config/               # 配置文件加载
│   │   └── httpapi/              # HTTP 路由和响应处理
│   └── go.mod                    # Go 模块配置
├── frontend/                     # Vue 3 + Vite Web 控制台
│   ├── src/
│   │   ├── api/                  # 前端 API 调用
│   │   ├── App.vue               # 应用主界面
│   │   ├── main.ts               # 前端入口
│   │   └── styles.css            # 全局样式
│   ├── package.json              # 前端依赖和脚本
│   ├── tsconfig.json             # TypeScript 配置
│   └── vite.config.ts            # Vite 配置
├── docs/                         # 项目文档
│   ├── architecture.md           # 架构说明
│   ├── commit-convention.md      # Agent 提交规范
│   ├── readme/                   # 多语言 README
│   └── references/               # 本地参考资料目录，不纳入 Git 追踪
├── .agents/                      # Codex repo-level skills
├── AGENTS.md                     # Codex / 通用 agent 项目指南
├── CLAUDE.md                     # Claude Code 项目指南
├── Makefile                      # 常用开发命令
├── LICENSE                       # MIT 开源协议
└── README.md                     # 简体中文说明文档
```

## 3. 快速开始

### 3.1 后端服务

调整唯一配置文件 `config/config.yaml`：

```yaml
http_addr: ":36900"
ceph_dashboard:
  base_url: https://ceph.example.com
  username: admin
  password: change-me
  insecure_tls: false
```

启动后端：

```bash
make backend-dev
```

或直接运行：

```bash
cd backend
go run ./cmd/server -config ../config/config.yaml
```

### 3.2 前端控制台

安装依赖并启动开发服务：

```bash
cd frontend
npm install
npm run dev
```

当前前端固定访问同源 `/api`，开发服务默认监听 `36901`，并将 `/api` 代理到 `http://localhost:36900`。生产打包后由 Go 后端提供同一个入口，不再单独维护前端配置文件。

## 4. 开发指南

### 4.1 前置要求

| 工具 | 建议版本 | 用途 |
|------|----------|------|
| Go | 1.26+ | 后端开发和测试 |
| Node.js | 20+ | 前端依赖和构建 |
| npm | 10+ | 前端包管理 |
| Ceph | 具备 Dashboard API 的版本 | 集群管理目标 |

### 4.2 常用命令

| 命令 | 描述 |
|------|------|
| `make backend-dev` | 启动 Go 后端服务 |
| `make backend-test` | 运行后端测试 |
| `make frontend-dev` | 启动 Vue 开发服务 |
| `make frontend-build` | 构建前端产物 |

### 4.3 后端测试

```bash
make backend-test
```

后端测试会将 Go 构建缓存写入 `backend/.cache/go-build`，避免本地沙箱或系统缓存权限影响测试。

## 5. API 说明

### 5.1 健康检查

| 方法 | 路径 | 描述 |
|------|------|------|
| `GET` | `/healthz` | 后端进程健康检查 |

响应示例：

```json
{
  "status": "ok"
}
```

### 5.2 集群摘要

| 方法 | 路径 | 描述 |
|------|------|------|
| `GET` | `/api/v1/cluster/summary` | 获取 Ceph 集群摘要 |

响应示例：

```json
{
  "health_status": "unknown",
  "version": "18.2.0"
}
```

## 6. 技术架构

### 6.1 后端分层

| 模块 | 路径 | 功能描述 |
|------|------|----------|
| 服务入口 | `backend/cmd/server` | 加载配置、初始化客户端、启动 HTTP 服务 |
| 配置模块 | `backend/internal/config` | 读取配置文件并生成运行配置 |
| Ceph 客户端 | `backend/internal/ceph` | 封装 Ceph Dashboard API 调用 |
| HTTP API | `backend/internal/httpapi` | 暴露前端所需的 REST API |

### 6.2 前端分层

| 模块 | 路径 | 功能描述 |
|------|------|----------|
| 应用入口 | `frontend/src/main.ts` | 初始化 Vue 应用 |
| 主界面 | `frontend/src/App.vue` | 集群管理控制台首页 |
| API 调用 | `frontend/src/api` | 与后端 API 通信 |
| 样式 | `frontend/src/styles.css` | 控制台基础视觉样式 |

### 6.3 Ceph Dashboard 集成

- 后端通过 `config/config.yaml` 指定 Dashboard API 地址
- 通过 `ceph_dashboard.username` 和 `ceph_dashboard.password` 完成认证
- 默认开启 TLS 校验，开发环境可通过 `ceph_dashboard.insecure_tls` 临时跳过
- 前端不直接持有 Ceph Dashboard 凭证，只访问 CephTower 后端 API

## 7. 使用场景

### 7.1 集群运维控制台

- 查看 Ceph 集群健康状态
- 汇总 Dashboard API 暴露的核心指标
- 为运维人员提供统一 Web 管理入口

### 7.2 多集群管理基础

- 后续可扩展多集群配置
- 支持将不同 Ceph Dashboard 实例纳入统一视图
- 为集群切换、权限隔离和审计能力预留架构空间

### 7.3 内部平台集成

- 可作为企业内部运维平台的一部分
- 后端 API 可继续接入统一认证、权限、审计和告警系统

## 8. 注意事项

1. `docs/references/` 用于存放参考资料和外部项目，不纳入 Git 追踪。
2. 不建议在前端保存 Ceph Dashboard 用户名、密码或 Token。
3. 生产环境请保持 TLS 校验开启，并使用密钥管理系统保存凭证。
4. 当前项目处于初始化阶段，Ceph 管理接口会按模块逐步扩展。

## 9. 开源协议

MIT License - 详见 [LICENSE](LICENSE) 文件。

## 10. 贡献

欢迎提交 Issues 和 Pull Requests，一起完善 Ceph 集群管理能力。

---

<div align="center">

**如果这个项目对您有帮助，请给个 Star 支持一下！**

</div>
