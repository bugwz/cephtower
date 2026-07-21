# Bluestore

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph tell osd.<id> bluefs debug_inject_read_zeros`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/BlueFS.cc`
- 标志：`tell, admin-socket`
- 含义：Injects 8K zeros into next BlueFS read. Debug only.

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph tell osd.<id> bluefs files list`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/BlueFS.cc`
- 标志：`tell, admin-socket`
- 含义：print files in bluefs

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph tell osd.<id> bluefs stats`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/BlueFS.cc`
- 标志：`tell, admin-socket`
- 含义：Dump internal statistics for bluefs.

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph tell osd.<id> bluestore allocator dump`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/Allocator.cc`
- 标志：`tell, admin-socket`
- 含义：dump allocator free regions

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph tell osd.<id> bluestore allocator fragmentation`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/Allocator.cc`
- 标志：`tell, admin-socket`
- 含义：give allocator fragmentation (0-no fragmentation, 1-absolute fragmentation)

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph tell osd.<id> bluestore allocator fragmentation histogram`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/Allocator.cc`
- 标志：`tell, admin-socket`
- 含义：build allocator free regions state histogram

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `alloc_unit` | `CephInt` | 否 | 是 | - |
| `num_buckets` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph tell osd.<id> bluestore allocator score`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/Allocator.cc`
- 标志：`tell, admin-socket`
- 含义：give score on allocator fragmentation (0-no fragmentation, 1-absolute fragmentation)

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph tell osd.<id> bluestore bluefs device info`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/BlueFS.cc`
- 标志：`tell, admin-socket`
- 含义：Shows space report for bluefs devices. This also includes an estimation for space available to bluefs at main device. alloc_size, if set, specifies the custom bluefs allocation unit size for the estimation above.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `alloc_size` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph tell osd.<id> bluestore collections`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/BlueAdmin.cc`
- 标志：`tell, admin-socket`
- 含义：list all collections

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph tell osd.<id> bluestore compression stats`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/BlueAdmin.cc`
- 标志：`tell, admin-socket`
- 含义：print compression stats, per collection

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `collection` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph tell osd.<id> bluestore list`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/BlueAdmin.cc`
- 标志：`tell, admin-socket`
- 含义：list objects in specific collection

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `collection` | `CephString` | 是 | 是 | - |
| `start` | `CephString` | 否 | 是 | - |
| `max` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph tell osd.<id> bluestore onode metadata`

- 组件/模块：`ceph tell` / `bluestore`
- 支持版本：20.2.2
- 权限：`rw`
- 来源：`src/os/bluestore/BlueAdmin.cc`
- 标志：`tell, admin-socket`
- 含义：print object internals

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `object_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; monitor 将请求转发给目标 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
