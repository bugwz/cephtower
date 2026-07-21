# Restful

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph restful create-key`

- 组件/模块：`mgr module` / `restful`
- 支持版本：16.2.15, 17.2.9, 18.2.8, 19.2.5
- 权限：`rw`
- 来源：`src/pybind/mgr/restful/module.py`
- 标志：`mgr`
- 含义：Create an API key with this name

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `key_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph restful create-self-signed-cert`

- 组件/模块：`mgr module` / `restful`
- 支持版本：16.2.15, 17.2.9, 18.2.8, 19.2.5
- 权限：`rw`
- 来源：`src/pybind/mgr/restful/module.py`
- 标志：`mgr`
- 含义：Create localized self signed certificate

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph restful delete-key`

- 组件/模块：`mgr module` / `restful`
- 支持版本：16.2.15, 17.2.9, 18.2.8, 19.2.5
- 权限：`rw`
- 来源：`src/pybind/mgr/restful/module.py`
- 标志：`mgr`
- 含义：Delete an API key with this name

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `key_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph restful list-keys`

- 组件/模块：`mgr module` / `restful`
- 支持版本：16.2.15, 17.2.9, 18.2.8, 19.2.5
- 权限：`r`
- 来源：`src/pybind/mgr/restful/module.py`
- 标志：`mgr`
- 含义：List all API keys

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph restful restart`

- 组件/模块：`mgr module` / `restful`
- 支持版本：16.2.15, 17.2.9, 18.2.8, 19.2.5
- 权限：`rw`
- 来源：`src/pybind/mgr/restful/module.py`
- 标志：`mgr`
- 含义：Restart API server

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
