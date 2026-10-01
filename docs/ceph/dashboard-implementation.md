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
| dashboard / health | `status`、`df detail`、`health detail`、`pg dump`、`osd df`、`osd dump` | 已有总览、PG 分布、对象统计、IOPS、恢复速率、scrub 状态、健康详情和静默 |
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
| block / mirroring | `rbd mirror pool/image`、`rbd mirror snapshot schedule` | 已有池状态、模式、peer、bootstrap token、镜像操作与镜像级调度；池/集群级调度管理待补齐 |
| block / iSCSI | ceph-iscsi REST API | 需要网关 endpoint，不能把所有操作替换成普通 `ceph` CLI；当前已有外部客户端 |
| block / NVMe-oF | 网关 gRPC | 当前已有 gRPC 客户端；子系统、namespace、listener、host、连接与 QoS 待完整对照 |
| cephfs / filesystem | `fs dump/status/get/set`、`fs volume`、`tell mds.* client ls/evict/perf dump` | 已有文件系统详情、客户端、授权、卷重命名、MDS 计数器会话趋势、Rank/备用 MDS 及池容量；其余完整字段待核对 |
| cephfs / subvolume | `fs subvolumegroup`、`fs subvolume`、`fs subvolume snapshot`、`fs clone` | 已有组范围采集、clone 状态/进度/失败展示、快照克隆、进行中任务取消和快照可见性；metadata 与其余完整参数需核对 |
| cephfs / snapshot schedule | `fs snap-schedule` | 已有全路径发现、精确状态、创建、删除、激活/停用、retention 和模块启用 |
| cephfs / directory | libcephfs 或 CephFS 数据面客户端 | 已有实时目录浏览、目录元数据、双维度配额、目录增删/重命名/移动及目录快照列表/创建/删除 |
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

### 对象健康、恢复和 scrub 指标

- `ceph status --format json` 提供池数量、对象数、恢复吞吐和 PG 状态；保留缺失字段为
  未知，不把旧版本或暂不可用的数据显示成零。
- `ceph pg dump summary --format json` 的 `pg_map.pg_stats_sum.stat_sum` 提供对象副本、
  degraded、misplaced 和 unfound 计数，和 Dashboard `get_pg_info()` 使用同一组字段。
- `ceph osd df --format json` 的逐 OSD `pgs` 计算平均 PG/OSD，忽略 CRUSH 根和主机节点。
- `ceph osd dump --format json` 提供 `noscrub`、`nodeep-scrub` 标志；结合 PG 状态，按
  Dashboard 语义输出 disabled、active 或 inactive。命令失败时输出未知，不伪造状态。
- 总览页新增对象健康副本比例、异常副本明细、恢复吞吐、scrub 状态、池数量和
  PG/OSD 密度。对象健康计算与参考前端一致：副本总数扣除三类异常副本。
- 附加命令是可选采集；失败不会覆盖已有总览，也不会让核心 `status`/`df` 数据失效。

## 已修正：CephFS 快照计划的文件系统参数

参考 `src/pybind/mgr/snap_schedule/module.py` 的 `snap_schedule_add`，位置参数顺序为
`path, snap_schedule, start, fs`。原实现把文件系统名作为第三个参数，实际传入了 start。
现在使用 `fs snap-schedule add <path> <schedule> --fs <filesystem>`，后置检查使用
`fs snap-schedule status <path> --fs <filesystem> --format json`，避免依赖默认文件系统。
测试明确核对写命令和后置检查的参数顺序。

该模块的 `list --recursive` JSON 只包含 schedule/retention 汇总，不能直接当成前端所需的
完整逐路径详情列表。实现与 Dashboard 控制器保持一致：先执行
`fs snap-schedule list / --recursive=true --fs=<filesystem>`，从逐行短格式发现并去重路径，
再对每个路径执行 `fs snap-schedule status <path> --fs=<filesystem> --format=json`。这样缓存的
每条记录保留 path、schedule、start、active、retention、创建/清理计数及子卷作用域。

无计划时 plain list 会返回 `ENOENT`；采集器使用同命令的 JSON 模式确认 `{}` 后才把结果视为
权威空列表，其他命令失败或非法结构会保留上次有效数据。新增
`GET /api/v1/filesystem/snapshot/schedules` 返回缓存列表；前端同时提供全部计划表和原有的指定
路径实时查询，全部计划可直接按其文件系统、路径、周期和开始时间执行启用、停用或删除。

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

## 已实现：CephFS 访问授权发现

CephFS Dashboard 的授权写入最终调用 `fs authorize`，但 Ceph 没有单独的授权列表命令。
CephTower 复用 `ceph auth ls --format json` 的 MDS caps，在不增加 CLI 调用的前提下把每条
`allow` 授权拆成 `cephfs_authorization` 资源。解析遵循 `MDSAuthCaps.cc` 的 cap 语法，支持：

- 逗号或分号分隔的多条授权，以及包含逗号的 `gids`；
- `fsname`、带引号或无引号的 `path`、`uid`、`gids`、`network` 和 `root_squash`；
- `r`、`rw`、`p`、`s`、`f` 扩展及 `*`/`all` 通配权限；
- 没有 `fsname` 的全局 MDS cap，使用 `*` 明确展示其作用域；
- 原始 MDS cap 与结构化字段同时展示，且认证密钥不会进入资源缓存。

授权创建成功后，`cephfs_authorization` 归属 `ceph_auth` 采集模块，立即重新执行同一条
`auth ls` 命令并更新授权列表；定时采集也会把已撤销或被修改的授权标记为过期。

## 已实现：CephFS 客户端会话

参考 Dashboard `CephFSClients` 与客户端表格，拓扑采集对每个文件系统执行
`tell mds.<fsname>:0 session ls --format json`。采集兼容 Ceph 版本中顶层数组及
`sessions` 对象两种 JSON 外形，保留完整 session 指标，并把 `client_metadata` 规范化为
类型、版本、主机和挂载根路径。列表展示与 Dashboard 对齐的 ID、类型、状态、版本、主机和
根路径，详情仍可查看 lease、cap、请求和存活时间等原始字段。

客户端驱逐限定到所属文件系统的 rank 0：
`tell mds.<fsname>:0 client evict id=<client_id>`，不再向 `mds.*` 广播可能与其他文件系统
同号客户端冲突的命令。后置检查复用同一 rank 的 `session ls`，成功后立即刷新
`cephfs_client` 缓存；没有活动 MDS 或返回结构无效时保留最后一次有效会话列表。

