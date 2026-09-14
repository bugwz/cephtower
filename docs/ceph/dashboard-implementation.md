# Ceph Dashboard 功能迁移与验证记录

本任务目标是以本地参考源码为依据，实现 CephTower 的展示与操作。当前仍在实施，
**不是全部功能已完成的验收报告**。参考目录不作为运行时依赖；不连接或部署测试集群。

## 如何追踪调用链

- 参考界面：`docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/ceph`。
- 参考 HTTP 服务：同目录的 `shared/api`；[完整静态索引](dashboard-source-index.md)。
- 参考接口实现：`dashboard/controllers`，底层 `dashboard/services`。
- Ceph CLI 契约：`src/mon/MonCommands.h`、各 mgr 模块的 CLI 装饰器与实现。
- 当前采集：`backend/internal/integration/ceph/collector*.go`。
- 当前读链路：reconciler → resource store → `handler.ReadResource` → `frontend/src/api/resource.ts`。
- 当前写链路：严格请求契约 → `handler.MutateResource` → `service/mutation` → executor。
- executor 接收参数数组，不拼 shell；现有集群凭证、权限、审计、超时体系继续使用。
- `integration/ceph/command` 有包装函数不等于 API 和前端已接通，必须逐层核对。

## 模块分析与后续核对范围

下表中的命令族用于定位实现，不是可直接执行的完整命令，也不保证所有 Ceph 版本支持。
“已有”仅表示本次开始时能找到相关代码，尚需逐项检查参数、响应与操作覆盖。

| 参考模块 | 数据获取方式 / 命令族 | 当前情况及待补齐范围 |
| --- | --- | --- |
| dashboard / health | `status`、`df detail`、`health detail`、`pg dump`、`osd dump` | 已有总览；本次补 PG 分布、IOPS、健康详情和静默。对象统计、恢复速率、scrub 状态待补 |
| cluster / host | `orch host ls --detail`、`orch ps`、`orch device ls`、`device ls-by-host` | 已有主机详情、SSH、SMART、服务操作；核对硬件和维护/排空全部参数 |
| cluster / monitor | `mon dump`、`quorum_status`、`tell mon.* perf dump` | 已有详情和性能计数；核对历史速率与时钟偏移 |
| cluster / mgr | `mgr dump`、`mgr module ls`、`config get/set` | 已有模块开关；模块配置项编辑与验证待核对 |
| cluster / osd | `osd dump/tree/df`、`osd metadata`、`tell osd.*`、`orch osd rm` | 已有状态、权重、scrub、部署/移除；详情、个体 flags、设备 class、histogram 待补 |
| pool / CRUSH / EC | `osd pool ls detail`、`osd pool get/set`、`osd crush rule`、`osd erasure-code-profile` | 已有管理表单；核对统计、配额、压缩、全部编辑参数与只读状态 |
| cluster / configuration | `config dump/help/set/rm` | 已补集群配置入口、全部选项浏览、按需元数据详情、设置和删除覆盖；修复范围键与字符串值语义 |
| cluster / ceph-user | `auth ls/add/caps/del/import/export` | CephX 实体与 CephTower 登录用户不同；已补独立列表、增删改、keyring 导入及所选实体导出，包含缓存脱敏与 API 链路测试 |
| cluster / logs | `log last` / 外部日志源 | 已接通 log last 直接读取、频道/级别/条数筛选、轮询与下载；移除仅发心跳的假流接口 |
| cluster / upgrade | `orch upgrade check/status/start/pause/resume/stop` | 已有状态和操作，版本选择及 daemon 范围参数待核对 |
| block / rbd | `rbd ls/info/create/resize/rm/feature`、`rbd snap`、`rbd trash`、`rbd namespace` | 已有基础管理；克隆/扁平化、配置、快照保护、任务、回收站完整操作需逐项补齐 |
| block / mirroring | `rbd mirror pool/image`、`rbd mirror snapshot schedule` | 当前为部分池级状态/配置；peer、bootstrap、调度需补齐；镜像启停、promote/demote/resync/snapshot 已接入 |
| block / iSCSI | ceph-iscsi REST API | 需要网关 endpoint，不能把所有操作替换成普通 `ceph` CLI；当前已有外部客户端 |
| block / NVMe-oF | 网关 gRPC | 当前已有 gRPC 客户端；子系统、namespace、listener、host、连接与 QoS 待完整对照 |
| cephfs / filesystem | `fs dump/status/get/set`、`fs volume`、`tell mds.* client ls/evict` | 已有文件系统详情与客户端接口；计数器、rename、auth 与目录操作待核对 |
| cephfs / subvolume | `fs subvolumegroup`、`fs subvolume`、`fs subvolume snapshot` | 已有基本管理；组范围采集、clone 状态、metadata 与完整参数需核对 |
| cephfs / snapshot schedule | `fs snap-schedule` | 已修正创建命令的文件系统参数；列表采集、删除、激活/停用、retention 待补齐 |
| cephfs / directory | libcephfs 或 CephFS 数据面客户端 | 目录配额、快照、浏览不能仅从 MON 获取；现有 cephfs-shell 数据面需核对 |
| nfs | `nfs cluster`、`nfs export` | 已有基础管理；完整 export 属性、CephFS/RGW FSAL 与 ingress 待核对 |
| smb | `smb show/apply/rm` 与模块资源定义 | 已有部分管理；域加入、用户组、资源校验与配置语义需核对 |
| rgw / user / account / role | `radosgw-admin user/account/role` 与 RGW Admin Ops | 当前存在基础页面；配额、subuser、caps、rate limit、角色策略等需逐项扩展 |
| rgw / bucket | RGW Admin Ops、S3 API、`radosgw-admin bucket` | 当前有桶及策略入口；生命周期、版本、锁、加密、tag、notification、replication 待核对 |
| rgw / multisite | `radosgw-admin realm/zonegroup/zone/period` | 已有部分管理；同步策略、placement/storage class、bootstrap/import 待核对 |
| monitoring / prometheus | Prometheus HTTP API | 时间序列/规则由 Prometheus 提供；本次修正规则页错误依赖 Alertmanager 的声明 |
| monitoring / alertmanager | Alertmanager HTTP API | 已有 alerts/silences；需核对规则分组、matcher 和错误展示 |
| grafana | Grafana HTTP API | 已有看板列表；嵌入配置与无 endpoint 行为需核对 |
| auth / role / multi-cluster | CephTower 自有认证、RBAC、集群管理 | 不直接复制 Dashboard 内部 JWT 用户库；根据用户可操作功能对照现有实现 |
| telemetry / settings / feedback | mgr 模块命令及外部服务 | 尚需区分 Ceph 集群功能和仅服务于原 Dashboard 的产品设置 |

