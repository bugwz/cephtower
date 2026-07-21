# Mds

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph daemon mds.<id> cache drop`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：trim cache and optionally request client to release all caps and flush the journal

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `timeout` | `CephInt` | 否 | 是 | 范围: 0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> cache status`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：show cache status

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

## `ceph daemon mds.<id> client config`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Config a CephFS client session

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `client_id` | `CephInt` | 是 | 是 | - |
| `option` | `CephString` | 是 | 是 | - |
| `value` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> client evict`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Evict client session(s) based on a filter

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `filters` | `CephString` | 是 | 是 | 可重复 |

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
| `filters` | `CephString` | 否 | 是 | 可重复 |

## `ceph daemon mds.<id> client ls`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：List client sessions based on a filter

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cap_dump` | `CephBool` | 否 | 否 | - |
| `filters` | `CephString` | 否 | 否 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> cpu_profiler`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
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

## `ceph daemon mds.<id> damage ls`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：List detected metadata damage

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

## `ceph daemon mds.<id> damage rm`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Remove a damage table entry

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `damage_id` | `CephInt` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> dirfrag ls`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：List fragments in directory

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> dirfrag merge`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：De-fragment directory by path

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `frag` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> dirfrag split`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Fragment directory by path

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `frag` | `CephString` | 是 | 是 | - |
| `bits` | `CephInt` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> dump cache`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：dump metadata cache (optionally to a file)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 否 | 是 | - |
| `timeout` | `CephInt` | 否 | 是 | 范围: 0 |

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
| `path` | `CephString` | 否 | 是 | - |

## `ceph daemon mds.<id> dump dir`

- 组件/模块：`admin socket` / `mds`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：dump directory by path

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `dentry_dump` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> dump inode`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：dump inode by inode number

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `number` | `CephInt` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> dump loads`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：dump metadata loads

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `depth` | `CephInt` | 否 | 是 | 范围: 0 |

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

无。

- `17.2.9` 参数：

无。

## `ceph daemon mds.<id> dump snaps`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：dump snapshots

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `server` | `CephChoices` | 否 | 是 | 可选值: --server |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> dump stray`

- 组件/模块：`admin socket` / `mds`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：dump stray folder content

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

## `ceph daemon mds.<id> dump tree`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：dump metadata cache for subtree

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `root` | `CephString` | 是 | 是 | - |
| `depth` | `CephInt` | 否 | 是 | - |
| `path` | `CephString` | 否 | 是 | - |

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
| `root` | `CephString` | 是 | 是 | - |
| `depth` | `CephInt` | 否 | 是 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `root` | `CephString` | 是 | 是 | - |
| `depth` | `CephInt` | 否 | 是 | - |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `root` | `CephString` | 是 | 是 | - |
| `depth` | `CephInt` | 否 | 是 | - |

## `ceph daemon mds.<id> dump_blocked_ops`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：show the blocked ops currently in flight

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

## `ceph daemon mds.<id> dump_blocked_ops_count`

- 组件/模块：`admin socket` / `mds`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：show the count of blocked ops currently in flight

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

## `ceph daemon mds.<id> dump_export_states`

- 组件/模块：`admin socket` / `mds`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：dump export states

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

## `ceph daemon mds.<id> dump_historic_ops`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：show recent ops

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

## `ceph daemon mds.<id> dump_historic_ops_by_duration`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：show recent ops, sorted by op duration

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

## `ceph daemon mds.<id> dump_ops_in_flight`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：show the ops currently in flight

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

## `ceph daemon mds.<id> exit`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Terminate this MDS

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

## `ceph daemon mds.<id> export dir`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：migrate a subtree to named MDS

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `rank` | `CephInt` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> flush journal`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Flush the journal to the backing store

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

## `ceph daemon mds.<id> flush_path`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：flush an inode (and its dirfrags)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> force_readonly`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Force MDS to read-only mode

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

## `ceph daemon mds.<id> get subtrees`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Return the subtree map

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

## `ceph daemon mds.<id> heap`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
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

