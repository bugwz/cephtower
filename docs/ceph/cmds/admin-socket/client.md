# Client

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph daemon client.<id> dump_cache`

- 组件/模块：`admin socket` / `client`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/client/Client.cc`
- 标志：`admin-socket`
- 含义：show in-memory metadata cache contents

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

## `ceph daemon client.<id> kick_stale_sessions`

- 组件/模块：`admin socket` / `client`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/client/Client.cc`
- 标志：`admin-socket`
- 含义：kick sessions that were remote reset

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

## `ceph daemon client.<id> mds_requests`

- 组件/模块：`admin socket` / `client`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/client/Client.cc`
- 标志：`admin-socket`
- 含义：show in-progress mds requests

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

## `ceph daemon client.<id> mds_sessions`

- 组件/模块：`admin socket` / `client`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/client/Client.cc`
- 标志：`admin-socket`
- 含义：show mds session state

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cap_dump` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 直接连接本机 daemon admin socket。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`。
- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph daemon client.<id> objecter_requests`

- 组件/模块：`admin socket` / `client`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/osdc/Objecter.cc`
- 标志：`admin-socket`
- 含义：show in-progress osd requests

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

## `ceph daemon client.<id> status`

- 组件/模块：`admin socket` / `client`
- 支持版本：16.2.15 - 20.2.2
- 权限：`admin-socket`
- 来源：`src/client/Client.cc`
- 标志：`admin-socket`
- 含义：show overall client status

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