## 已实现：总览及健康链路

参考依据：

- `dashboard/controllers/health.py` 的 `basic_health`、`client_perf` 和 `pg_info`。
- `dashboard/services/ceph_service.py` 的 `get_client_perf`、`get_pg_info`。
- `src/mon/MonCommands.h` 中 `health`、`health mute`、`health unmute`。
- `src/mon/HealthMonitor.cc` 和 `health_check.h` 的 checks/mutes JSON 格式。

实现：

1. `ceph health detail --format json` 单独采集健康详情，避免 `status` 的摘要造成详情丢失。
2. 将 `checks` 与 `mutes` 合并；已恢复但仍有 sticky 静默的记录保留，支持用户取消。
3. 详情命令失败将 `health_check` 标记为不可用，交给 reconciler 保留过期缓存。
4. `POST /api/v1/health/mute` 支持可选 `ttl` 和 `sticky`，严格检查时长和布尔类型。
5. `DELETE /api/v1/health/mute` 沿用现有取消操作；成功后前端刷新健康采集。
6. 展示 PG 状态数量/占比、读写 IOPS、读写吞吐、MGR/MDS active/standby。
7. 健康表格展示展开详情、当前是否触发、静默到期和持续静默状态。
8. `alert/rules` 由 Prometheus 提供，前端依赖声明同步修正。

## 已修正：CephFS 快照计划的文件系统参数

参考 `src/pybind/mgr/snap_schedule/module.py` 的 `snap_schedule_add`，位置参数顺序为
`path, snap_schedule, start, fs`。原实现把文件系统名作为第三个参数，实际传入了 start。
现在使用 `fs snap-schedule add <path> <schedule> --fs <filesystem>`，后置检查使用
`fs snap-schedule status <path> --fs <filesystem> --format json`，避免依赖默认文件系统。
测试明确核对写命令和后置检查的参数顺序。

该模块的 `list --recursive` JSON 只包含 schedule/retention 汇总，不能直接当成前端所需的
完整逐路径详情列表。后续采集必须核对 `status` 和路径发现方式，不能简单套用已有列表模型。

## 已实现：CephX 认证实体管理

参考 `dashboard/controllers/ceph_users.py` 的 `user_list`、`user_create`、`user_edit`、
`user_delete`、`export`，以及 `src/mon/MonCommands.h` 的 auth 命令定义。

| API | Ceph 命令及语义 |
| --- | --- |
| `GET /api/v1/ceph/users` | `auth ls --format json` 定期采集后的实体、类型、caps；列表不包含 key |
| `POST /api/v1/ceph/user` | `auth add <entity> <subsystem> <caps> ...`；由 Ceph 生成密钥 |
| `PATCH /api/v1/ceph/user` | `auth caps <entity> <subsystem> <caps> ...`；替换全部权限 |
| `DELETE /api/v1/ceph/user` | `auth rm <entity>`；参考旧名称 auth del 已在源码标为 deprecated |
| `POST /api/v1/ceph/users/import` | `auth import -i -`，keyring 从 stdin 传入 |
| `GET /api/v1/ceph/users/export` | 对每个显式选择的实体执行 `auth export <entity>`，返回 keyring |

- 新增 `ceph_auth` 采集模块及 `ceph_ceph_user` 专用实体表，使用现有数据库迁移机制。
- 表中只保存实体、类型和权限，不保存 Ceph 返回的密钥；导出结果不进入资源缓存。
- 导入请求使用 `keyring` 敏感字段，审计自动脱敏；错误响应不回显导入/导出命令的原始输出。
- 导出限定显式选中的 1–100 个实体，响应设置 `Cache-Control: no-store`。
- 新增集群管理 → CephX 用户入口，支持搜索、类型过滤、分页、权限表单、批量删除与下载导出。
- 编辑会提示权限整体替换；删除确认说明影响；写操作完成后重新采集认证实体。
- caps 按单个参数传递，支持空格和带引号的路径；拒绝选项注入、NUL 和非法子系统。
- 无集群 API 集成测试实际经过路由、请求契约、命令执行模拟、资源入库、列表、导入、导出及审计。

## 已实现：集群配置与运行日志

配置依据：`dashboard/controllers/cluster_configuration.py`、`src/mon/ConfigMonitor.cc`、
`src/common/options.cc` 和 `MonCommands.h`。新增集群配置页面，从已有 `config ls`/`config dump`
缓存浏览选项和已设置的作用域，选中后调用 `GET /api/v1/configuration/option` →
`ceph config help <name> --format json`，展示类型、级别、说明、默认值、范围、枚举、
标签及运行时可更新性。避免每次刷新为数千个选项逐一启动命令。

- 修复配置值只允许“标识符”的错误限制；支持空串、空格、负数及多行 JSON 内容。
- 使用 `--value=<value>` 明确传值；空值/多行值在 `--` 后使用 `--value` 与独立参数。已直接运行参考
  `ceph_argparse.validate` 核对空值、负数、空格和以 `--` 开头的值，避免 CLI 解析歧义。
- 设置、删除、资源版本检查和入库键均保留 `osd/host:node-a`、`osd/class:ssd` 等作用域。
- 后置检查使用 `config dump`，不把带 mask 的作用域误用为单一 daemon 身份查询。
- 重新采集后展示 Ceph 实际返回的值，本地已提交配置不再覆盖发现值。
- 密码类选项的 name/value 结构在审计及资源存储中脱敏；命令参数标为敏感。

日志依据：`dashboard/controllers/logs.py` 的 `load_buffer`、`src/mon/LogMonitor.cc` 和
`src/common/LogEntry.cc`。`GET /api/v1/logs` 接收 `cluster_id`、`channel`、`level`、`limit`，
调用 `ceph log last <limit> <level> <channel> --format json`，返回 `items` 和 `observed_at`。
支持 cluster/audit/cephadm/全部频道、5 个最低级别，限制 1–500 条；最新日志在前。
序列号以字符串输出，避免浏览器丢失大整数精度。