## `ceph daemon mds.<id> lock path`

- 组件/模块：`admin socket` / `mds`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：lock a path

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `locks` | `CephString` | 否 | 是 | 可重复 |
| `ap_dont_block` | `CephBool` | 否 | 否 | - |
| `ap_freeze` | `CephBool` | 否 | 否 | - |
| `await` | `CephBool` | 否 | 否 | - |
| `lifetime` | `CephFloat` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

### 版本差异

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `locks` | `CephString` | 否 | 是 | 可重复 |

## `ceph daemon mds.<id> lockup`

- 组件/模块：`admin socket` / `mds`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：sleep with mds_lock held (dev)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `millisecs` | `CephInt` | 是 | 是 | 范围: 0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> op get`

- 组件/模块：`admin socket` / `mds`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：get op

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `flags` | `CephChoices` | 否 | 是 | 可重复; 可选值: locks |
| `id` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> op kill`

- 组件/模块：`admin socket` / `mds`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：kill op

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> openfiles ls`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：List the opening files and their caps

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

## `ceph daemon mds.<id> ops`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：show the ops currently in flight

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `flags` | `CephChoices` | 否 | 是 | 可重复; 可选值: locks |
| `path` | `CephString` | 否 | 是 | - |

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

无。

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `flags` | `CephChoices` | 否 | 是 | 可重复; 可选值: locks |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `flags` | `CephChoices` | 否 | 是 | 可重复; 可选值: locks |

## `ceph daemon mds.<id> osdmap barrier`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Wait until the MDS has this OSD map epoch

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `target_epoch` | `CephInt` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> quiesce db`

- 组件/模块：`admin socket` / `mds`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：submit queries to the local QuiesceDbManager

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `roots` | `CephString` | 否 | 是 | 可重复 |
| `set_id` | `CephString` | 否 | 否 | - |
| `timeout` | `CephFloat` | 否 | 否 | 范围: 0 |
| `expiration` | `CephFloat` | 否 | 否 | 范围: 0 |
| `await_for` | `CephFloat` | 否 | 否 | 范围: 0 |
| `await` | `CephBool` | 否 | 否 | - |
| `if_version` | `CephInt` | 否 | 否 | 范围: 0 |
| `include` | `CephBool` | 否 | 否 | - |
| `exclude` | `CephBool` | 否 | 否 | - |
| `reset` | `CephBool` | 否 | 否 | - |
| `release` | `CephBool` | 否 | 否 | - |
| `query` | `CephBool` | 否 | 否 | - |
| `all` | `CephBool` | 否 | 否 | - |
| `cancel` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> quiesce path`

- 组件/模块：`admin socket` / `mds`
- 支持版本：19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：quiesce a subtree

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `await` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> respawn`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Respawn this MDS

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

## `ceph daemon mds.<id> scrub abort`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Abort in progress scrub operations(s)

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

## `ceph daemon mds.<id> scrub pause`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Pause in progress scrub operations(s)

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

## `ceph daemon mds.<id> scrub purge_status`

- 组件/模块：`admin socket` / `mds`
- 支持版本：20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Purge status of scrub tag|all

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `tag` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> scrub resume`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Resume paused scrub operations(s)

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

## `ceph daemon mds.<id> scrub start`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：scrub and inode and output results

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `scrubops` | `CephChoices` | 否 | 是 | 可重复; 可选值: force, recursive, repair, scrub_mdsdir |
| `tag` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> scrub status`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Status of scrub operations(s)

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

## `ceph daemon mds.<id> scrub_path`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：scrub an inode and output results

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `scrubops` | `CephChoices` | 否 | 是 | 可重复; 可选值: force, recursive, repair |
| `tag` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> session config`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Config a CephFS client session

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `client_id` | `CephInt` | 是 | 是 | - |
| `option` | `CephString` | 是 | 是 | - |
| `value` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> session evict`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Evict client session(s) based on a filter

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `filters` | `CephString` | 是 | 是 | 可重复 |

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
| `filters` | `CephString` | 否 | 是 | 可重复 |

