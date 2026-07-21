# Prometheus

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph healthcheck history clear`

- 组件/模块：`mgr module` / `prometheus`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/prometheus/module.py`
- 标志：`mgr`
- 含义：Clear the healthcheck history

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph healthcheck history ls`

- 组件/模块：`mgr module` / `prometheus`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/prometheus/module.py`
- 标志：`mgr`
- 含义：List all the healthchecks being tracked The format options are parsed in ceph_argparse, before they get evaluated here so we can safely assume that what we have to process is valid. ceph_argparse will throw a ValueError if the cast to our Format class fails. Args: format (Format, optional): output format. Defaults to Format.plain. Returns: HandleCommandResult: return code, stdout and stderr returned to the caller

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph prometheus file_sd_config`

- 组件/模块：`mgr module` / `prometheus`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/prometheus/module.py`
- 标志：`mgr`
- 含义：Return file_sd compatible prometheus config for mgr cluster

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