前端支持搜索、暂停/启用 10 秒轮询、取消过期请求和下载当前结果；读取失败不会伪装成
空日志。移除原先只发送心跳、从不提供日志的 `/logs/stream`，以真实轮询替代。
该视图只提供 MON 缓冲区记录，不宣称具备长期日志归档。

新增测试覆盖：命令参数、JSON 错误/空数组、日志顺序和大序列号、配置 metadata、
配置值与 mask 作用域、凭证脱敏，以及路由到命令模拟和入库的完整链路。

## 本轮检查结果

- `make test-backend` 通过（全量 Go 测试及 OpenAPI 一致性检查）。
- `make test-frontend` 通过（TypeScript 类型检查与 Vite 生产构建）。
- `git diff --check` 通过。
- 未提交 Git commit；未修改用户已有的实验环境配置变更。

## 验证边界

无真实集群，fixture 是合成/仓库测试数据，不能当作实机测试记录。
测试覆盖详细健康 JSON、告警恢复后的 sticky 记录、失败保留缓存标志、
命令参数、非法 TTL/类型，以及请求契约。完整检查使用 `make test-backend` 和
`make test-frontend`；OpenAPI 由仓库生成器更新。

仍需实机验证：Ceph 权限与版本差异、静默过期语义、网络故障、规模性能，以及各外部
网关/Prometheus 的实际协议。没有因缺乏集群而在生产路径返回模拟成功。

### MGR module state and schemas

- Reference: `src/pybind/mgr/dashboard/controllers/mgr_modules.py` and
  `src/mon/MgrMonitor.cc` in the Ceph reference tree.
- Combine `ceph mgr module ls --format json` activation state with
  `ceph mgr dump --format json` available-module metadata. Preserve enabled,
  always-on, force-disabled, can-run, load-error and module-option schema fields.
- Exclude the reference dashboard's internal selftest module. Missing command
  output is unavailable data, rather than an empty authoritative module list.
- Module UI now displays load failures and schema details; activation triggers
  collection before rereading the list. The module configuration action opens
  the shared editor filtered to `mgr/<module>/` options, with default scope `mgr`.
  It supports setting and removing overrides and inspecting option metadata.
  `ConfigMonitor.cc` includes manager options in both config ls and config help.
  Writes follow `PyModuleConfig::set_config`: config set/rm with scope mgr.
  Configuration resource identities encode scope and full name together so
  slash-containing option names cannot be confused with scope masks.
- Verified backend tests, OpenAPI consistency and frontend production build.
  Fixtures cover disabled-module objects and force-disabled always-on modules;
  no live Ceph cluster was used.

### OSD on-demand inspection

- Reference `dashboard/controllers/osd.py` get/histogram and `OSD.cc` command handler.
- Added GET `/osd/inspection` with explicit metadata/histogram section selection.
  Executes `ceph osd metadata <id> --format json` or
  `ceph tell osd.<id> perf histogram dump --format json` through the native executor.
- OSD details expose cached state, live metadata and histogram axes/bucket values.
  Histogram rendering now uses a scrollable two-dimensional heat table, with
  inclusive Ceph-provided bucket ranges, axis names, cumulative counts, logarithmic
  color intensity and expandable raw data. Shape mismatches are not plotted.
  Axis ordering and open-ended buckets follow common/perf_histogram.h and .cc.
  Each live section reports errors and supports rereading; no offline success data.
- Exposed the existing deep-scrub operation in the OSD list.
- Backend tests verify command arguments, read-only execution, invalid targets,
  malformed responses and offline failures. Backend/OpenAPI checks and frontend
  build passed; real-cluster and browser visual validation remain outstanding.

- OSD in/out/scrub, reweight and removal now explicitly recollect native resources
  before reloading cached rows. Removal messaging reflects an asynchronous request,
  not completion of physical removal. Frontend build and diff checks passed.

### OSD deployment preview output

- The dry-run command output is returned in action details and rendered in the
  deployment form. Editing inputs clears the preceding preview.
- Preview executes orch apply osd --dry-run, is marked read-only in executor
  metadata, skips the unrelated daemon post-check, and does not persist a
  configured deployment resource. Output uses the existing text redactor.
- A service test verifies exact arguments, one command only, output propagation,
  secret redaction and failure propagation. Backend/OpenAPI checks passed.

- OSD management now shows removal queue status, remaining PGs, host, replacement
  and force flags, and process start time from orch osd rm status. The table exposes
  observation age/staleness and is included in explicit OSD refresh requests.

- OSD, removal queue and MGR module lists now consume all API cursor pages.
  Shared pagination also serves the configuration editor, aggregates stale metadata
  and rejects repeated cursors instead of looping or silently truncating results.

- Snapshot schedule creation supports optional start, subvol and group, matching
  snap_schedule/module.py. Post-check retains subvolume/group scope. Command tests
  verify exact scope, start placement and rejection of control-line input.

- Snapshot schedule status now uses a direct scoped command. Its UI supports
  activate/deactivate/remove of an exact period/start and path-level retention
  add/remove. Retention expressions reject duplicate units and malformed input;
  command tests cover operations and invalid expressions. Full-path discovery
  remains incomplete. The obsolete cached list endpoint and UI have been removed;
  creation now lives alongside direct scoped querying and actions.

- Schedule responses require path, period, start and boolean active state before
  being exposed to actionable UI. Tests reject null rows and ambiguous identities.

### Snapshot schedule discovery boundary

The reference CLI recursive JSON list returns one requested path and flattened
schedule/retention arrays; it does not emit each matching path. Plain output uses
`Schedule.__str__` with unescaped space/newline-separated paths, so arbitrary
filesystem names cannot be recovered reliably. Full-path discovery still needs a
structured data-plane enumeration approach. Scoped status remains authoritative.
Schedule mutations no longer write the obsolete per-filesystem configured cache,
which could collapse multiple paths and schedules into one resource key.

- RBD snapshot UI now exposes clone and rollback via the existing native action
  endpoint. Rollback confirms the target and data-loss effect. Form mutations
  recollect RBD state. Tests verify source/destination namespace preservation,
  exact commands/post-checks and rejection of a missing clone destination.

- RBD group member add/remove and group snapshot creation now connect UI to
  native commands, with qualified-group command tests. Namespace discovery covers
  group membership and snapshots; image and trash scopes are also preserved.
  Group snapshot create/delete/rollback now use explicit group and snapshot targets; broader mirroring remains outstanding.

