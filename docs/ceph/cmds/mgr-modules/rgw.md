# Rgw

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph rgw admin`

- 组件/模块：`mgr module` / `rgw`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/rgw/module.py`
- 标志：`mgr`
- 含义：rgw admin

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `params` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rgw realm bootstrap`

- 组件/模块：`mgr module` / `rgw`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/rgw/module.py`
- 标志：`mgr`
- 含义：Bootstrap new rgw realm, zonegroup, and zone

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `realm_name` | `CephString` | 否 | 是 | - |
| `zonegroup_name` | `CephString` | 否 | 是 | - |
| `zone_name` | `CephString` | 否 | 是 | - |
| `port` | `CephInt` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `zone_endpoints` | `CephString` | 否 | 是 | - |
| `start_radosgw` | `CephBool` | 否 | 否 | - |
| `skip_realm_components` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

### 版本差异

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `realm_name` | `CephString` | 否 | 是 | - |
| `zonegroup_name` | `CephString` | 否 | 是 | - |
| `zone_name` | `CephString` | 否 | 是 | - |
| `port` | `CephInt` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `zone_endpoints` | `CephString` | 否 | 是 | - |
| `start_radosgw` | `CephBool` | 否 | 否 | - |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `realm_name` | `CephString` | 否 | 是 | - |
| `zonegroup_name` | `CephString` | 否 | 是 | - |
| `zone_name` | `CephString` | 否 | 是 | - |
| `port` | `CephInt` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `zone_endpoints` | `CephString` | 否 | 是 | - |
| `start_radosgw` | `CephBool` | 否 | 否 | - |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `realm_name` | `CephString` | 否 | 是 | - |
| `zonegroup_name` | `CephString` | 否 | 是 | - |
| `zone_name` | `CephString` | 否 | 是 | - |
| `port` | `CephInt` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `zone_endpoints` | `CephString` | 否 | 是 | - |
| `start_radosgw` | `CephBool` | 否 | 否 | - |

## `ceph rgw realm reconcile`

- 组件/模块：`mgr module` / `rgw`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/rgw/module.py`
- 标志：`mgr`
- 含义：Bootstrap new rgw zone that syncs with existing zone

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `realm_name` | `CephString` | 否 | 是 | - |
| `zonegroup_name` | `CephString` | 否 | 是 | - |
| `zone_name` | `CephString` | 否 | 是 | - |
| `update` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rgw realm tokens`

- 组件/模块：`mgr module` / `rgw`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/rgw/module.py`
- 标志：`mgr`
- 含义：-

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rgw realm zone-creds create`

- 组件/模块：`mgr module` / `rgw`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/rgw/module.py`
- 标志：`mgr`
- 含义：Create credentials for new zone creation

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `realm_name` | `CephString` | 否 | 是 | - |
| `endpoints` | `CephString` | 否 | 是 | - |
| `sys_uid` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rgw realm zone-creds remove`

- 组件/模块：`mgr module` / `rgw`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/rgw/module.py`
- 标志：`mgr`
- 含义：Create credentials for new zone creation

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `realm_token` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rgw zone create`

- 组件/模块：`mgr module` / `rgw`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/rgw/module.py`
- 标志：`mgr`
- 含义：Bootstrap new rgw zone that syncs with zone on another cluster in the same realm

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `zone_name` | `CephString` | 否 | 是 | - |
| `realm_token` | `CephString` | 否 | 是 | - |
| `port` | `CephInt` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `start_radosgw` | `CephBool` | 否 | 否 | - |
| `zone_endpoints` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rgw zone modify`

- 组件/模块：`mgr module` / `rgw`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/rgw/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `realm_name` | `CephString` | 是 | 是 | - |
| `zonegroup_name` | `CephString` | 是 | 是 | - |
| `zone_name` | `CephString` | 是 | 是 | - |
| `realm_token` | `CephString` | 是 | 是 | - |
| `zone_endpoints` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph rgw zonegroup modify`

- 组件/模块：`mgr module` / `rgw`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/rgw/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `realm_name` | `CephString` | 是 | 是 | - |
| `zonegroup_name` | `CephString` | 是 | 是 | - |
| `zone_name` | `CephString` | 是 | 是 | - |
| `hostnames` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