## 已修正：CephFS Shell 与目录配额

`cephfs-shell` 的入口参数只支持 `--config`、`--fs`、`--batch` 和 `--test`，不能复用
`ceph`/`rbd` 的 `--conf --name --keyring` 参数前缀。executor 现在通过 `CEPH_CONF` 和
`CEPH_ARGS` 向 libcephfs 传递临时配置、客户端实体及 keyring，并校验客户端实体名称，避免
把认证参数当成 shell 子命令参数或注入环境参数。

目录配额写入明确使用所选文件系统：
`cephfs-shell --fs <filesystem> setxattr <path> ceph.quota.max_bytes <bytes>`，随后在同一
文件系统执行 `getxattr` 并核对返回值。路径支持空格和引号，传给 shell 前按单一 token
转义；拒绝 NUL、换行和逗号，其中逗号在 cephfs-shell 中是多命令分隔符。这样写操作不再
依赖默认文件系统，也不会把“命令退出成功但值未生效”当成成功。

## 已实现：CephFS 目录浏览与双维度配额

参考 Dashboard `CephFS.ls_dir()`、`get_directory()` 和 `get_quotas()`，目录页改为按需读取，
不再查询一个从未被采集器填充的 `cephfs_entry` 缓存。`GET /api/v1/filesystem/entries`
接收文件系统和绝对路径，执行：

1. `cephfs-shell --fs <filesystem> ls -la <path>`，解析直属目录的权限、大小、UID、GID、
   修改时间和名称；文件不会混入目录树。
2. 对当前非根目录及每个直属子目录执行
   `cephfs-shell --fs <filesystem> quota get <path>`，读取 `max_bytes` 与 `max_files`。
3. cephfs-shell 用退出码 9 表示扩展属性未设置；仅该退出码会按 Dashboard 语义转换为
   0（不限制），其他退出、超时和不可解析输出均返回错误，不伪装成空目录。

executor 为非交互调用写入临时 `cephfs-shell.conf` 并关闭颜色，确保 `ls -l` 输出不会夹带
ANSI 转义。配额读取最多 8 路并发，且单层目录最多接受 500 个子目录，以限制一次请求触发
的数据面调用数量；路径中的
逗号、换行和 NUL 被拒绝，含空格或引号的路径作为一个 shell token 转义。

目录页支持选择文件系统、输入路径、进入子目录、返回上级，并实时显示两种配额。配额写入
支持同时或分别设置 `max_bytes`、`max_files`，0 表示不限制；双字段写入分成两个明确的
`setxattr` 步骤，最后在同一个指定文件系统回读两个属性并逐值核对。能力探测改用
cephfs-shell 实际支持的 `--help`，不再调用不存在的 `--version`。

## 已实现：CephFS Rank 与备用 MDS

对照 Dashboard `fs_status()`、`_find_standby_replays()` 与 Rank 表，新增实时
`GET /api/v1/filesystem/mds`：`ceph fs dump --format json` 按 `mdsmap.fs_name` 选择 FS，
从 `in`、`up`、`info` 关联活跃 Rank，未分配的 in Rank 标为 failed，standby-replay
独立标为 `<rank>-s`；保留 laggy 标记和字符串 GID。全局 standbys 单独展示，明确不是
该文件系统独占资源。核验重复 Rank、GID/名称、info 键与 GID 一致性及 up→info 关联，
不把畸形 FS Map 当作空结果。FS 不存在返回 404。

随后 `ceph mds metadata --format json` 读取实例 Ceph 版本；元数据失败不隐藏有效拓扑，
单独提示，缺失版本显示 —。性能采样新增 `mds_mem.dn/dir/cap`、
`mds_sessions.session_count` 和 `mds_log.replay`，Rank 表按名称及 GID 匹配已有样本，
展示 Dentries/Inodes/Dirs/Caps、采样时间及活动速率。active 使用客户端请求速率，
standby-replay 使用日志回放速率，其他状态与缺失样本不伪造活动。客户端数优先 Rank 0，
缺失或为零时参考其他 Rank，但不求和；保持字符串精度。拓扑和性能分别采样、展示时间，
不宣称为跨命令原子快照。离线测试覆盖 failed/replay/laggy、备用实例、版本局部失败、
拓扑身份异常、路由及前端 GID/速率/客户端计数逻辑；尚未实机验证。

## 已实现：CephFS 存储池容量

对照 `CephFS.fs_status()` 的 pools 表与 `CephfsDetailComponent` 容量展示，详情页新增
元数据池/数据池实时容量表。`GET /api/v1/filesystem/pools` 接收 `cluster_id`、`fs`，
执行 `ceph fs get <fs> --format json` 读取 `metadata_pool` 与 `data_pools`，再执行
`ceph df detail --format json`，按 pool ID 精确匹配，只返回该文件系统关联的池。
池名来自原生 df 条目，匹配缺失的池仍保留 ID 并标记错误，不把其他池的值用于替代。
逻辑存储量使用 `stats.stored`，可用量使用 `stats.max_avail`，估算容量为两者之和；
`stats.bytes_used` 另列为物理占用，不混入参考 Dashboard 的逻辑使用率口径。
容量是当前集群空间、放置约束下的估算，不能视为预留空间；共享池的值也不是此 FS 独占量。

API 以十进制字符串输出容量，后端 big.Int 加总和前端 BigInt 百分比计算避免大整数失真。
页面显示二进制容量单位、精确字节 tooltip、用途和逻辑使用率，支持独立实时读取及详情页
刷新联动。缺失/非法统计显示 null 或池级错误，容量为零时不计算无意义的百分比。
离线测试覆盖多个数据池、不同于逻辑量的物理占用、超 uint64 总和、零值、缺失统计、
重复 ID、原生命令失败、路由调用与前端格式/使用率边界；尚未经过真实集群验证。

## 已实现：CephFS MDS 性能计数器

