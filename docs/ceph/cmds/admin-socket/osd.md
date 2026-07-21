# Osd

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph daemon osd.<id> bench`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：OSD benchmark: write <count> <size>-byte objects(with <obj_size> <obj_num>), (default count=1G default size=4MB). Results in log.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `count` | `CephInt` | 否 | 是 | - |
| `size` | `CephInt` | 否 | 是 | - |
| `object_size` | `CephInt` | 否 | 是 | - |
| `object_num` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> cache drop`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Drop all OSD caches

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> cache status`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Get OSD caches statistics

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> calc_objectstore_db_histogram`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Generate key value histogram of kvdb(rocksdb) which used by bluestore

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> clear_shards_repaired`

- 组件/模块：`admin socket` / `osd`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：clear num_shards_repaired to clear health warning

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `count` | `CephInt` | 否 | 是 | 范围: 0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> cluster_log`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：log a message to the cluster log

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `level` | `CephChoices` | 是 | 是 | 可选值: error |
| `message` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> compact`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Commpact object store's omap. WARNING: Compaction probably slows your requests

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> cpu_profiler`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：run cpu profiling on daemon

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `arg` | `CephChoices` | 是 | 是 | 可选值: status, flush |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> debug dump_missing`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：dump missing objects to a named file

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `filename` | `CephFilepath` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> debug kick_recovery_wq`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：set osd_recovery_delay_start to <val>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `delay` | `CephInt` | 是 | 是 | 范围: 0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> deep-scrub`

- 组件/模块：`admin socket` / `osd`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Trigger a deep scrub

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> deep_scrub`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15, 17.2.9, 18.2.8
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Trigger a scheduled deep scrub

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 否 | 是 | - |
| `time` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_blocked_ops`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show the blocked ops currently in flight

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `filterstr` | `CephString` | 否 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_blocked_ops_count`

- 组件/模块：`admin socket` / `osd`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show the count of blocked ops currently in flight

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `filterstr` | `CephString` | 否 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_blocklist`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：dump blocklisted clients and times

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_historic_ops`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show recent ops

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `filterstr` | `CephString` | 否 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_historic_ops_by_duration`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show slowest recent ops, sorted by duration

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `filterstr` | `CephString` | 否 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_historic_slow_ops`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show slowest recent ops

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `filterstr` | `CephString` | 否 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_objectstore_kv_stats`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：print statistics of kvdb which used by bluestore

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_op_pq_state`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：dump op priority queue state

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_ops_in_flight`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show the ops currently in flight

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `filterstr` | `CephString` | 否 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_osd_network`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Dump osd heartbeat network ping times

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `value` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_osd_pg_stats`

- 组件/模块：`admin socket` / `osd`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Dump OSD PGs' statistics

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_pg_recovery_stats`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：dump pg recovery statistics

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_pgstate_history`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show recent state history

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_pool_statfs`

- 组件/模块：`admin socket` / `osd`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Dump store's statistics for the given pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `poolid` | `CephInt` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_recovery_reservations`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show recovery reservations

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_scrub_reservations`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show scrub reservations

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_scrubs`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：print scheduled scrubs

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> dump_watchers`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show clients which have active watches, and on which objects

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> flush_journal`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：flush the journal to permanent store

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> flush_pg_stats`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：flush pg stats

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> flush_store_cache`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Flush bluestore internal cache

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> get_heap_property`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：get malloc extension heap property

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `property` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> get_latest_osdmap`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：force osd to update the latest map from the mon

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> get_mapped_pools`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：dump pools whose PG(s) are mapped to this OSD.

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> getomap`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：output entire object map

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `objname` | `CephObjectname` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> heap`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show heap usage info (available only if compiled with tcmalloc)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `heapcmd` | `CephChoices` | 是 | 是 | 可选值: dump, start_profiler, stop_profiler, release, get_release_rate, set_release_rate, stats |
| `value` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> injectdataerr`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：inject data error to an object

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `objname` | `CephObjectname` | 是 | 是 | - |
| `shardid` | `CephInt` | 否 | 是 | 范围: 0..255 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> injectecclearreaderr`

- 组件/模块：`admin socket` / `osd`
- 支持版本：20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：clear read error injects for object in an EC pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `objname` | `CephObjectname` | 是 | 是 | - |
| `shardid` | `CephInt` | 是 | 是 | 范围: 0..255 |
| `type` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> injectecclearwriteerr`

- 组件/模块：`admin socket` / `osd`
- 支持版本：20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：clear write error inject for object in an EC pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `objname` | `CephObjectname` | 是 | 是 | - |
| `shardid` | `CephInt` | 是 | 是 | 范围: 0..255 |
| `type` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> injectecreaderr`

- 组件/模块：`admin socket` / `osd`
- 支持版本：20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：inject error for read of object in an EC pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `objname` | `CephObjectname` | 是 | 是 | - |
| `shardid` | `CephInt` | 是 | 是 | 范围: 0..255 |
| `type` | `CephInt` | 否 | 是 | - |
| `when` | `CephInt` | 否 | 是 | - |
| `duration` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> injectecwriteerr`

