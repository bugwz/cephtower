# Mgr

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph daemon mgr.<id> dump_blocked_ops`

- 组件/模块：`admin socket` / `mgr`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mgr/DaemonServer.cc`
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

## `ceph daemon mgr.<id> dump_blocked_ops_count`

- 组件/模块：`admin socket` / `mgr`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mgr/DaemonServer.cc`
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

## `ceph daemon mgr.<id> dump_historic_ops`

- 组件/模块：`admin socket` / `mgr`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mgr/DaemonServer.cc`
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

## `ceph daemon mgr.<id> dump_historic_ops_by_duration`

- 组件/模块：`admin socket` / `mgr`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mgr/DaemonServer.cc`
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

## `ceph daemon mgr.<id> dump_historic_slow_ops`

- 组件/模块：`admin socket` / `mgr`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mgr/DaemonServer.cc`
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

## `ceph daemon mgr.<id> dump_ops_in_flight`

- 组件/模块：`admin socket` / `mgr`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mgr/DaemonServer.cc`
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

## `ceph daemon mgr.<id> dump_osd_network`

- 组件/模块：`admin socket` / `mgr`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mgr/ClusterState.cc`
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

## `ceph daemon mgr.<id> mgr_status`

- 组件/模块：`admin socket` / `mgr`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mgr/Mgr.cc`
- 标志：`admin-socket`
- 含义：Dump mgr status

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

## `ceph daemon mgr.<id> status`

- 组件/模块：`admin socket` / `mgr`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mgr/MgrStandby.cc`
- 标志：`admin-socket`
- 含义：show status

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