参考 Dashboard `CephFS._mds_counters()` 的 11 项非标签计数器与 `CephfsChartComponent`
的 inode/request 趋势，新增实时 `GET /api/v1/filesystem/performance`。
先执行 `ceph fs get <fs> --format json`，核验 `mdsmap.fs_name` 并从 `info` 定位该文件系统
的全部 MDS（包含 standby-replay，不引入无归属的全局 standby）；再逐实例执行
`ceph tell mds.<name> perf dump --format json`。`ceph_context.cc` 注册的 `perf dump`
返回非标签指标，与参考 mgr `get_unlabeled_counter` 的指标来源对应。此读取链不依赖
volumes 模块或外部 Prometheus。MDS 名称/GID 校验、排序和最多 256 实例限制防止不受控目标。
每个实例返回身份、rank、state、采样时间和指标：
`mds_server.handle_client_request`、`mds_log.ev`、`mds_cache.num_strays`、
`mds.exported`、`mds.exported_inodes`、`mds.imported`、`mds.imported_inodes`、
`mds.inodes`、`mds.caps`、`mds.subtrees`、`mds_mem.ino`。
uint64 值与 GID 以字符串提供，缺失指标为 null；错误单独挂在实例上，不伪造 0 或吞掉错误。

详情页每轮完成后间隔 10 秒采样，手动采样可用，最多保留 60 次当前页面会话样本。
展示全部当前值与 inode 数/请求速率图；速率由累计差值除以真实采样秒数得到，使用 BigInt
先求差再转换，实例 GID 变化、负差、缺失值、读取失败与无效时间都断开趋势。
原生命令不提供 ceph-mgr 已保存历史，因此不将页面会话趋势冒充 mgr 历史；超过安全整数
范围的 gauge 保留精确表值，不绘制失真的图点。空关联 MDS 与局部失败均有明确提示。
离线测试覆盖命令作用域、大整数、部分失败、缺失项、畸形返回、实时路由和前端速率边界；
尚未进行真实集群验证。

## 已实现：CephFS 文件系统卷重命名

对照 Dashboard `CephFSUi.rename()`、CephFS 编辑表单和 volumes 模块的
`rename_fs_volume()`，文件系统列表新增独立“重命名卷”操作。
`PUT /api/v1/filesystem` 要求 `cluster_id`、`fs`、`new_name`、`confirmed: true`，
执行 `ceph fs volume rename <old> <new> --yes-i-really-mean-it`。任务风险级别为 high，
界面确认提醒维护窗口、客户端 CephX 重新授权、可能的存储池与 MDS 服务变更。
只调用完整的 volumes 操作，不使用仅改变 FS Map 名称的 `fs rename` 替代。
原生模块负责调整 MDS 服务和存储池，多个数据池时可能仅返回未重命名全部数据池的提示；
保留脱敏后的原生命令输出与警告到任务详情，不将名称核验冒充全部底层对象的验证。
后置 `fs volume ls --format json` 必须同时验证新名称存在、旧名称消失；空、畸形或重复
对象不能证明成功。成功后刷新 filesystem 和 pool 缓存；MDS 服务需重新观测。
拒绝无确认、相同名称和不合法名称；离线测试覆盖原生命令及回读失败，尚无真实集群验证。

## 已实现：CephFS 目录生命周期与快照

### 目录创建与删除

对照 Dashboard `mk_dirs()` 与 `rm_dir()`，目录页可在当前位置创建相对路径，或删除直属
空目录。`POST /api/v1/filesystem/entry` 调用
`cephfs-shell --fs <fs> mkdir -p -m 0755 <absolute-path>`，允许同时创建缺失父目录。
`DELETE /api/v1/filesystem/entry` 调用 `cephfs-shell --fs <fs> rmdir <absolute-path>`，
由 Ceph 检查目录是否为空。两项写操作都回读父目录的 `ls -la`，逐名称验证存在或消失，
并复用异步任务和审计。根目录、快照目录、通配符、逗号和换行路径会被拒绝，避免误操作。
测试覆盖含空格的目录、非法作用域、无法解析的后置列表以及路由到任务执行的完整调用链。

### 目录重命名与移动

目录行新增目标绝对路径表单，`PATCH /api/v1/filesystem/entry` 接收 `fs`、`path`、
`destination`。对照 Dashboard `rename_path()` 与原生 shell `do_mv()`，执行
`cephfs-shell --fs <fs> mv <source> <destination>`，最终调用 LibCephFS `rename`。
执行前读取源父目录，确认源是目录；执行后分别读取目标父目录和源父目录，验证目标目录
存在、源目录消失。未能读取或解析列表会让任务失败，而不是将命令接受当作验证成功。
拒绝根目录、快照路径、源和目标相同、目标位于源子树，以及通配符/逗号/换行。
保留 Ceph 原生替换语义：可能替换空目标目录，非空目标由 Ceph 拒绝，界面明确提示。
这不是事务性的目录身份验证，其他客户端并发修改仍可能造成竞态；尚未进行真实集群验证。
离线测试覆盖跨父目录含空格命令、作用域校验、路由任务及前置/双后置调用链。

### 目录快照

目录浏览页新增当前目录的快照表、创建表单和删除确认。参考 Dashboard
`CephFS.ls_snapshots()`、`mk_snapshot()`、`rm_snapshot()`，数据面调用链为：

- `GET /api/v1/filesystem/entry/snapshots` → `cephfs-shell --fs <fs> ls -la <path>/.snap`。
- `POST /api/v1/filesystem/entry/snapshot` → `cephfs-shell --fs <fs> snap create <name> <path>`。
- `DELETE /api/v1/filesystem/entry/snapshot` → `cephfs-shell --fs <fs> snap delete <name> <path>`。

创建和删除通过同目录快照列表核验目标名称存在或消失；不可解析的列表不能证明删除成功。
列表忽略普通文件、`.`、`..` 和以下划线开头的内部快照，与参考 Dashboard 的过滤一致。
展示 `ls -l` 提供的修改时间，不将其冒充 libcephfs 的 `st_ctime` 创建时间。

`client_snapdir` 是客户端虚拟目录名称。executor 给临时 cephfs-shell 客户端设置
`--client_snapdir=.snap`，使列表与 `snap` 内部 `conf_get('client_snapdir')` 的路径一致，
不会修改集群配置。写请求明确限定集群、文件系统、目录和名称，复用任务执行与审计链路。
无集群测试覆盖目录快照发现、命令参数、名称校验、后置核验和 API 路由。

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

- The schedule page reads the collected `snap_schedule` manager-module state before
  enabling query or mutation controls. A disabled runnable module can be enabled through
  the existing strict manager-module API; unavailable modules surface `can_run` errors
  instead of failing the first schedule command.

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
  execution has been performed. Pool- and cluster-level schedule administration remains unfinished.

### RBD image mirroring actions