- 组件/模块：`admin socket` / `osd`
- 支持版本：20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：inject error for write of object in an EC pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `objname` | `CephObjectname` | 是 | 是 | - |
| `shardid` | `CephInt` | 是 | 是 | 范围: 0..255 |
| `type` | `CephInt` | 否 | 是 | - |
| `when` | `CephInt` | 否 | 是 | - |
| `duration` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> injectfull`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Inject a full disk (optional count times)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `type` | `CephString` | 否 | 是 | - |
| `count` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> injectmdataerr`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：inject metadata error to an object

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `objname` | `CephObjectname` | 是 | 是 | - |
| `shardid` | `CephInt` | 否 | 是 | 范围: 0..255 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> list_devices`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：list OSD devices.

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> list_unfound`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：list unfound objects on this pg, perhaps starting at an offset given in JSON

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 否 | 是 | - |
| `offset` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> log`

- 组件/模块：`admin socket` / `osd`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：dump pg_log of a specific pg

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> mark_unfound_lost`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：mark all unfound objects in this pg as lost, either removing or reverting to a prior version if one is available

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 否 | 是 | - |
| `mulcmd` | `CephChoices` | 是 | 是 | 可选值: revert, delete |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> ops`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show the ops currently in flight

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `filterstr` | `CephString` | 否 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> pg`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |
| `cmd` | `CephChoices` | 是 | 是 | 可选值: deep-scrub |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

### 版本差异

- `16.2.15` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |
| `cmd` | `CephChoices` | 是 | 是 | 可选值: deep_scrub |
| `time` | `CephInt` | 否 | 是 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |
| `cmd` | `CephChoices` | 是 | 是 | 可选值: deep_scrub |
| `time` | `CephInt` | 否 | 是 | - |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |
| `cmd` | `CephChoices` | 是 | 是 | 可选值: deep_scrub |
| `time` | `CephInt` | 否 | 是 | - |

## `ceph daemon osd.<id> query`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：show details of a specific pg

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> reset_pg_recovery_stats`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：reset pg recovery statistics

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> reset_purged_snaps_last`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Reset the superblock's purged_snaps_last

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> rmomapkey`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：remove omap key

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `objname` | `CephObjectname` | 是 | 是 | - |
| `key` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> rotate-stored-key`

- 组件/模块：`admin socket` / `osd`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Update the stored osd_key

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> schedule-deep-scrub`

- 组件/模块：`admin socket` / `osd`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Schedule a deep scrub

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 否 | 是 | - |
| `time` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> schedule-scrub`

- 组件/模块：`admin socket` / `osd`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Schedule a scrub

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 否 | 是 | - |
| `time` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> scrub`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Trigger a scheduled scrub

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

### 版本差异

- `16.2.15` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 否 | 是 | - |
| `time` | `CephInt` | 否 | 是 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 否 | 是 | - |
| `time` | `CephInt` | 否 | 是 | - |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 否 | 是 | - |
| `time` | `CephInt` | 否 | 是 | - |

## `ceph daemon osd.<id> scrub-abort`

- 组件/模块：`admin socket` / `osd`
- 支持版本：20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Abort an ongoing scrub. Cancel any operator-initiated scrub

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> scrub_purged_snaps`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Scrub purged_snaps vs snapmapper index

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> scrubdebug`

- 组件/模块：`admin socket` / `osd`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：debug the scrubber

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |
| `cmd` | `CephChoices` | 是 | 是 | 可选值: block, unblock, set, unset |
| `value` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> send_beacon`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：send OSD beacon to mon immediately

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> set_heap_property`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：update malloc extension heap property

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `property` | `CephString` | 是 | 是 | - |
| `value` | `CephInt` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> set_recovery_delay`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：Delay osd recovery by specified seconds

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `utime` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> setomapheader`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：set omap header

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `objname` | `CephObjectname` | 是 | 是 | - |
| `header` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> setomapval`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：set omap key

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `objname` | `CephObjectname` | 是 | 是 | - |
| `key` | `CephString` | 是 | 是 | - |
| `val` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> smart`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：probe OSD devices for SMART data.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `devid` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> status`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：high-level status of OSD

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> trim stale osdmaps`

- 组件/模块：`admin socket` / `osd`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：cleanup any existing osdmap from the store in the range of 0 up to the superblock's oldest_map.

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon osd.<id> truncobj`

- 组件/模块：`admin socket` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osd/OSD.cc`
- 标志：`admin-socket`
- 含义：truncate object to length

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `objname` | `CephObjectname` | 是 | 是 | - |
| `len` | `CephInt` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