### RBD pool mirroring collection and configuration

- Mirroring API returns one resource per pool, matching the list UI and mutation identity.
- Pool mode is read with `rbd mirror pool info`; enabled pools additionally use
  `rbd mirror pool status --verbose --format json` for health summaries, daemon records,
  and per-image status, following `src/tools/rbd/action/MirrorPool.cc` output fields.
- The UI exposes disabled/image/pool configuration and confirms disabling mirroring.
  Successful changes trigger native collection before refreshing the table.
- Fixture tests cover two independent pool identities and status fields. No live cluster
  execution has been performed. Peer/bootstrap management and mirror snapshot scheduling remain unfinished.

### RBD image mirroring actions

- Image rows expose journal/snapshot enable, disable, promote, demote, resync, and
  mirror snapshot creation through the existing strict image-action API.
- Commands follow `src/tools/rbd/action/MirrorImage.cc`, with full pool/namespace/image
  targets. Postchecks use mirror image status, or ordinary image info after disable.
- Demote, disable and resync show operation-specific confirmation. The native CLI
  enforces mirror mode and primary-state prerequisites and returns errors to the UI.
- Namespace-aware command tests cover all seven operations. Forced promotion and
  forced disable, peer configuration, and mirror scheduling remain pending.

### RBD mirroring peers

- Added `POST /rbd/mirroring/peer` with strict add/remove actions, per-pool UI forms,
  explicit peer UUID deletion confirmation, and refresh after native collection.
- Commands map to `rbd mirror pool peer add/remove` in `MirrorPool.cc`; add uses
  explicit remote cluster/client and rx-only/rx-tx direction. Postcheck reads pool info.
- Adding currently requires remote connection configuration already present on the
  execution host. Remote credential setup, peer editing and bootstrap remain pending.
- Command tests verify argument boundaries, postchecks, missing targets and invalid
  directions. These are fixture checks, not real cluster validation.

### RBD peer editing

- Peer forms now update site name, client name, remote monitor addresses and direction
  via `rbd mirror pool peer set <pool> <uuid> <field> <value>`.
- The API validates field names and accepts rx-only/tx-only/rx-tx for updates, matching
  Ceph's separate add/set direction rules. Monitor address strings remain one argument.
- Tests cover supported fields, invalid direction, missing UUID and rejected key-file
  parameters. Credential transfer/bootstrap remains a separate unfinished capability.

### Mirroring identity fields and malformed responses

- Pool tables display site_name, mirror_uuid and remote_namespace from native pool info.
- Collector validates the native mode field (disabled/image/pool/init-only); null,
  missing, non-string and unrecognized modes mark mirroring unavailable instead of
  producing empty authoritative state. Tests verify cache-preservation signaling.
- Peer monitor addresses and credentials are not returned by ordinary pool info;
  obtaining those requires a dedicated secret-aware detail path and remains pending.

### RBD group rename and removal

- Group management now exposes rename/remove through `POST /rbd/group/action`.
- Native commands follow `src/tools/rbd/action/Group.cc`. Rename accepts a new group
  name and keeps the source pool and namespace; removal requires an explicit group.
- Postchecks list groups with explicit pool/namespace options. UI confirms removal
  and recollects after successful operations.
- Tests exercise default/named namespaces and reject cross-scope rename, empty path
  components and missing targets. Real cluster behavior remains unverified.

### RBD group snapshot rename

- Group snapshot forms support rename through the existing strict snapshot action API.
- Native command is `rbd group snap rename <pool/namespace/group@old> <new-name>`,
  matching `execute_group_snap_rename` in `Group.cc`. Destination paths are rejected;
  the operation remains within the source group. Postcheck lists that group's snapshots.
- Tests cover namespace-qualified source, plain destination, missing names and invalid
  destination paths. Live cluster verification is still outstanding.

### RBD image rename

- Image actions expose rename. New names stay within the source pool/namespace, as
  required by `src/tools/rbd/action/Rename.cc`; paths and snapshot suffixes are rejected.
- The command uses complete source/destination specs and checks destination image info.
  Successful mutations trigger recollection before rendering the image list.
- Tests cover a namespace-qualified image, destination postcheck and invalid new names.
  No live cluster validation has been performed.

### Snapshot rename UI and group request contract correction

- Ordinary RBD snapshot rows now expose rename through the existing PATCH snapshot API,
  preserving encoded image identity and the original snapshot name.
- Review found the group action request contract had not been registered despite passing
  command tests. Registered it, regenerated OpenAPI, and added explicit registration and
  request validation tests. Prior command-level checks did not establish API completeness.

### Mutation registration verification

- Replaced the self-referential request-contract coverage test with an AST scan of
  actual handler MutateResource registrations. Every registered action must have a
  request contract; missing registrations such as the earlier group-action omission
  now fail the backend test suite. Dynamic action registrations require explicit coverage.
- This verifies handler-to-contract wiring, not execution against a live Ceph cluster.

### RBD trash purge and detail fields

- Trash rows expose confirmed purge of expired images in their pool/namespace.
  Native purge and postcheck use explicit pool/namespace flags; malformed scopes fail.
- Collection uses `trash ls --long` and displays native deleted_at, status, source and
  parent fields, matching `Trash.cc`. Ordinary short listing omits these fields.
- Namespace-aware purge tests pass; real cluster validation remains outstanding.

### RBD trash deferment

- Move-to-trash accepts optional expires_at in the image action API and form.
- RFC3339 timestamps with timezone are normalized to UTC for `trash mv --expires-at`,
  matching `Trash.cc` expiration handling. Omission preserves the native default.
- Tests cover timezone conversion and invalid date inputs. No live cluster verification.

### RBD trash purge cutoff

- Purge accepts an optional RFC3339 expired_before value, normalized to UTC and passed
  to native `trash purge --expired-before`. The confirmation displays the selected cutoff.
- Tests verify namespace preservation, timezone conversion and invalid timestamp rejection.
  Pool usage threshold and scheduled purge remain unfinished.

### RBD image info collection

- Default and named namespace images now fetch `rbd info <spec> --format json`.
  Features, parent identity and detailed fields are exposed in the image table.
- Source fields follow `src/tools/rbd/action/Info.cc`; details retain striping,
  timestamps and other native attributes. Failed info reads mark image collection
  unavailable instead of treating missing details as authoritative.