- Image collection reuses `rbd info <pool[/namespace]/image> --format json` and
  exposes its `mirroring.mode`, `state`, `global_id`, and `primary` fields. The
  image table now shows the synchronization mode, state, primary/secondary role,
  and the next scheduled mirror snapshot without adding another per-image command.
- Image rows expose journal/snapshot enable, disable, promote, demote, resync, and
  mirror snapshot creation through the existing strict image-action API.
- Commands follow `src/tools/rbd/action/MirrorImage.cc`, with full pool/namespace/image
  targets. Postchecks use mirror image status, or ordinary image info after disable.
- Demote, disable and resync show operation-specific confirmation. The native CLI
  enforces mirror mode and primary-state prerequisites and returns errors to the UI.
- Promotion optionally maps to `rbd mirror image promote <spec> --force`; the UI
  requires an explicit switch and warns about dual-primary and data-conflict risk.
  Disable similarly maps to `rbd mirror image disable <spec> --force`, matching
  `MirrorImage.cc` support for disabling a non-primary image. Both paths require an
  explicit switch and operation-specific high-risk confirmation.
- Strict request validation rejects non-boolean force values and force on other actions.
- Namespace-aware command tests cover all seven operations, forced promotion, and
  forced disable. Fixture tests cover enabled and absent mirroring metadata.
  Live-cluster validation remains pending.

### RBD mirror snapshot schedules

- RBD image rows now show the effective snapshot schedule, its image/namespace/pool/
  cluster inheritance source, all interval/start-time pairs, and the next scheduled run.
- Collection executes one recursive `rbd mirror snapshot schedule list --format json`
  and one global `status --format json` per storage refresh, then applies the most
  specific schedule to every image without per-image command fan-out.
- Image actions add a schedule, remove one interval/start pair, or remove all schedules
  defined directly on that image. They map to native `schedule add/remove --image=...`
  commands and verify the exact image-level state through a recursive list readback.
- Interval and start-time inputs are strictly validated. Fixture tests cover exact
  command boundaries, normalized timezone-aware start times, readback failures,
  inheritance precedence, next-run merging, malformed native data, and namespaces.
  Pool- and cluster-level schedule administration and live-cluster validation remain.

### RBD mirroring peers

- Added `POST /rbd/mirroring/peer` with strict add/remove actions, per-pool UI forms,
  explicit peer UUID deletion confirmation, and refresh after native collection.
- Commands map to `rbd mirror pool peer add/remove` in `MirrorPool.cc`; add uses
  explicit remote cluster/client and rx-only/rx-tx direction. Postcheck reads pool info.
- Direct peer add still requires remote connection configuration already present on the
  execution host. Peer editing and bootstrap token exchange are available separately.
- Command tests verify argument boundaries, postchecks, missing targets and invalid
  directions. These are fixture checks, not real cluster validation.

### RBD peer editing

- Peer forms now update site name, client name, remote monitor addresses and direction
  via `rbd mirror pool peer set <pool> <uuid> <field> <value>`.
- The API validates field names and accepts rx-only/tx-only/rx-tx for updates, matching
  Ceph's separate add/set direction rules. Monitor address strings remain one argument.
- Tests cover supported fields, invalid direction, missing UUID and rejected key-file
  parameters. Bootstrap token exchange is implemented as a separate secret-aware flow.

### RBD mirroring bootstrap tokens

- Pool rows expose token creation and import forms following Ceph Dashboard's bootstrap
  workflow. Disabled pools are enabled in image mode before either operation.
- Creation maps to `rbd mirror pool peer bootstrap create <pool> --site-name=<site>`;
  import maps to `rbd mirror pool peer bootstrap import <pool> -` with explicit
  `--site-name=<site>` and `--direction=<rx-only|rx-tx>` options. The token is passed
  only through stdin.
- The two APIs execute synchronously instead of entering the durable operation queue so
  bootstrap secrets are never persisted in operation parameters. Requests and responses
  use `Cache-Control: no-store`; audit request redaction removes the token value.
- Token shape, pool/site names, and direction are validated before execution. Fixture
  tests cover automatic pool enablement, exact arguments, stdin-only secret handling,
  postchecks, invalid input, malformed native output, and secret-free errors. No live
  cluster execution has been performed.

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
  Features, parent identity, size, object count/size, format, creation time, data
  pool, block prefix, order, and stripe layout are exposed as typed table fields.
- Source fields follow `src/tools/rbd/action/Info.cc`; details retain striping,
  timestamps and other native attributes. Failed info reads mark image collection
  unavailable instead of treating missing details as authoritative.
- Images with `fast-diff` additionally run `rbd du <spec> --format json`. The head
  row supplies current-image usage, while summing snapshot and head rows supplies
  total usage including snapshots, matching the two values shown by Dashboard.
  Images without `fast-diff` omit usage rather than triggering an expensive scan.
- Fixture coverage verifies the exact usage command, snapshot aggregation, 64-bit
  capacity fields, features, parent snapshot, and layout retention. Collection adds
  one usage request only for eligible images; large-cluster performance is unverified.

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
- Default and named namespace images now share the scoped `ls --long --pool` collector.
  This removes the earlier duplicate default-namespace list/enrichment path and lets one
  image usage result feed both image and snapshot observations.
- Updated fixture data to match native output, including an image plus its snapshot.
  Namespace tests verify exactly one list and one eligible usage command per scope.
- Earlier fixtures did not represent the actual native long-list schema; their passing
  results were insufficient evidence for live compatibility. Backend checks now use the
  corrected shape. Real cluster validation is still outstanding.

### RBD snapshot purge

- Image rows expose confirmed purge of unprotected snapshots via native `snap purge`.
  The API action preserves full image scope and postchecks with snap ls.
- Snapshot protected/timestamp fields come from `rbd snap ls <spec> --format json`,
  following `Snap.cc`; string protection values are normalized to booleans for API and
  UI action state.
- For fast-diff images, the already collected `rbd du` rows are joined by snapshot name
  and exposed as `used_bytes`/`disk_usage`, without another command per snapshot. The
  action form defaults to protect or unprotect according to the observed state.
- Non-snapshot-mirrored images query `rbd children <image@snapshot> --format json`,
  matching `Children.cc`; child pool, namespace, and image identities are displayed.
  Snapshot-mirrored images skip this fan-out, matching Dashboard behavior.
- Delete is disabled with an explanatory tooltip while the snapshot is protected or
  still has child images. The submit path repeats the guard before issuing any request.
