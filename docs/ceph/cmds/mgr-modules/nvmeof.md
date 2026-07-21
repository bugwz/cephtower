# Nvmeof

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph nvmeof connection list`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof gateway get_log_level`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof gateway get_stats`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof gateway info`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof gateway listener_info`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof gateway set_log_level`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `log_level` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof gateway version`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof get_subsystems`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof host add`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `host_nqn` | `CephString` | 是 | 是 | - |
| `dhchap_key` | `CephString` | 否 | 是 | - |
| `dhchap_controller_key` | `CephString` | 否 | 是 | - |
| `psk` | `CephString` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof host change_controller_key`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `host_nqn` | `CephString` | 是 | 是 | - |
| `dhchap_controller_key` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof host change_key`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `host_nqn` | `CephString` | 是 | 是 | - |
| `dhchap_key` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof host del`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `host_nqn` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof host del_controller_key`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `host_nqn` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof host del_key`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `host_nqn` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof host list`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `clear_alerts` | `CephBool` | 否 | 否 | - |
| `gw_group` | `CephString` | 否 | 否 | - |
| `traddr` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof listener add`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `host_name` | `CephString` | 是 | 是 | - |
| `traddr` | `CephString` | 是 | 是 | - |
| `trsvcid` | `CephInt` | 否 | 是 | - |
| `adrfam` | `CephInt` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `secure` | `CephBool` | 否 | 否 | - |
| `verify_host_name` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof listener del`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `host_name` | `CephString` | 是 | 是 | - |
| `traddr` | `CephString` | 是 | 是 | - |
| `trsvcid` | `CephInt` | 是 | 是 | - |
| `adrfam` | `CephInt` | 否 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `gw_group` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof listener list`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace add`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `rbd_image_name` | `CephString` | 是 | 是 | - |
| `rbd_pool` | `CephString` | 否 | 是 | - |
| `nsid` | `CephString` | 否 | 是 | - |
| `create_image` | `CephBool` | 否 | 否 | - |
| `size` | `CephString` | 否 | 否 | - |
| `rbd_image_size` | `CephString` | 否 | 否 | - |
| `trash_image` | `CephBool` | 否 | 否 | - |
| `block_size` | `CephInt` | 否 | 否 | - |
| `load_balancing_group` | `CephInt` | 否 | 否 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `no_auto_visible` | `CephBool` | 否 | 否 | - |
| `disable_auto_resize` | `CephBool` | 否 | 否 | - |
| `read_only` | `CephBool` | 否 | 否 | - |
| `gw_group` | `CephString` | 否 | 否 | - |
| `traddr` | `CephString` | 否 | 否 | - |
| `rados_namespace` | `CephString` | 否 | 否 | - |
| `encryption_format` | `CephString` | 否 | 否 | - |
| `encryption_algorithm` | `CephString` | 否 | 否 | - |
| `key_id` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace add_host`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `host_nqn` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `gw_group` | `CephString` | 否 | 否 | - |
| `traddr` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace change_load_balancing_group`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `load_balancing_group` | `CephInt` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace change_visibility`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `auto_visible` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `gw_group` | `CephString` | 否 | 否 | - |
| `traddr` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace del`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |
| `force` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace del_host`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `host_nqn` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace get`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace get_io_stats`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace list`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 否 | 是 | - |
| `nsid` | `CephString` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace refresh_size`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace resize`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `rbd_image_size` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace set_auto_resize`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `auto_resize_enabled` | `CephBool` | 是 | 否 | - |
| `gw_group` | `CephString` | 否 | 否 | - |
| `traddr` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace set_qos`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `rw_ios_per_second` | `CephInt` | 否 | 是 | - |
| `rw_mbytes_per_second` | `CephInt` | 否 | 是 | - |
| `r_mbytes_per_second` | `CephInt` | 否 | 是 | - |
| `w_mbytes_per_second` | `CephInt` | 否 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `gw_group` | `CephString` | 否 | 否 | - |
| `traddr` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace set_rbd_trash_image`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `rbd_trash_image_on_delete` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof namespace update`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `nsid` | `CephString` | 是 | 是 | - |
| `rbd_image_size` | `CephInt` | 否 | 是 | - |
| `load_balancing_group` | `CephInt` | 否 | 是 | - |
| `rw_ios_per_second` | `CephInt` | 否 | 是 | - |
| `rw_mbytes_per_second` | `CephInt` | 否 | 是 | - |
| `r_mbytes_per_second` | `CephInt` | 否 | 是 | - |
| `w_mbytes_per_second` | `CephInt` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |
| `trash_image` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof spdk_log_level disable`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `extra_log_flags` | `CephString` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof spdk_log_level get`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `all_log_flags` | `CephBool` | 否 | 否 | - |
| `gw_group` | `CephString` | 否 | 否 | - |
| `traddr` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof spdk_log_level set`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `log_level` | `CephString` | 否 | 是 | - |
| `print_level` | `CephString` | 否 | 是 | - |
| `extra_log_flags` | `CephString` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof subsystem add`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `enable_ha` | `CephBool` | 否 | 否 | - |
| `max_namespaces` | `CephInt` | 否 | 否 | - |
| `no_group_append` | `CephBool` | 否 | 否 | - |
| `serial_number` | `CephString` | 否 | 否 | - |
| `dhchap_key` | `CephString` | 否 | 否 | - |
| `gw_group` | `CephString` | 否 | 否 | - |
| `traddr` | `CephString` | 否 | 否 | - |
| `network_mask` | `CephString` | 否 | 否 | - |
| `port` | `CephInt` | 否 | 否 | - |
| `secure_listeners` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof subsystem add_kmip_server_endpoint`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `server_name` | `CephString` | 是 | 是 | - |
| `address` | `CephString` | 否 | 是 | - |
| `port` | `CephInt` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof subsystem add_network`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `network_mask` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof subsystem change_key`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `dhchap_key` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof subsystem del`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `force` | `CephString` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof subsystem del_key`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof subsystem del_kmip_server_endpoint`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `server_name` | `CephString` | 是 | 是 | - |
| `address` | `CephString` | 否 | 是 | - |
| `port` | `CephInt` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof subsystem del_network`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `network_mask` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof subsystem get`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 是 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof subsystem list`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof subsystem list_kmip_server_endpoints`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/dashboard/controllers/nvmeof.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 否 | 是 | - |
| `server_name` | `CephString` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `traddr` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof top cpu`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/dashboard/services/nvmeof_top_cli.py`
- 标志：`mgr, poll`
- 含义：NVMeoF Top CPU Tool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `server_address` | `CephString` | 否 | 是 | - |
| `server_port` | `CephInt` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `descending` | `CephBool` | 否 | 否 | - |
| `sort_by` | `CephString` | 否 | 否 | - |
| `with_timestamp` | `CephBool` | 否 | 否 | - |
| `no_header` | `CephBool` | 否 | 否 | - |
| `period` | `CephFloat` | 否 | 否 | - |
| `session_id` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nvmeof top io`

- 组件/模块：`mgr module` / `dashboard`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/dashboard/services/nvmeof_top_cli.py`
- 标志：`mgr, poll`
- 含义：NVMeoF Top IO Tool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `nqn` | `CephString` | 否 | 是 | - |
| `server_address` | `CephString` | 否 | 是 | - |
| `server_port` | `CephInt` | 否 | 是 | - |
| `gw_group` | `CephString` | 否 | 是 | - |
| `descending` | `CephBool` | 否 | 否 | - |
| `sort_by` | `CephString` | 否 | 否 | - |
| `with_timestamp` | `CephBool` | 否 | 否 | - |
| `summary` | `CephBool` | 否 | 否 | - |
| `no_header` | `CephBool` | 否 | 否 | - |
| `period` | `CephFloat` | 否 | 否 | - |
| `session_id` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