## `ceph daemon mds.<id> session kill`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Evict a client session by id

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `client_id` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon mds.<id> session ls`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：List client sessions based on a filter

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cap_dump` | `CephBool` | 否 | 否 | - |
| `filters` | `CephString` | 否 | 否 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：object; `sessions` 数组中的 `session` 对象字段由 `MDSRankDispatcher::dump_sessions` 与 `Session::dump` 输出确认。

#### 返回字段

| 字段 | 类型 | 说明 | 源码确认 |
| --- | --- | --- | --- |
| `sessions[]` | `array` | session 列表 | `src/mds/MDSRank.cc` |
| `sessions[].session.id` | `integer` | client/session 编号 | `src/mds/SessionMap.cc` |
| `sessions[].session.entity` | `object` | 客户端 entity 实例 | `src/mds/SessionMap.cc` |
| `sessions[].session.state` | `string` | session 状态 | `src/mds/SessionMap.cc` |
| `sessions[].session.num_leases` | `integer` | lease 数量 | `src/mds/SessionMap.cc` |
| `sessions[].session.num_caps` | `integer` | capability 数量 | `src/mds/SessionMap.cc` |
| `sessions[].session.caps[]` | `array[object]` | `cap_dump=true` 时输出的 capability 列表 | `src/mds/SessionMap.cc` |
| `sessions[].session.request_load_avg` | `integer` | open/stale session 的请求负载均值 | `src/mds/SessionMap.cc` |
| `sessions[].session.uptime` | `float` | session 存活时间 | `src/mds/SessionMap.cc` |
| `sessions[].session.requests_in_flight` | `integer` | 正在处理的请求数 | `src/mds/SessionMap.cc` |
| `sessions[].session.num_completed_requests` | `integer` | 已完成请求数 | `src/mds/SessionMap.cc` |
| `sessions[].session.num_completed_flushes` | `integer` | 已完成 flush 数 | `src/mds/SessionMap.cc` |
| `sessions[].session.reconnecting` | `boolean` | 是否处于 reconnecting | `src/mds/SessionMap.cc` |
| `sessions[].session.recall_caps` | `object` | recall caps 状态对象 | `src/mds/SessionMap.cc` |
| `sessions[].session.release_caps` | `object` | release caps 状态对象 | `src/mds/SessionMap.cc` |
| `sessions[].session.recall_caps_throttle` | `object` | recall caps throttle 状态 | `src/mds/SessionMap.cc` |
| `sessions[].session.recall_caps_throttle2o` | `object` | second-order recall caps throttle 状态 | `src/mds/SessionMap.cc` |
| `sessions[].session.session_cache_liveness` | `object` | session cache liveness 状态 | `src/mds/SessionMap.cc` |
| `sessions[].session.cap_acquisition` | `object` | cap acquisition 状态 | `src/mds/SessionMap.cc` |
| `sessions[].session.last_trim_completed_requests_tid` | `integer` | 最近 trim completed requests tid | `src/mds/SessionMap.cc` |
| `sessions[].session.last_trim_completed_flushes_tid` | `integer` | 最近 trim completed flushes tid | `src/mds/SessionMap.cc` |
| `sessions[].session.delegated_inos[]` | `array[object]` | delegated inode range 列表 | `src/mds/SessionMap.cc` |
| `sessions[].session.delegated_inos[].start` | `string` | inode range 起点 | `src/mds/SessionMap.cc` |
| `sessions[].session.delegated_inos[].length` | `integer` | inode range 长度 | `src/mds/SessionMap.cc` |

### 版本差异

- `16.2.15` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cap_dump` | `CephBool` | 否 | 否 | - |

## `ceph daemon mds.<id> status`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：high-level status of MDS

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

## `ceph daemon mds.<id> tag path`

- 组件/模块：`admin socket` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/mds/MDSDaemon.cc`
- 标志：`admin-socket`
- 含义：Apply scrub tag recursively

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `tag` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