- Snapshot actions now include independent copy alongside clone. Copy maps to
  `rbd copy <pool[/namespace]/image@snapshot> <destination>` from `Copy.cc`; clone
  continues to map to `rbd clone`. Both validate full destination scope and verify the
  created image with `rbd info`.
- Namespace fixtures cover protected state, timestamp, usage, children, identity, and
  command deduplication. Command tests cover namespaced copy, clone, and purge. Live
  validation remains pending.

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

### Zone 原生详情

参考 Dashboard 的 `rgw-multisite-zone-form.component.ts::getZonePlacementData`，
结合原生 ZONE_GET 和 RGWZoneParams::dump，枚举后逐项执行
`radosgw-admin zone get --rgw-zone <name> --format json`。
通过现有资源 API 提供 Realm、放置池、存储类别及内部池配置，页面增加对应列；
完整详情保留其它原生字段。Zone 端点属于 Zonegroup 的成员 Zone，已在该页展示。
名称或 ID 异常时标记 Zone 采集不完整。系统密钥由现有持久化前递归脱敏处理；
离线回归覆盖详情、异常身份、空数据以及凭据脱敏后保留放置池。
Zone 编辑与真实集群验证仍待完成。

### Realm 编辑与默认 Realm

对照 RgwMultisite::edit_realm 接入 PATCH /rgw/realm：名称变更执行
`realm rename --rgw-realm <name> --realm-new-name <new_name>`，
选中默认时随后执行 `realm default --rgw-realm <new_name>`，最后按新名称读取详情。
无名称变化且未选择默认会拒绝空操作。多个写命令不具有事务性。
页面提交后重新采集 Realm 和 RGW 状态，默认标记来自 realm list 的 default_info，
不再用请求体覆盖 Realm 的真实采集数据。离线测试覆盖命令顺序、目标名称、
布尔校验和默认标记；未执行真实集群验证。

### 创建默认 Realm

Realm 创建表单增加“设为默认 Realm”，POST /rgw/realm 接受可选布尔字段 default。
依据参考 RgwMultisite::create_realm，选中时在 realm create 命令添加 --default；
未选中不传此开关，保留 Ceph 原生默认选择行为。拒绝非布尔参数，
执行后沿用 Realm/RGW 状态刷新。离线测试覆盖省略、false、true 和非法参数。

### Zonegroup 创建参数

依据 RgwMultisite::create_zonegroup，创建表单/API 支持 Realm、逗号分隔端点、
master 和 default；后端分别传入 --rgw-realm、--endpoints、--master、--default。
省略选项时使用 Ceph 原生行为，布尔值 false 不传对应开关。创建后重新采集
Zonegroup，停止以请求数据覆盖原生详情。离线测试覆盖完整参数和非法类型；
真实集群验证、Zonegroup 编辑仍待完成。

### Zonegroup 基础编辑

对照 RgwMultisite::edit_zonegroup/modify_zonegroup，新增 PATCH /rgw/zonegroup
及编辑表单，支持重命名、替换非空端点、设为主/默认 Zonegroup。写入后执行
period update --commit，并显式使用 Realm ID 限定目标，最后按新名称读取详情。
无 Realm 的 Zonegroup 编辑、清空端点、成员和放置目标编辑仍待补齐。
多个命令不具有事务性；测试覆盖顺序、Period 范围、读回目标及无效输入，未实机验证。

### 无 Realm 的 Zonegroup 编辑及 Realm 参数修正

原生 ZONEGROUP_MODIFY 将 --realm-id 用于重新绑定 Realm，并非查询范围限制。
编辑命令移除此参数，Realm ID 仅用于 Period 提交；没有 Realm 时跳过 Period，
直接读取更新后的 Zonegroup。增加独立 Zonegroup 命令回归，页面允许空 Realm ID。

### Zonegroup 成员管理

参考 add_or_remove_zone，在 Zonegroup 编辑 API 和表单增加 add_zones/remove_zones。
依次执行 zonegroup add/remove --rgw-zonegroup <新名称> --rgw-zone <成员>，
成员变更后再提交所属 Realm 的 Period，并读回详情。拒绝重复、交叉和非法成员名。
离线测试覆盖重命名后的成员目标、命令顺序及参数异常；未实机验证。

### 多站点默认状态展示

原生 REALM_LIST、ZONEGROUP_LIST、ZONE_LIST 均输出 default_info ID。采集器按
详情 ID 比较后通过 API 提供 is_default，Zonegroup 与 Zone 页面展示“当前上下文默认”。
命令未指定 Realm 上下文，因此此标记代表该次命令解析出的默认对象，
不声称展示每个 Realm 各自的默认对象。缺失或非法 default_info 保持未知，
不伪装成 false；空字符串明确表示没有匹配对象。离线测试覆盖三类资源的
匹配、不匹配、空、缺失及非法字段。

### Zone 创建拓扑参数

依据 RgwMultisite::create_zone，Zone 创建表单/API 支持 Zonegroup、端点及
master/default 开关，映射为 --rgw-zonegroup、--endpoints、--master、--default。
创建后刷新 Zone 和 Zonegroup 原生数据，停止持久化请求覆盖层。离线测试覆盖
完整命令和非法参数。系统凭据、同步配置、分层参数与真实集群验证仍待补齐。

### Zone 创建同步来源

对照 create_zone 和原生参数解析，创建 API/表单增加 sync_from_all 与 sync_from。
布尔选项显式生成 --sync-from-all=true/false，指定来源生成 --sync-from <列表>。
省略时使用 Ceph 默认行为，false 不会被当作省略。离线测试覆盖两个布尔值、
逗号分隔来源及非法类型；未实机验证。

### 创建归档 Zone

参考 rgw-multisite-zone-form 的 archive_zone 开关，Zone 创建表单新增“归档 Zone”，
选中时提交 tier_type=archive，原生命令传入 --tier-type archive。未选中不传参数，
沿用普通 Zone 的原生行为。API 限定目前表单支持的 archive 值；离线测试覆盖
命令及非法类型/取值，未实机验证。

### Zone 系统用户凭据

Zone 创建增加密码输入框，成对接受 access_key/secret_key，映射参考 create_zone
的 --access-key/--secret 参数。参数标记为敏感，API 字段为 writeOnly，沿用
审计及原生详情递归脱敏；不保存请求覆盖层。测试覆盖敏感参数位置和不完整凭据。
未实机验证。

### Zone 的拓扑和同步详情