- Fixture coverage verifies features, parent snapshot and striping detail retention.
  Collection adds one native request per image; large-cluster performance is unverified.

### RBD feature mutation

- Image forms enable/disable exclusive-lock, object-map, fast-diff and journaling;
  deep-flatten is disable-only, matching Dashboard RbdService allowed feature sets.
- Native commands use `rbd feature enable/disable <spec> <feature>` from Feature.cc,
  with image info postcheck and recollection. Ceph enforces feature dependencies.
- Tests cover namespace-qualified commands, unknown features and rejected enable of
  deep-flatten. Multi-feature dependency ordering and live validation remain pending.

### RBD explicit shrink

- Resize forms expose allow_shrink with confirmation describing truncated-data loss.
- Backend adds `--allow-shrink` only for explicit true; sizes retain byte suffixes.
  Native `Resize.cc` rejects shrinking without this flag.
- Tests check both flag paths and namespaced image identity. Filesystem preparation
  remains the operator's responsibility; no live image resize was performed.

### RBD create layout options

- Create forms/API accept data_pool, stripe_unit (bytes), and stripe_count.
- Native create passes --data-pool, --stripe-unit with byte suffix, and --stripe-count,
  following ArgumentTypes.cc. Numeric layout fields must be positive integers.
- Command tests cover full namespace identity, units, and zero-count rejection.
  Ceph validates layout compatibility; no live image creation was performed.

### RBD object size

- Image creation supports object_size, sent to native --object-size in bytes.
- UI offers powers of two from 4 KiB to 32 MiB; backend validates the same range and
  shape. ArgumentTypes.cc documents the native range. Boundary tests cover valid
  endpoints, undersize, oversize and non-power-of-two values.
- No live cluster creation has been performed.

### Native RBD long-list correction

- List.cc emits image (not name) for long-list records and mixes image and snapshot
  entries in the same array. Corrected the wire field and skip records with snapshot.
- Updated fixture data to match native output, including an image plus its snapshot.
  Namespace tests now verify that snapshot rows do not duplicate image/snapshot collection.
- Earlier fixtures did not represent the actual native long-list schema; their passing
  results were insufficient evidence for live compatibility. Backend checks now use the
  corrected shape. Real cluster validation is still outstanding.

### RBD snapshot purge

- Image rows expose confirmed purge of unprotected snapshots via native `snap purge`.
  The API action preserves full image scope and postchecks with snap ls.
- Snapshot protected/timestamp fields were verified against Snap.cc native list output.
- Command tests cover namespaced image purge. Live cluster validation remains pending.

### RBD runtime status

- Image collection reads `rbd status <spec> --format json` and displays watchers,
  migration and persistent_cache data through runtime_status, matching Status.cc.
- Empty responses mark collection unavailable. Fixture tests retain all three sections.
- This adds another request per image; large-cluster collection cost and live behavior
  remain unverified. Status reflects the last collection, not a continuous live stream.

### RBD observed-state ownership

- RBD mutations no longer synthesize resource configurations from request bodies.
  Native collection owns displayed resource state; audit/task records retain operations.
- RBD DTOs ignore configured overlays so request names/targets cannot override native
  identities or leak destination fields into observed resources.
- Regression tests verify DTO ownership and that RBD mutations do not write synthetic
  state. UI already recollects after successful operations. Live validation is pending.

### RBD mutation scope correction

- PATCH image rename preserves namespace; image deletion postcheck lists the same
  pool and namespace instead of default namespace only.
- Encoded image specs require pool/image or pool/namespace/image, without empty
  components, extra levels or snapshot suffixes. Tests cover malformed specs and scope.
- Backend checks pass; real cluster mutation validation remains outstanding.

### RBD malformed list protection

- Null image, snapshot, trash and group lists mark their resource kind unavailable;
  valid empty arrays remain authoritative empty results. Missing snapshot names and
  trash IDs also mark the collection incomplete instead of silently omitting records.
- Backend regression checks exercise null list handling. Live failure-mode validation
  remains pending.

### RBD group identity details

- Groups collect native `group info` and expose group_id in the table. Missing IDs
  mark group collection unavailable.
- Group fixtures now match Group.cc: member state is numeric and snapshot list uses
  snapshot for the name field. Native member namespace is retained in raw details.
- Backend/frontend checks pass; live group inspection remains unverified.

### RBD namespace native shape correction

- Namespace.cc returns an array of objects with name, not an array of strings.
  Corrected collection and namespace fixtures; invalid names mark collection unavailable.
- Namespace rows now carry an explicit namespace field. Removed redundant/empty table
  columns. Namespaced image/snapshot/trash/group tests use the corrected native shape.
- Earlier namespace fixtures were inaccurate; real cluster validation remains pending.

### RBD effective image configuration

- Image collection reads `config image list <spec> --format json`; configuration rows
  preserve name/value/source from Config.cc and are displayed in image details.
- Fixture coverage verifies image-level QoS value/source retention. Null replies mark
  collection unavailable. Per-image configuration editing remains unfinished.
- Additional per-image collection cost and real-cluster behavior remain unverified.

### RBD image configuration editing

- Image forms set/remove non-secret RBD configuration overrides via native config image
  set/remove. Postcheck lists effective configuration, followed by recollection.
- Names require rbd_ prefix; set values are nonempty single-line strings. Ceph validates
  supported options and values. Tests cover namespace scope and missing set values.
- Secret configuration and empty-value overrides are not supported by this form.
  Real-cluster validation remains pending.

### Resource action menus

- Extended row operations are grouped in a click-open menu; detail/edit/delete remain
  directly accessible. Column width accounts for visible controls. Capability blocking
  applies to the menu and each operation.
- Added mirroring to the manual-refresh resource mapping. RBD mutation-triggered
  recollection already covered mirroring.
- Frontend typecheck/build passes; browser layout and keyboard interaction are unverified.

### Nested resource details

- Complex list cells open a field-focused drawer instead of flattening objects into
  long table strings. Object arrays render as paginated field tables in RecordDetail;
  nested objects use formatted text with bounded scrolling.
- Removed the 12-field detail cutoff so collected attributes are all accessible.
- Frontend build/typecheck passes; browser layout and interaction are unverified.

### Operation-specific form inputs

- Resource forms support conditional fields. RBD copy/rename targets, trash expiry,
  clone destinations, configuration values and group rename fields appear only for
  their relevant operation. Required inputs validate before submission.
