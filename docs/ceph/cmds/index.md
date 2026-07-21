# Ceph 命令参考

> 来源：`docs/references/ceph` 的 Git tags：`v16.2.15`、`v17.2.9`、`v18.2.8`、`v19.2.5`、`v20.2.2`。
> 本文档由 `tools/generate_ceph_command_docs.py` 生成。

本文档整理 Ceph monitor/mgr command table 与 mgr Python 模块声明的 `ceph ...` 命令，
用于后续在 `backend/internal/integrations/ceph` 中新增直接执行 Ceph CLI 的能力。

## 返回约定

- 命令可通过 `ceph <prefix> --format json` 或 `--format json-pretty` 请求 JSON 输出时，仅记录 JSON 输出形态。
- 每条命令的返回信息包含成功/失败通道和 JSON 支持情况。
- 返回字段只在该命令支持的所有版本源码中逐项校验通过后记录，并在字段表中标注源码确认位置。
- Ceph 源码的命令表声明参数、权限、模块与帮助文本；没有集中声明完整返回 schema，未校验字段不会写入文档。
- 写操作通常返回确认文本或空输出；失败时返回非 0 退出码，并在 stdout/stderr 中携带错误说明。
- `admin-socket/` 目录记录 `ceph daemon <daemon>.<id> <cmd>` 这类 per-daemon 命令；远程场景通常也可通过 `ceph tell <daemon>.<id> <cmd>` 调用。
- `tell/` 目录把同一批 per-daemon 命令展开为 `ceph tell <daemon>.<id> <cmd>` 的调用形式。

## 目录

- [admin-socket/bluestore](admin-socket/bluestore.md)
- [admin-socket/client](admin-socket/client.md)
- [admin-socket/common](admin-socket/common.md)
- [admin-socket/mds](admin-socket/mds.md)
- [admin-socket/mgr](admin-socket/mgr.md)
- [admin-socket/mon](admin-socket/mon.md)
- [admin-socket/osd](admin-socket/osd.md)
- [cluster/core](cluster/core.md)
- [components/cephfs](components/cephfs.md)
- [components/mds](components/mds.md)
- [components/mgr](components/mgr.md)
- [components/mon](components/mon.md)
- [components/osd](components/osd.md)
- [components/pg](components/pg.md)
- [mgr-modules/alerts](mgr-modules/alerts.md)
- [mgr-modules/balancer](mgr-modules/balancer.md)
- [mgr-modules/cephadm](mgr-modules/cephadm.md)
- [mgr-modules/crash](mgr-modules/crash.md)
- [mgr-modules/dashboard](mgr-modules/dashboard.md)
- [mgr-modules/device](mgr-modules/device.md)
- [mgr-modules/feedback](mgr-modules/feedback.md)
- [mgr-modules/hello](mgr-modules/hello.md)
- [mgr-modules/influx](mgr-modules/influx.md)
- [mgr-modules/insights](mgr-modules/insights.md)
- [mgr-modules/iostat](mgr-modules/iostat.md)
- [mgr-modules/k8sevents](mgr-modules/k8sevents.md)
- [mgr-modules/nfs](mgr-modules/nfs.md)
- [mgr-modules/nvme-gw](mgr-modules/nvme-gw.md)
- [mgr-modules/nvmeof](mgr-modules/nvmeof.md)
- [mgr-modules/orchestrator](mgr-modules/orchestrator.md)
- [mgr-modules/progress](mgr-modules/progress.md)
- [mgr-modules/prometheus](mgr-modules/prometheus.md)
- [mgr-modules/rbd](mgr-modules/rbd.md)
- [mgr-modules/restful](mgr-modules/restful.md)
- [mgr-modules/rgw](mgr-modules/rgw.md)
- [mgr-modules/scrub](mgr-modules/scrub.md)
- [mgr-modules/service](mgr-modules/service.md)
- [mgr-modules/smb](mgr-modules/smb.md)
- [mgr-modules/telegraf](mgr-modules/telegraf.md)
- [mgr-modules/telemetry](mgr-modules/telemetry.md)
- [mgr-modules/tell](mgr-modules/tell.md)
- [mgr-modules/test_orchestrator](mgr-modules/test_orchestrator.md)
- [mgr-modules/zabbix](mgr-modules/zabbix.md)
- [tell/bluestore](tell/bluestore.md)
- [tell/client](tell/client.md)
- [tell/common](tell/common.md)
- [tell/mds](tell/mds.md)
- [tell/mgr](tell/mgr.md)
- [tell/mon](tell/mon.md)
- [tell/osd](tell/osd.md)
