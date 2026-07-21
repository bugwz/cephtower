# Telemetry

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph telemetry channel ls`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：List all channels

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry collection ls`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：List all collections

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry diff`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Show the diff between opted-in collection and available collection

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry disable channel`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Disable a list of channels

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `channels` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry disable channel all`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Disable all channels

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `channels` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry enable channel`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Enable a list of channels

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `channels` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry enable channel all`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Enable all channels

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `channels` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry off`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Disable telemetry reports from this cluster

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry on`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Enable telemetry reports from this cluster

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `license` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry preview`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Preview a sample report of the most recent collections available (except for 'device')

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `channels` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry preview-all`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Preview a sample report of the most recent collections available of all channels (including 'device')

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry preview-device`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Preview a sample device report of the most recent device collection

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry send`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Force sending data to Ceph telemetry

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `endpoint` | `CephString` | 否 | 是 | - |
| `license` | `CephString` | 否 | 是 | - |

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
| `endpoint` | `CephChoices` | 否 | 是 | 可重复; 可选值: ceph, device |
| `license` | `CephString` | 否 | 是 | - |

## `ceph telemetry show`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Show last report or report to be sent

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `channels` | `CephString` | 否 | 是 | - |

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
| `channels` | `CephString` | 是 | 是 | 可重复 |

## `ceph telemetry show-all`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Show report of all channels

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry show-device`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Show last device report or device report to be sent

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph telemetry status`

- 组件/模块：`mgr module` / `telemetry`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/telemetry/module.py`
- 标志：`mgr`
- 含义：Show current configuration

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