- Unmounted fields do not preserve values, preventing stale input across action changes.
- Frontend typecheck/build passes; browser interaction remains unverified.

### Create snapshots from image rows

- Image action menus create snapshots using the selected encoded image identity, so
  pool and namespace need not be typed again. Existing native snapshot API is reused.
- Corrected the snapshot path in the manual-refresh mapping.
- Frontend typecheck/build passes; browser interaction remains unverified.

### Ceph time fields in details

- Detail views recognize deletion/snapshot/create/access/modify timestamps and last_update,
  including fields in nested tables.
- Explicitly zoned ISO timestamps use browser local formatting. Native ctime-style
  values without timezone remain verbatim rather than receiving an assumed timezone.
- Frontend typecheck/build passes; browser timezone scenarios remain unverified.

### RBD watcher identifier precision

- Native Status.cc emits unsigned client/cookie identifiers. Collector converts these
  fields to decimal strings before persistence/API serialization to avoid browser rounding.
- Tests retain 9007199254740993 and uint64 maximum 18446744073709551615 exactly.
  This verifies fixture serialization behavior, not a live watcher connection.

### RBD destination path validation

- Creation, copy, deep-copy and clone now validate complete pool/image or
  pool/namespace/image paths with the same rules as encoded source identities.
- Regression tests reject bare names, empty components, extra levels and snapshot
  suffixes for copy/clone destinations. Real cluster operations remain unverified.

### RGW user suspension correction

- Replaced nonexistent user modify --suspended with native user suspend/enable.
  Combined edits run modify then the state subcommand and user info postcheck.
- UI recognizes native numeric suspended state and permits max_buckets=-1 (unlimited).
- Tests cover suspend/enable alone and combined with edits. Source evidence is
  radosgw-admin.cc command registration and RGW user JSON encoding. No live RGW test.

### RGW editable email clearing

- User edits send empty email explicitly; backend uses --email= to retain empty argument
  semantics. Missing email remains unchanged for API callers. Multiline values fail.
- Native ceph_argparse confirms binary system flags accept true/false values and named
  arguments support equals syntax. Tests cover clearing and invalid email lines.
- Backend/frontend checks pass; live RGW validation is pending.

### RGW user creation requirements

- Creation requires display_name, matching RGW user validation, and supports optional
  max_buckets including -1 for unlimited. Invalid limits fail before native execution.
- Tests cover missing display name and bucket limits. Backend/frontend checks pass;
  actual RGW creation remains unverified.

### RGW observed user state

- User mutations trigger native user recollection; request bodies no longer synthesize
  user resource state or override observed fields. UID is the full native list identifier.
- Tests verify native suspended/email values take precedence and no synthetic create
  state is written. Backend/frontend checks pass; live RGW validation remains pending.

### RGW user storage statistics

- User collection reads native user stats without sync/reset mutation flags. Display
  retains stats and last_stats_sync/update. Account users are labeled account scope,
  following radosgw-admin.cc owner selection.
- Fixture tests cover account scope and full UID. Real RGW accounting freshness and
  collection performance remain unverified.

### RGW incomplete detail handling

- User/account/role detail failures and empty responses no longer emit placeholder
  resources. Their failure mappings mark collection unavailable and preserve old state.
- Null list replies are also unavailable. A regression test verifies null user details
  produce no placeholder and no nil-map write.
- Backend checks pass; live failure-mode validation remains pending.

### RGW account detail fields

- Maps native account id/name to account_id/account_name used by the UI.
- Account tables expose quota, bucket_quota and user/role/group/bucket/access-key limits,
  following RGWAccountInfo JSON encoding in rgw_common.cc.
- Fixture tests cover identity mapping and quota retention; live account validation pending.

### RGW 角色原生数据格式与创建约束

- 对照 `src/rgw/radosgw-admin/radosgw-admin.cc` 的角色列表实现，直接采集完整对象数组，避免按字符串名称列表解析时丢失全部角色。
- 页面显示原生 `RoleName`、`Path`、`Arn`、`AssumeRolePolicyDocument`、`MaxSessionDuration`、`CreateDate` 字段。
- 创建角色的信任策略改为必填 JSON 对象，API 契约与表单同步；Ceph 负责最终 IAM 策略语义校验。
- 增加原生对象数组采集回归测试。当前角色操作仍只有创建，删除、策略管理、租户和账户范围尚待实现；未进行真实集群验证。

### RGW 角色删除操作

- 新增 `DELETE /api/v1/rgw/role`，使用必填 `name` 构造 `radosgw-admin role delete --role-name <name>`，后续执行 `role list` 检查；对应原生 `OPT::ROLE_DELETE` 分支。
- 角色列表新增删除确认操作，创建和删除后主动采集 `rgw_role`。角色资源不再写入或合并请求参数作为展示状态。
- 命令参数及缺失名称回归测试、`make test-backend`（含 OpenAPI 一致性检查）、`make test-frontend` 均通过。
- 本节取代上节“角色操作只有创建”的状态描述。角色策略管理、租户/账户范围仍未完成；未进行真实集群验证。

### RGW 角色编辑

- 新增 `PATCH /api/v1/rgw/role` 和角色列表编辑表单，维护信任策略 JSON 与最大会话时长（3600–43200 秒，匹配 `rgw_role.h`）。
- 命令链：`role-trust-policy modify --role-name ... --assume-role-policy-doc ...` → `role update --role-name ... --max-session-duration ...` → `role get`。参考原生 `ROLE_TRUST_POLICY_MODIFY`、`ROLE_UPDATE` 分支及 Dashboard `RGWRoleEndpoints.role_update`。
- 两次写入不具备事务性，第二步失败可能留下第一步的修改。命令链及输入验证测试、后端测试、OpenAPI 检查、前端构建通过，未进行真实集群验证。
- 尚待角色内联/托管权限策略管理及租户、账户范围实现。

### RGW 角色内联权限策略

- 新增 `POST /api/v1/rgw/role/policy`，按 `action=put/delete` 调用 `role-policy put/delete --role-name ... --policy-name ...`；写入通过 `--perm-policy-doc` 提交 JSON 对象，随后 `role get` 检查。
- 角色行操作提供新增/替换与删除策略表单及确认，操作后重新采集；列表展示原生 `PermissionPolicies`（PolicyName/PolicyValue）和 `ManagedPermissionPolicies`。
- 对照 `radosgw-admin.cc` 的 `ROLE_POLICY_PUT/DELETE` 与 `rgw_role.cc` 序列化字段实现。命令参数及非法 JSON 回归测试、后端测试/OpenAPI 检查、前端构建均通过。
- 尚未进行实机验证；托管策略挂载/解除及租户、账户范围尚待实现。