RGWZone::dump 中的端点、归档类型、同步来源、只读标记及支持特性属于
Zonegroup 成员记录，而非 ZoneParams。采集后按原生 Zone ID 关联这些记录，
通过 zonegroup_memberships 提供给 Zone 页面详情表，保留每个 Zonegroup 的名称、
ID、Realm 和主 Zone 标记，避免同名误关联或多组信息互相覆盖。
离线测试覆盖名称变化、同名不同 ID、多组成员及 false 值保留；未实机验证。

### Zone 重命名

新增 PATCH /rgw/zone 与重命名表单，执行 zone rename --rgw-zone <旧名称>
--zone-new-name <新名称>，可指定 --rgw-zonegroup 更新成员名称；原生代码未指定时
尝试默认组，读取组失败或没有该成员仍可能返回成功。页面在唯一成员组时预填，
随后刷新 Zone/Zonegroup。提供 Realm ID 时提交该 Realm Period，最后按新名称读回。
离线测试覆盖组与 Period 参数、无 Realm 和空操作；其它 Zone 编辑仍待补齐。

### Zone 端点编辑

对照 modify_zone，PATCH /rgw/zone 和编辑表单支持替换非空端点列表。
指定 Zonegroup 后执行 zone modify --rgw-zone <新名称> --endpoints <列表>
--rgw-zonegroup <组>；仅修改端点不触发重命名，组合操作先重命名再修改，
最后提交提供的 Realm Period 并读回。要求显式选择组避免更新默认组中的错误成员。
测试覆盖两种执行路径和缺少组名；清空端点及其它 Zone 编辑参数仍待补齐。

### Zone 同步来源编辑

对照 modify_zone，编辑 API/表单支持 sync_from_all 和 sync_from，明确区分
保持原值、开启、关闭。显式开启全部同步且提供来源时，按参考实现使用
--sync-from-rm；其余使用 --sync-from。编辑要求指定 Zonegroup，沿用 Period
提交和刷新流程。测试覆盖布尔参数及两种来源命令；未实机验证。

### Zone 主/默认选项编辑

对照 modify_zone，Zone 编辑 API 和表单支持设为主 Zone、设为默认 Zone，
分别生成 --master、--default。false 表示不修改该选项，不表示取消主/默认身份。
要求指定 Zonegroup，支持单独更新和重命名后更新，沿用 Period 提交。
离线测试覆盖两种执行路径及布尔校验；未实机验证。

### Zone 归档类型编辑

参考 modify_zone 和同步模块注册，空 tier_type 映射默认 rgw 模块，archive
映射归档模块。编辑表单区分保持原值、普通、归档；普通显式生成 --tier-type=，
归档生成 --tier-type=archive，避免把空值误认为省略。测试覆盖类型切换和
非法类型，沿用指定 Zonegroup 与 Period 流程；未实机验证。

### Zone 只读编辑

原生 ZONE_MODIFY 的 is_read_only_set 区分省略与显式 false。编辑 API/表单
增加 read_only，生成 --read-only=true/false，支持保持原值、开启、关闭。
只读状态仍从 Zonegroup 成员详情读取；沿用指定组和 Period 提交流程。
离线测试覆盖开启、关闭及非法类型，未实机验证。

### Zone 系统凭据编辑

Zone 编辑增加成对系统凭据更新，映射 modify_zone 的 --access-key/--secret。
创建与编辑共用参数校验和敏感位置标记，组合重命名时敏感标记附在 modify
后续命令上；凭据省略则保持原值，不回填已有密钥。测试覆盖单独更新、重命名
组合及执行前拒绝不完整凭据，沿用审计脱敏；未实机验证。

### Zonegroup 更新后的关联刷新

Zone 的 zonegroup_memberships 由同次采集中的 Zonegroup 详情派生。后端
ReconcileKinds 只持久化请求类型，因此 Zonegroup 操作和手动刷新必须同时
请求 rgw_zonegroup、rgw_zone，否则 Zone 页面可能保留旧成员配置。
现已补齐这两条刷新入口；真实集群交互验证仍待完成。

### Zone 编辑的成员读回校验

指定 Zonegroup 的 Zone 编辑在原有 Zone 读回之后，再执行 zonegroup get，
确认组名、新成员名和成员 ID，并比较本次请求的端点、只读、归档类型、
全来源同步以及设为主 Zone 的结果。读回不匹配返回 post_check_failed。
避免原生命令仅返回成功却未修改目标成员时误报成功；同步来源集合、默认
选择和系统密钥值尚未做结果比较。离线测试覆盖字段不匹配与异常返回。

### Zone 同步来源结果校验

原生 zonegroup 成员更新对 sync_from 执行集合插入或删除，不替换整组。
读回校验现按本次操作检查指定来源全部存在（追加）或全部不存在（删除），
允许保留其它来源，拒绝缺失、null 和非法类型的列表。离线测试覆盖追加、
删除和异常返回；默认选择及密钥值的结果比较仍待补齐。

### Zone 凭据读回验证

Zone 创建或编辑携带系统凭据时，校验 zone get 返回的名称、ID 和 system_key
中的 Access/Secret Key 是否与请求一致。仅在内存比较，不将返回密钥写入结果
或错误消息；不匹配统一返回 post_check_failed。离线测试覆盖创建、重命名、
旧密钥和缺失字段；未实机验证。

### Zone 写操作身份校验

Zone 创建和编辑统一验证 zone get 返回的名称及非空原生 ID，不再仅在请求
携带凭据时验证身份。没有凭据的普通更新也会拒绝空返回、缺少 ID 或错误对象，
携带凭据时继续比较密钥。离线测试覆盖有效无凭据详情及各类错误身份。

### CephFS 子卷克隆状态与取消

对照 volumes 模块的 `fs subvolume info`、`fs clone status` 和
`fs clone cancel`，采集默认组及所有命名组中的子卷。以 info 的 type=clone 识别克隆，
不以 source 是否存在作为身份判断；来源缺失或为 N/A 仍查询克隆状态，保留快照的
已删除克隆不查询目录克隆状态。向页面提供 state、source、progress_report 和 failure。
子卷与快照缓存键包含组名，更新、删除、快照创建/删除也保留组作用域，避免同名
资源碰撞或操作错误对象。

