# Rbd

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph rbd mirror snapshot schedule add`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Add rbd mirror snapshot schedule

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `level_spec` | `CephString` | 是 | 是 | - |
| `interval` | `CephString` | 是 | 是 | - |
| `start_time` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd mirror snapshot schedule list`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：List rbd mirror snapshot schedule

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `level_spec` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd mirror snapshot schedule remove`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Remove rbd mirror snapshot schedule

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `level_spec` | `CephString` | 是 | 是 | - |
| `interval` | `CephString` | 否 | 是 | - |
| `start_time` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd mirror snapshot schedule status`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Show rbd mirror snapshot schedule status

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `level_spec` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd perf image counters`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Retrieve current RBD IO performance counters

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool_spec` | `CephString` | 否 | 是 | - |
| `sort_by` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

### 版本差异

- `16.2.15` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool_spec` | `CephString` | 否 | 是 | - |
| `sort_by` | `CephChoices` | 否 | 是 | 可选值: write_ops, write_bytes, write_latency, read_ops, read_bytes, read_latency |

## `ceph rbd perf image stats`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Retrieve current RBD IO performance stats

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool_spec` | `CephString` | 否 | 是 | - |
| `sort_by` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

### 版本差异

- `16.2.15` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool_spec` | `CephString` | 否 | 是 | - |
| `sort_by` | `CephChoices` | 否 | 是 | 可选值: write_ops, write_bytes, write_latency, read_ops, read_bytes, read_latency |

## `ceph rbd task add flatten`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Flatten a cloned image asynchronously in the background

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `image_spec` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd task add migration abort`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Abort a prepared migration asynchronously in the background

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `image_spec` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd task add migration commit`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Commit an executed migration asynchronously in the background

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `image_spec` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd task add migration execute`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Execute an image migration asynchronously in the background

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `image_spec` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd task add remove`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Remove an image asynchronously in the background

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `image_spec` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd task add trash remove`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Remove an image from the trash asynchronously in the background

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `image_id_spec` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd task cancel`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Cancel a pending or running asynchronous task

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `task_id` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd task list`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：List pending or running asynchronous tasks

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `task_id` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd trash purge schedule add`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Add rbd trash purge schedule

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `level_spec` | `CephString` | 是 | 是 | - |
| `interval` | `CephString` | 是 | 是 | - |
| `start_time` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd trash purge schedule list`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：List rbd trash purge schedule

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `level_spec` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd trash purge schedule remove`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Remove rbd trash purge schedule

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `level_spec` | `CephString` | 是 | 是 | - |
| `interval` | `CephString` | 否 | 是 | - |
| `start_time` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rbd trash purge schedule status`

- 组件/模块：`mgr module` / `rbd_support`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/rbd_support/module.py`
- 标志：`mgr`
- 含义：Show rbd trash purge schedule status

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `level_spec` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
