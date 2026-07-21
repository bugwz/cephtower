# Common

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph daemon <daemon>.<id> 0`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/admin_socket.cc`
- 标志：`admin-socket`
- 含义：-

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

## `ceph daemon <daemon>.<id> 1`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：-

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

## `ceph daemon <daemon>.<id> 2`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：-

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

## `ceph daemon <daemon>.<id> abort`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：-

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

## `ceph daemon <daemon>.<id> assert`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：-

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

## `ceph daemon <daemon>.<id> config diff`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：dump diff of current config and default config

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

## `ceph daemon <daemon>.<id> config diff get`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：dump diff get <field>: dump diff of current and default config setting <field>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `var` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon <daemon>.<id> config get`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：config get <field>: get the config value

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `var` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon <daemon>.<id> config help`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：get config setting schema and descriptions

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `var` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon <daemon>.<id> config set`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：config set <field> <val> [<val> ...]: set a config variable

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `var` | `CephString` | 是 | 是 | - |
| `val` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon <daemon>.<id> config show`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：dump current config settings

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

## `ceph daemon <daemon>.<id> config unset`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：config unset <field>: unset a config variable

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `var` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon <daemon>.<id> counter dump`

- 组件/模块：`admin socket` / `common`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：dump all labeled and non-labeled counters and their values

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

## `ceph daemon <daemon>.<id> counter schema`

- 组件/模块：`admin socket` / `common`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：dump all labeled and non-labeled counters schemas

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

## `ceph daemon <daemon>.<id> dump_mempools`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：get mempool stats

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

## `ceph daemon <daemon>.<id> get_command_descriptions`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/admin_socket.cc`
- 标志：`admin-socket`
- 含义：list available commands

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

## `ceph daemon <daemon>.<id> git_version`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/admin_socket.cc`
- 标志：`admin-socket`
- 含义：get git sha1

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

## `ceph daemon <daemon>.<id> help`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/admin_socket.cc`
- 标志：`admin-socket`
- 含义：list available commands

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

## `ceph daemon <daemon>.<id> injectargs`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：inject configuration arguments into running daemon

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `injected_args` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon <daemon>.<id> leak_some_memory`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：-

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

## `ceph daemon <daemon>.<id> log dump`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：dump recent log entries to log file

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

## `ceph daemon <daemon>.<id> log flush`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：flush log entries to log file

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

## `ceph daemon <daemon>.<id> log reopen`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：reopen log file

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

## `ceph daemon <daemon>.<id> perf dump`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：dump perfcounters value

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `logger` | `CephString` | 否 | 是 | - |
| `counter` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon <daemon>.<id> perf histogram dump`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：dump perf histogram values

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `logger` | `CephString` | 否 | 是 | - |
| `counter` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon <daemon>.<id> perf histogram schema`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：dump perf histogram schema

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

## `ceph daemon <daemon>.<id> perf reset`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：perf reset <name>: perf reset all or one perfcounter name

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `var` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon <daemon>.<id> perf schema`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：dump perfcounters schema

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

## `ceph daemon <daemon>.<id> perfcounters_dump`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：-

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

## `ceph daemon <daemon>.<id> perfcounters_schema`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/ceph_context.cc`
- 标志：`admin-socket`
- 含义：-

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

## `ceph daemon <daemon>.<id> version`

- 组件/模块：`admin socket` / `common`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/common/admin_socket.cc`
- 标志：`admin-socket`
- 含义：get ceph version

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