CephFS 快照页新增克隆操作，支持源组、目标组和 pool layout，分别映射
`--group_name`、`--target_group_name` 和 `--pool_layout`。快照列表逐项调用
`fs subvolume snapshot info`，展示创建时间、数据池和待处理克隆标记。子卷页仅对 pending 或
in-progress 状态开放取消操作，调用 `fs clone cancel` 后以 `fs clone status`
读回。离线测试覆盖命名组采集、状态字段和精确命令参数；未在真实集群验证。

原生进度使用 `percentage cloned`、`amount cloned`、`files cloned`，来自 volumes 的
stats_util 和 `_get_clone_progress_report`。页面展示百分比进度条、原生容量比及文件比，
缺失报告不假定为 0% 或完成。来源按文件系统/组/子卷@快照展示，N/A 明确标记不可用；
错误展示原生 error_msg 与 errno。离线夹具改用真实字段，测试覆盖采集→API、类型识别、
缺失来源和页面组件输出。取消条件仍只使用成功读取的 clone_state。

### CephFS 子卷挂载命令

对照参考子卷 Attach 与 cephfs-mount-details，已就绪子卷的详情展示可复制的内核、
FUSE 和 NFS 挂载模板。FSID 使用集群 API 的发现数据（原生 status/discovery），
路径使用 `fs subvolume info` 经缓存/API 提供的 path，不推导子卷目录，也不复用服务端
客户端用户名或密钥。缺失/无效 FSID、路径或错误集群身份时不生成命令；仅 complete
状态展示入口。FUSE 按本地文档使用当前 `--client_fs` 参数，不复制旧兼容选项。
客户端名称与挂载目录可编辑，shell 参数统一引用，离线测试验证特殊路径不会被 shell
解释为命令。客户端仍需准备对应集群 ceph.conf/keyring 和本地目录；本功能不执行
任何挂载。NFS 模板明确要求已有导出并手动替换主机、端口、导出路径，不假设子卷
内部路径就是 NFS export。测试覆盖命令参数、缺失信息、集群身份及组件输出。

### CephFS 快照克隆依赖

对照快照列表的 Pending Clones 列与 volumes `snapshot_info/get_pending_clones`，
快照详情展示 `pending_clones` 的目标子卷和 target_group，省略目标组表示 `_nogroup`；
列表把 has_pending_clones 的 yes/no 显示为中文状态，并展示 orphan_clones_count。
原生 info 已由 `fs subvolume snapshot info` 采集，不新增额外命令。未知、无依赖、
目标列表缺失和孤儿记录分别显示，不能把未知当作无依赖；有待处理目标或孤儿记录
时禁用删除并提示刷新和检查集群，不自动取消克隆或清理索引。原生 snapshot rm 仍
负责执行时的依赖检查，缓存提示不能替代它。离线 API 测试核对组范围和完整原生
依赖数据，前端测试覆盖默认组、命名组、未知数据及组件输出；未连接真实集群。

### CephFS 子卷高级删除

对照参考界面删除确认框与 `fs subvolume rm` 原生参数，子卷页新增高级删除，
可选择 `--retain-snapshots` 保留已有快照，并可用 `--force` 删除失败或已取消的
克隆。请求仍携带资源版本前置条件并进行二次危险操作确认；普通删除入口保持简洁。
离线测试覆盖命名组、两个开关的精确顺序及作用域读回；未在真实集群验证。

### CephFS 组范围读取与版本检查修正

再次逐层检查发现，采集器的子卷和快照 natural_key 已包含组名，但通用接口的
详情查找、If-Match 检查与操作锁仍省略组名。现统一使用
`filesystem/group/subvolume[/snapshot]` 的缓存键；未指定组时明确使用 `_nogroup`。
命令参数仍携带请求中的原生 group，不以另一个组的同名缓存资源进行版本校验。

快照列表不再在未选子卷时错误设置 `/` 父键，集群范围页面能显示全部快照；
选择文件系统、子卷与组时使用精确父键，其他范围通过原生 fs/group/subvolume 字段过滤。
快照表格增加组名列，以区分不同组的同名快照。

离线 API 测试同时建立默认组、team-a、team-b 中的同名子卷和快照，单独提升 team-b
资源版本，验证详情读取、全量/范围列表、错误版本拒绝、异步执行时再次检查版本、
组范围锁以及 resize、snapshot rm、clone cancel 的准确 argv。未连接真实集群。

### CephFS 子卷完整创建参数

子卷创建改用原生文档中的命名参数，不再把内部保留名 `_nogroup` 作为显式组传入。
默认组会省略 `--group_name`，UID/GID 留空时继承父组，不再隐式固定为 0。
表单和 API 新增 earmark、Unicode normalization 与 case-sensitive 选项，并映射为
`--earmark`、`--normalization` 和 `--casesensitive`。离线测试覆盖默认组、命名组、
可选所有者以及完整参数组合；未在真实集群验证。

### CephFS 子卷摘要字段

参考子卷列表使用 `fs subvolume info` 的 bytes_pcent 和 created_at 展示配额使用率
与创建时间。子卷表格展示 bytes_pcent，并将原生 created_at 规范为 ceph_created_at，
避免资源信封的同名缓存记录创建时间覆盖它。子卷组及快照使用同样的字段规范，
表格统一格式化原生时间，原生 info 不返回时间时显示缺失，不用缓存时间代替。
无需新增命令或推导近似值；离线测试验证字段隔离，未连接真实集群。

### CephFS 资源展示与筛选字段一致性

子卷组的文件系统范围由采集器写入 filesystem，移除不存在的 fs 空列。通用资源
列表现在与 FieldColumn 定义一致：筛选选项请求、表格列筛选键及列表 API 请求使用
filterKey 指定的原生字段，filterKey=false 的列不请求选项也不显示筛选器。快照孤儿
记录列使用原生省略含义计算显示 0，因此禁用此展示列的筛选，避免把显示 0 当作
后端真实存在的字段值。测试验证字段去重、原生字段映射、禁用筛选及实际 API 参数；
不改变采集命令或缓存字段结构。

### CephFS 子卷就绪状态

子卷列表展示 `fs subvolume info` 的 state、type 和 pool_namespace，区分普通子卷、
克隆子卷和隔离命名空间。资源 status 改为原生 state；缺失时为 unknown，不再固定
为 available。参考 volumes 的 V2 info 在 snapshot-retained 状态只返回类型、特性和
状态：子卷目录已删除但快照保留。因此页面显示明确警告，只有 complete 状态开放
配额编辑和目录快照可见性入口，不对保留快照的子卷发起目录可见性读取。保留快照
仍可在快照页面管理，删除和克隆取消仍按各自规则处理。后端原生命令继续验证操作
结果；页面就绪限制不替代后端权限、版本校验或原生命令的实际状态检查。
离线测试覆盖原生采集→数据库→API 的 complete、snapshot-retained、unknown 状态、
克隆状态与命名空间，以及全部状态的页面就绪规则。未连接真实集群。