### 角色字段独立更新与托管策略前置条件

- 角色更新 API 允许单独提交信任策略或会话时长，拒绝空更新，并在最后一次写入后读取角色。页面仅在信任策略文本改变时提交该字段，避免单改时长时覆盖策略。
- 独立字段更新与最终读取回归测试、后端测试/OpenAPI 检查、前端构建通过。
- 核对原生 `ROLE_POLICY_ATTACH` 分支发现其要求账户角色；当前角色采集只覆盖默认租户，因此本轮未接入托管策略挂载按钮。账户范围采集和操作标识仍是后续必要工作，不能视为已完成。

### 账户角色范围

- 采集账户详情后执行 `role list --account-id <id>`，角色采用 `AccountId/RoleName` 标识，默认租户角色仍使用角色名；拒绝采集结果中账户 ID 与请求不一致的角色。
- 角色创建表单增加账户 ID，列表展示账户；创建、编辑、删除、内联权限策略 API 与原生命令均传递账户范围，后续读取也使用同一账户。
- 账户列表或详情获取失败会同时将角色采集标记为不完整，避免把未采集的账户角色当作已删除。
- 账户身份隔离和写入/读取范围回归测试、前后端检查通过。未进行实机验证；非默认租户角色枚举及托管策略挂载仍待实现。

### 角色创建参数与托管策略源码核对

- 创建角色支持 `description`、`max_session_duration`，分别传给 `--description`、`--max-session-duration`；会话时长验证为 3600–43200 整数秒，未提交时使用 Ceph 默认值。页面新增输入并展示原生 Description。
- 参数和边界回归测试、后端测试/OpenAPI 检查、前端构建均通过，未实机验证。
- 托管策略 attach/detach 分支调用 `get_role(role_name, tenant, account_id)` 后 `load_by_id`；`RGWRole` 名称构造函数不设置 `info.id`，而 RadosRole::load_by_id 直接按 info.id 读取。参考源码存在定位不一致，需要进一步验证或寻找可用服务接口，本轮未将此命令接入可操作按钮。

### RGW 账户删除

- 对照 Dashboard `rgw_iam.py` 账户 DELETE 操作及原生 `ACCOUNT_RM`，新增 `DELETE /api/v1/rgw/account` → `account rm --account-id` → `account list`。
- 页面提供删除确认，创建/删除后重新采集账户和角色；账户展示仅采用采集结果，不再叠加请求参数。账户非空限制交由 `rgw::account::remove` 检查，没有引入级联清理。
- 命令及必填参数测试、后端测试/OpenAPI 检查、前端构建通过，尚未实机验证。账户修改、限额和配额操作仍需继续补齐。

### RGW 账户编辑与数量上限

- 对照 Dashboard `rgw_iam.py::set` 和原生 `rgw::account::modify`，新增 `PATCH /api/v1/rgw/account`，执行 `account modify` 后 `account get`。
- 页面支持账户名称、邮箱、max_users/max_roles/max_groups/max_buckets/max_access_keys；数量限制验证为 -1 或非负 32 位整数。
- 原生 modify 不允许更换租户，且忽略空名称和邮箱，故不提供租户编辑并拒绝清空名称/邮箱，避免虚报更新成功。
- 参数、零值、负值与小数验证测试，后端测试/OpenAPI 检查、前端构建均通过；未实机验证。账户容量/对象配额操作仍待实现。

### RGW 账户统计展示

- 对照 `rgw_account.cc::stats`，采集 `account stats --account-id <id> --format json`，保留 stats、last_synced、last_updated 于 storage_stats。
- 账户列表增加容量与对象统计详情入口；未触发 sync/reset，命令失败或缺失 stats 标记账户采集不完整。
- 账户原生数据回归测试已包含统计返回，后端测试/OpenAPI 检查、前端构建及 diff 检查通过。未进行真实集群验证；配额设置仍待实现。

### 账户与默认 Bucket 配额设置

- 新增 `PUT /api/v1/rgw/account/quota` 和两种范围的配额表单，使用 `quota enable/disable --account-id --quota-scope account/bucket --max-size --max-objects`，随后读取账户；原生命令同次写入限额和启用状态。
- 容量以字节提交并由 Ceph 向上取整至 KiB；API -1 表示无限制，容量命令用 -1024 避免参考源码中 unsigned rgw_rounded_kb(-1) 溢出为零，最终原生 clamp 得到 -1。对象上限直接使用 -1。
- 回归测试覆盖范围、单位、非法负值、小数及超出安全整数的值。后端测试/OpenAPI 检查与前端构建通过；尚未实机验证。

### RGW 用户配额

- 新增 `PUT /api/v1/rgw/user/quota`，支持 user/bucket 范围、启停及容量/对象上限，用户列表增加两类配额操作与字段展示。
- 启用用 `quota enable --uid` 同时保存限额；停用因原生 set_quota_info 的 disable 分支忽略限额，先 `quota set` 再 `quota disable`，最后 `user info`，两次写入不具备事务性。
- 保留 tenant$user 标识并验证输入；测试覆盖租户 UID、无限容量、零对象限额及停用命令链。前端构建通过，未实机验证。

### RGW 用户管理权限

