# Device

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph device check-health`

- 组件/模块：`mgr module` / `devicehealth`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/devicehealth/module.py`
- 标志：`mgr`
- 含义：Check life expectancy of devices

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device get-health-metrics`

- 组件/模块：`mgr module` / `devicehealth`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/devicehealth/module.py`
- 标志：`mgr`
- 含义：Show stored device metrics for the device

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `devid` | `CephString` | 是 | 是 | - |
| `sample` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device info`

- 组件/模块：`ceph mgr` / `mgr`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：Show information about a device

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `devid` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device light`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Enable or disable the device light. Default type is `ident` 'Usage: device light (on|off) <devid> [ident|fault] [--force]'

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `enable` | `CephString` | 是 | 是 | - |
| `devid` | `CephString` | 是 | 是 | - |
| `light_type` | `CephString` | 否 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device ls`

- 组件/模块：`ceph mgr` / `mgr`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：Show devices

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device ls-by-daemon`

- 组件/模块：`ceph mgr` / `mgr`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：Show devices associated with a daemon

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `who` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device ls-by-host`

- 组件/模块：`ceph mgr` / `mgr`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：Show devices on a host

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `host` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device ls-lights`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：List currently active device indicator lights

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device monitoring off`

- 组件/模块：`mgr module` / `devicehealth`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/devicehealth/module.py`
- 标志：`mgr`
- 含义：Disable device health monitoring

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device monitoring on`

- 组件/模块：`mgr module` / `devicehealth`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/devicehealth/module.py`
- 标志：`mgr`
- 含义：Enable device health monitoring

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device predict-life-expectancy`

- 组件/模块：`mgr module` / `devicehealth`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/devicehealth/module.py`
- 标志：`mgr`
- 含义：Predict life expectancy with local predictor

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `devid` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device query-daemon-health-metrics`

- 组件/模块：`mgr module` / `devicehealth`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/devicehealth/module.py`
- 标志：`mgr`
- 含义：Get device health metrics for a given daemon

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `who` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device rm-life-expectancy`

- 组件/模块：`ceph mgr` / `mgr`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：Clear predicted device life expectancy

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `devid` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device scrape-daemon-health-metrics`

- 组件/模块：`mgr module` / `devicehealth`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/devicehealth/module.py`
- 标志：`mgr`
- 含义：Scrape and store device health metrics for a given daemon

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `who` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device scrape-health-metrics`

- 组件/模块：`mgr module` / `devicehealth`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/devicehealth/module.py`
- 标志：`mgr`
- 含义：Scrape and store device health metrics

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `devid` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph device set-life-expectancy`

- 组件/模块：`ceph mgr` / `mgr`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：Set predicted device life expectancy

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `devid` | `CephString` | 是 | 是 | - |
| `from` | `CephString` | 是 | 是 | - |
| `to` | `CephString` | 否 | 是 | - |

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
| `devid` | `CephString` | 是 | 是 | - |
| `from` | `CephString` | 是 | 是 | - |
| `to` | `CephString` | 是 | 是 | - |
