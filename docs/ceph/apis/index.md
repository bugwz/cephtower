# Ceph 20.2.2 Mgr Dashboard API

本文档集整理自 Ceph v20.2.2 源码内置 Dashboard OpenAPI 描述，用于本项目后续通过 mgr dashboard API 操作 Ceph 集群。

## 来源与调用约定

- OpenAPI 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`
- 版本来源：`docs/references/ceph/CMakeLists.txt` 中的 `VERSION 20.2.2`
- 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- API 基础路径：OpenAPI `basePath` 为 `/`，接口路径以 `/api/...` 为主。
- 认证方式：带 `security: [{jwt: []}]` 的接口使用 Bearer JWT。通常先调用 `POST /api/auth` 获取 `token`，后续请求使用 `Authorization: Bearer <token>`。
- 公开接口：未声明 `security` 的接口通常为公开入口或由控制器单独处理，调用前仍应结合部署侧 Dashboard 配置确认。
- 内容类型：请求体通常为 `application/json`；响应内容类型通常为 `application/vnd.ceph.api.v1.0+json`，部分接口使用其他 API 版本 MIME。
- 异步任务：许多写操作可能返回 `202`，表示操作仍在执行，需要查询任务队列。
- 通用错误：多数接口包含 `400`、`401`、`403`、`500`。具体响应体以运行时 Dashboard 返回为准。
- 版本兼容性总览：[compatibility.md](compatibility.md)

## 分类索引

| 分类 | 路径域 | 文档 | 接口数 |
| --- | --- | --- | --- |
| 认证 | `auth` | [endpoints/auth.md](endpoints/auth.md) | 3 |
| 块存储 RBD | `block` | [endpoints/block.md](endpoints/block.md) | 48 |
| CephFS | `cephfs` | [endpoints/cephfs.md](endpoints/cephfs.md) | 45 |
| 集群 | `cluster` | [endpoints/cluster.md](endpoints/cluster.md) | 13 |
| 集群配置 | `cluster_conf` | [endpoints/cluster_conf.md](endpoints/cluster_conf.md) | 6 |
| CRUSH 规则 | `crush_rule` | [endpoints/crush_rule.md](endpoints/crush_rule.md) | 4 |
| Daemon | `daemon` | [endpoints/daemon.md](endpoints/daemon.md) | 2 |
| 纠删码配置 | `erasure_code_profile` | [endpoints/erasure_code_profile.md](endpoints/erasure_code_profile.md) | 4 |
| 功能开关 | `feature_toggles` | [endpoints/feature_toggles.md](endpoints/feature_toggles.md) | 1 |
| 反馈 | `feedback` | [endpoints/feedback.md](endpoints/feedback.md) | 5 |
| Grafana | `grafana` | [endpoints/grafana.md](endpoints/grafana.md) | 3 |
| 硬件 | `hardware` | [endpoints/hardware.md](endpoints/hardware.md) | 1 |
| 健康状态 | `health` | [endpoints/health.md](endpoints/health.md) | 6 |
| 主机 | `host` | [endpoints/host.md](endpoints/host.md) | 10 |
| iSCSI | `iscsi` | [endpoints/iscsi.md](endpoints/iscsi.md) | 7 |
| 日志 | `logs` | [endpoints/logs.md](endpoints/logs.md) | 1 |
| Mgr 模块 | `mgr` | [endpoints/mgr.md](endpoints/mgr.md) | 6 |
| Monitor | `monitor` | [endpoints/monitor.md](endpoints/monitor.md) | 1 |
| MOTD | `motd` | [endpoints/motd.md](endpoints/motd.md) | 2 |
| 多集群 | `multi-cluster` | [endpoints/multi-cluster.md](endpoints/multi-cluster.md) | 9 |
| NFS Ganesha | `nfs-ganesha` | [endpoints/nfs-ganesha.md](endpoints/nfs-ganesha.md) | 6 |
| NVMe-oF | `nvmeof` | [endpoints/nvmeof.md](endpoints/nvmeof.md) | 40 |
| OSD | `osd` | [endpoints/osd.md](endpoints/osd.md) | 20 |
| 性能计数器 | `perf_counters` | [endpoints/perf_counters.md](endpoints/perf_counters.md) | 8 |
| 存储池 | `pool` | [endpoints/pool.md](endpoints/pool.md) | 6 |
| Prometheus | `prometheus` | [endpoints/prometheus.md](endpoints/prometheus.md) | 9 |
| RGW | `rgw` | [endpoints/rgw.md](endpoints/rgw.md) | 92 |
| 角色 | `role` | [endpoints/role.md](endpoints/role.md) | 6 |
| 服务 | `service` | [endpoints/service.md](endpoints/service.md) | 7 |
| 设置 | `settings` | [endpoints/settings.md](endpoints/settings.md) | 5 |
| SMB | `smb` | [endpoints/smb.md](endpoints/smb.md) | 16 |
| 概览 | `summary` | [endpoints/summary.md](endpoints/summary.md) | 1 |
| 任务 | `task` | [endpoints/task.md](endpoints/task.md) | 1 |
| Telemetry | `telemetry` | [endpoints/telemetry.md](endpoints/telemetry.md) | 2 |
| Dashboard 用户 | `user` | [endpoints/user.md](endpoints/user.md) | 7 |

## 统计

- 路径数：281
- 接口操作数：403
- 分类数：35

## 各版本接口数量

| 版本 | 路径数 | 接口操作数 |
| --- | --- | --- |
| v16.2.15 | 134 | 195 |
| v17.2.9 | 147 | 214 |
| v18.2.8 | 195 | 278 |
| v19.2.5 | 220 | 314 |
| v20.2.2 | 281 | 403 |