### CephFS 配额使用率和权限展示

子卷与子卷组列表根据原生 info 的 bytes_quota、bytes_used、bytes_pcent 展示二进制
容量及使用率进度条。`infinite` 显示“无限制”，不再向用户显示 `undefined`；缺失的
使用率保持不可用，不用缺失信息推导百分比。超过 100% 时保留原生百分比并告警，
进度条宽度限制为 100%。整数容量超过 JavaScript 安全精度时只接受精确数字字符串，
不展示已经失真的数字。原生 mode 展示 rwx 和八进制权限，并保留 setuid、setgid、
sticky 位以及 0000 无权限状态。数据仍由 `fs subvolume info` 和
`fs subvolumegroup info` 经现有采集/API 链路提供；离线测试覆盖格式和组件输出。

### CephFS 子卷配额更新

子卷编辑支持取消配额限制和禁止缩容。取消限制生成原生大小 `inf`，保护开关生成
`--no_shrink`，并继续保留命名组作用域。表单从 bytes_quota 预填当前状态，API 使用
独立的精确请求结构；离线测试覆盖无限配额、缩容保护和缺失大小校验。

组/子卷配额编辑使用十进制文本输入，更新 API 的 size 为正整数字符串，不接受
旧数值契约。前后端范围统一为 1..9223372036854775807 字节，命令 argv 原样保留
精确值。采集器在 json.Number 尚未失真时将 bytes_quota/bytes_used 转为字符串，
缓存与 API 继续保持精度。取消配额时隐藏缩容保护并发送 no_shrink=false，后端也
不生成无意义的无限配额保护参数，与原生 `_resize` 一致。保护文案为“不允许配额
低于已用空间”，而非禁止一切配额缩小。离线测试覆盖大整数采集→API、异步操作→
准确 argv、数值契约拒绝、上下界、异常值和表单预填；未连接真实集群。

### CephFS 子卷组编辑

参考 `cephfs-subvolumegroup-form.component.ts` 的编辑流程通过 create API 更新已有组的
数据池、UID/GID、权限及配额。本地 volumes 模块 `create_subvolume_group` 对已有组
调用 `set_group_attrs`；未提供 UID/GID 时会设为 0，未提供 pool 时重新继承祖先布局，
mode 也有默认值。因此属性更新要求四个字段同时提供，不把省略字段当作“不修改”。

本项目 `PATCH /filesystem/subvolume/group` 新增可选的 pool、uid、gid、mode，
并支持 unlimited 和 no_shrink。配额使用 `fs subvolumegroup resize`，取消限制生成
`inf`；保护开关实际禁止配额低于已用空间，不是禁止一切缩小。属性更新用
`fs subvolumegroup create` 的命名参数，省略 size，避免覆盖刚设置的配额；
属性操作前以 info 确认组存在，最后核对 bytes_quota、data_pool、uid、gid 和权限位。

表单从当前 info 字段预填，mode 去除目录类型位后显示八进制；缺失或超出浏览器安全
整数范围的配额不自动舍入预填。用户可仅更新配额，或显式选择同时更新属性。
数据池布局更新不迁移既有子卷，配额与属性分步执行，不承诺原子事务；确认框提示
后续步骤失败可能已有部分修改生效。离线测试覆盖命令顺序、更新前存在性检查、
逐字段读回、权限位、部分失败、严格 API 契约及版本检查；未在真实集群验证。

### CephFS 文件系统启用状态与创建时间

参考 `cephfs-list.component.ts` 显示 mdsmap.enabled 与 mdsmap.created。
本地 `MDSMap::dump` 在 `fs dump --format json` 输出这两个原生字段，现由采集器
保留为 enabled 与 created，经 reconciler、资源缓存及文件系统 GET/list 接口提供给
列表和详情页。enabled 使用可空布尔，不根据 MDS 数量或缓存 status 推测；false
明确显示未启用，缺失显示未知。created 保留原生字符串，与 CephTower 缓存记录的
created_at 分开，缺失时不填充当前时间。

离线测试走真实 NativeProvider → reconciler → store → API 链路，验证 enabled 的
true/false/缺失、Ceph 创建时间原值与本地缓存时间分离，并拒绝字符串伪布尔响应。
前端单元检查确认缺失和非布尔值不会误显示已启用；未在真实集群验证。

### CephFS 扩展访问权限

对照参考授权弹窗，`fs authorize` 表单新增 quota、snapshot 和 root squash。
读写权限可组合为 `rwp`、`rws` 或 `rwps`，root squash 作为独立 capability 参数；
只读模式忽略仅适用于写权限的 p/s 选择。离线测试覆盖完整组合和只读约束。

### CephFS 子卷快照可见性

对照参考 `cephfs-subvolume-form`、`cephfs-subvolume.service.ts` 和 controller 的
snapshot-visibility GET/PUT，子卷详情新增实时读取与允许快照浏览开关。
本项目接口为 `GET/PUT /filesystem/subvolume/snapshot/visibility`，请求明确携带
cluster_id、fs、subvolume、group；PUT 的 visible 必须为 JSON boolean。
调用 `ceph fs subvolume snapshot_visibility get/set`，默认组省略 --group_name，
命名组始终保留；设置后读回 0/1 并核对目标值，失败不报告成功。

本地参考 Dashboard 有此功能，但本地 volumes 模块未包含该 CLI 定义，因此另外
核对了[上游模块源码](https://github.com/ceph/ceph/blob/main/src/pybind/mgr/volumes/module.py)
及[原生命令文档](https://docs.ceph.com/en/latest/cephfs/fs-volumes/#controlling-subvolume-snapshot-visibility)。
开关只在客户端启用 client_respect_subvolume_snapshot_visibility 后生效，目前仅支持
FUSE/libcephfs；不会自动修改客户端配置，也不代表删除快照。页面明确说明限制，旧集群
命令不支持、输出异常或命令失败时展示未知及错误，不默认当作开启。
离线测试覆盖默认/命名组、布尔契约、非法作用域、失败读回和异步 API 全链路；
未在真实 Ceph 集群验证。