- 对照 Dashboard 用户 capability 表单与原生 caps add/rm，新增 `POST /api/v1/rgw/user/caps` 及用户行操作，支持单个类型的 read/write/read,write/* 权限添加、移除。
- 命令使用 tenant$user 原始 UID 和独立 caps 参数，随后 user info 并刷新采集；拒绝在类型中拼接多个权限表达式，Ceph 最终校验支持的类型。
- 命令及类型验证回归测试、后端测试/OpenAPI 检查通过。未进行真实集群验证。

### RGW 用户限流读取

- 对照 `show_user_ratelimit` 与 `RGWRateLimitInfo::dump`，采集 `ratelimit get --uid <uid> --ratelimit-scope user --format json`，解析 user_ratelimit 对象并展示为 rate_limit。
- 保留 enabled、max_read_ops/max_write_ops/max_read_bytes/max_write_bytes；命令失败或返回对象缺失标记用户采集不完整。
- 原生嵌套结构回归测试、后端测试/OpenAPI 检查、前端构建通过，未实机验证。限流修改和全局限流仍待实现。

### RGW 用户限流修改

- 新增 `PUT /api/v1/rgw/user/ratelimit` 及用户操作表单，提交 enabled 和四项读写请求/字节限额。页面注明每 RGW 每分钟计量，0 表示无限制。
- 对照 set_ratelimit_info，先 ratelimit set 保存限额，再 enable/disable，最后 get；保留租户 UID，拒绝负数、非整数及超出安全整数范围的值。两次写入不具备事务性。
- 启停命令链、参数回归测试、后端测试/OpenAPI 检查、前端构建均通过，未实机验证。全局和 Bucket 限流仍待补齐。

### Bucket 限流读取

- 对照 show_bucket_ratelimit，采集 `ratelimit get --bucket ... --ratelimit-scope bucket`，有 tenant 时显式传入；解析 bucket_ratelimit 并在 Bucket 列表展示。
- Bucket 详情失败或空对象不再生成占位数据，详情与限流失败均标记 Bucket 采集不完整。
- 原生嵌套返回回归测试、后端测试/OpenAPI 检查、前端构建通过，未实机验证。Bucket 限流修改及全局限流仍待实现。

### Bucket 限流修改

- 新增 `PUT /api/v1/rgw/bucket/ratelimit`，从编码 Bucket ID 解析租户与名称，使用 ratelimit set → enable/disable → get，三步传递相同的 bucket/tenant/scope。
- Bucket 行操作支持四项限额和启停，完成后重新采集；原生命令使用 rgw_admin 能力，不要求 S3 端点，不将请求体持久化为 Bucket 展示状态。
- 租户命令链回归测试、后端测试/OpenAPI 检查、前端构建均通过，未实机验证；两次写入不具备事务性，全局限流仍待实现。

### Bucket 全局列表租户标识修正

- 原生全局 bucket list 遍历 bucket metadata 键，键为 `[tenant/]bucket`。采集现在拆分名称和租户，为 bucket stats 与 ratelimit get 分别传递 --bucket/--tenant。
- 校验统计返回的 bucket、tenant 与请求一致，再构造编码资源标识，避免重复编码租户前缀或误用其他租户数据。
- 租户 Bucket fixture 改为原生列表格式，并校验展示名与编码标识；后端测试（含 OpenAPI 检查）通过。本次未改前端，未进行真实集群验证。

### Bucket 统计字段展示

- 对照原生 bucket_stats 输出，列表展示 tenant、versioning、num_shards、placement_rule、zonegroup、bucket_quota、object_lock_enabled、mfa_enabled、reshard_status、id、creation_time、mtime。
- 原生 Bucket ID 使用 id 字段，与操作使用的编码 natural_key 分开；版本控制编辑读取已采集的 suspended 状态。
- 前端构建通过；字段来自现有 bucket stats 采集，无后端变更，未实机验证。

### 单个 Bucket 配额设置

- 新增 `PUT /api/v1/rgw/bucket/quota` 与行操作表单，编码标识解析为 bucket/tenant，调用 quota enable 或 set→disable，随后 bucket stats 并刷新采集。
- 对照 set_quota_info，容量为字节并向上取整到 KiB，负值 -1 表示无限制；停用命令忽略限额，因此先单独保存，两次写入不具备事务性。
- 租户范围及启停命令链测试、后端测试/OpenAPI 检查、前端构建通过。未进行真实集群验证。

### 多步操作最终读取修复

- 共用执行器现在使用命令链最后一个已声明的后置检查（含其二进制），不再忽略 followup.check；空 RBD/RGW 检查不再被转换成单独的 --format json。
- 实际 Service.Execute 回归测试验证限流依次执行 set、enable、get，且读取标记为非写入；后端测试/OpenAPI 检查通过。
- 此修复保证声明的读取被执行，不代表已比较读取值和请求值。尤其参考源码 Bucket quota 外层忽略 set_bucket_quota 返回值，仍需进一步增加结果字段一致性校验；未实机验证。

### Bucket 配额读取值校验

- Bucket 配额操作现在比较最终 bucket stats 的 tenant、bucket、bucket_quota.enabled/max_size/max_objects 与请求，容量按 KiB 向上取整。
- 写入返回成功但结果未变化、读取字段缺失、错误租户或 JSON 无效均返回 post_check_failed，补上参考原生配额调用可能忽略写入错误的缺口。
- Service.Execute 模拟执行器回归测试及后端测试/OpenAPI 检查通过。未实机验证；该结果一致性校验目前针对单个 Bucket 配额。

### 默认 Realm 全局限流读取

- RGW 总览新增 `global ratelimit get --format json` 采集，展示默认 Realm 的 user_ratelimit、bucket_ratelimit、anonymous_ratelimit。任一范围对象缺失标记状态采集不完整。
- 总览移除不对应 RGWStatus 的用户字段，显示 realms 与 global_rate_limit；限额单位为每 RGW 每分钟。
- 三类原生返回结构测试、后端测试/OpenAPI 检查及前端构建通过，未实机验证。其他 Realm 的限流选择及全局写入操作仍待实现。

### Realm 原生详情

- Realm 名称列表后逐项执行 `realm get --rgw-realm <name>`，保留 id、name、current_period、epoch；页面增加 Epoch 展示。
- 返回身份不匹配或读取失败标记 Realm 采集不完整，不生成仅含名称的占位详情。多站点列表返回 null 也标记不完整。
- 原生详情及错误身份回归测试、后端测试/OpenAPI 检查和前端构建通过，未实机验证。Zonegroup、Zone 完整详情仍待补齐。

### Zonegroup 原生详情

依据 `src/rgw/radosgw-admin/radosgw-admin.cc` 的 ZONEGROUP_GET 和
`src/rgw/rgw_zone.cc` 的 RGWZoneGroup::dump，枚举后调用
`radosgw-admin zonegroup get --rgw-zonegroup <name> --format json`。
详情通过现有资源 API 提供，页面展示 Realm、主 Zone、成员、端点、放置目标、
主机名、同步策略和启用特性。名称不匹配、缺少 ID 或空返回均视为采集失败，
不生成只有名称的详情记录，并保留该资源类型的采集不完整状态。
新增离线测试覆盖嵌套详情及异常返回；尚无真实集群验证，Zonegroup 编辑操作待补齐。
