# Orchestrator

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph orch`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Start, stop, restart, redeploy, or reconfig an entire service (i.e. all daemons)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `action` | `CephString` | 是 | 是 | - |
| `service_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch alertmanager get-credentials`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
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

## `ceph orch alertmanager set-credentials`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `username` | `CephString` | 否 | 是 | - |
| `password` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Update the size or placement for a service or apply a large yaml spec

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `service_type` | `CephString` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |
| `continue_on_error` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply iscsi`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Scale an iSCSI service

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `api_user` | `CephString` | 是 | 是 | - |
| `api_password` | `CephString` | 是 | 是 | - |
| `trusted_ip_list` | `CephString` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply jaeger`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Apply jaeger tracing services

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `es_nodes` | `CephString` | 否 | 是 | - |
| `without_query` | `CephBool` | 否 | 否 | - |
| `placement` | `CephString` | 否 | 否 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply mds`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Update the number of MDS instances for the given fs_name

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply mgmt-gateway`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Add a cluster gateway service (cephadm only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `port` | `CephInt` | 否 | 是 | - |
| `ssl` | `CephBool` | 否 | 否 | - |
| `enable_auth` | `CephBool` | 否 | 否 | - |
| `virtual_ip` | `CephString` | 否 | 否 | - |
| `placement` | `CephString` | 否 | 否 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply nfs`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Scale an NFS service

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `svc_id` | `CephString` | 是 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `format` | `CephString` | 否 | 否 | - |
| `port` | `CephInt` | 否 | 否 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply nvmeof`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Scale an nvmeof service

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `group` | `CephString` | 是 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply oauth2-proxy`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Add a cluster gateway service (cephadm only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `https_address` | `CephString` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply osd`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Create OSD daemon(s) on all available devices

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `all_available_devices` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |
| `method` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply rgw`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Update the number of RGW instances for the given zone

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `svc_id` | `CephString` | 是 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `realm` | `CephString` | 否 | 否 | - |
| `zonegroup` | `CephString` | 否 | 否 | - |
| `zone` | `CephString` | 否 | 否 | - |
| `networks` | `CephString` | 否 | 否 | - |
| `port` | `CephInt` | 否 | 否 | - |
| `ssl` | `CephBool` | 否 | 否 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply smb`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Apply an SMB network file system gateway service configuration.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `config_uri` | `CephString` | 是 | 是 | - |
| `features` | `CephString` | 否 | 是 | - |
| `join_sources` | `CephString` | 否 | 是 | - |
| `custom_dns` | `CephString` | 否 | 是 | - |
| `include_ceph_users` | `CephString` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch apply snmp-gateway`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Add a Prometheus to SNMP gateway service (cephadm only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `snmp_version` | `CephString` | 是 | 是 | - |
| `destination` | `CephString` | 是 | 是 | - |
| `port` | `CephInt` | 否 | 是 | - |
| `engine_id` | `CephString` | 否 | 是 | - |
| `auth_protocol` | `CephString` | 否 | 是 | - |
| `privacy_protocol` | `CephString` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `unmanaged` | `CephBool` | 否 | 否 | - |
| `dry_run` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch cancel`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Cancel ongoing background operations

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr cert check`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr cert get`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cert_name` | `CephString` | 是 | 是 | - |
| `service_name` | `CephString` | 否 | 否 | - |
| `hostname` | `CephString` | 否 | 否 | - |
| `no_exception_when_missing` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr cert ls`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `show_details` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr cert rm`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cert_name` | `CephString` | 是 | 是 | - |
| `service_name` | `CephString` | 否 | 否 | - |
| `hostname` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr cert set`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Sets the cert from the --cert argument or from -i <cert-file>.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cert_name` | `CephString` | 是 | 是 | - |
| `cert` | `CephString` | 否 | 否 | - |
| `service_name` | `CephString` | 否 | 否 | - |
| `hostname` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr cert-key set`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Sets the cert-key pair from -i <pem-file>, which must be a valid PEM file containing both the certificate and the private key.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `entity` | `CephString` | 是 | 是 | - |
| `cert` | `CephString` | 否 | 否 | - |
| `key` | `CephString` | 否 | 否 | - |
| `cert_name` | `CephString` | 否 | 否 | - |
| `service_name` | `CephString` | 否 | 否 | - |
| `hostname` | `CephString` | 否 | 否 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr entity ls`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr generate-certificates`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `module_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr key get`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `key_name` | `CephString` | 是 | 是 | - |
| `service_name` | `CephString` | 否 | 否 | - |
| `hostname` | `CephString` | 否 | 否 | - |
| `no_exception_when_missing` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr key ls`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr key rm`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `key_name` | `CephString` | 是 | 是 | - |
| `service_name` | `CephString` | 否 | 否 | - |
| `hostname` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr key set`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Sets the key from the --key argument or from -i <key-file>.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `key_name` | `CephString` | 是 | 是 | - |
| `key` | `CephString` | 否 | 否 | - |
| `service_name` | `CephString` | 否 | 否 | - |
| `hostname` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch certmgr reload`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch client-keyring ls`

- 组件/模块：`mgr module` / `cephadm`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/cephadm/module.py`
- 标志：`mgr`
- 含义：List client keyrings under cephadm management

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch client-keyring rm`

- 组件/模块：`mgr module` / `cephadm`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/cephadm/module.py`
- 标志：`mgr`
- 含义：Remove client keyring from cephadm management

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `entity` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch client-keyring set`

- 组件/模块：`mgr module` / `cephadm`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/cephadm/module.py`
- 标志：`mgr`
- 含义：Add or update client keyring under cephadm management

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `entity` | `CephString` | 是 | 是 | - |
| `placement` | `CephString` | 是 | 是 | - |
| `owner` | `CephString` | 否 | 是 | - |
| `mode` | `CephString` | 否 | 是 | - |
| `no_ceph_conf` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch daemon`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Start, stop, restart, redeploy, reconfig, or rotate-key for a specific daemon

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `action` | `CephString` | 是 | 是 | - |
| `name` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch daemon add`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Add daemon(s)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `daemon_type` | `CephString` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch daemon add iscsi`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Start iscsi daemon(s)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `api_user` | `CephString` | 是 | 是 | - |
| `api_password` | `CephString` | 是 | 是 | - |
| `trusted_ip_list` | `CephString` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch daemon add mds`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Start MDS daemon(s)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch daemon add nfs`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Start NFS daemon(s)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `svc_id` | `CephString` | 是 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch daemon add nvmeof`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Start nvmeof daemon(s)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephString` | 是 | 是 | - |
| `group` | `CephString` | 是 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch daemon add osd`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Create OSD daemon(s) on specified host and device(s) (e.g., ceph orch daemon add osd myhost:/dev/sdb)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `svc_arg` | `CephString` | 否 | 是 | - |
| `method` | `CephString` | 否 | 是 | - |
| `skip_validation` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch daemon add rgw`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Start RGW daemon(s)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `svc_id` | `CephString` | 是 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `port` | `CephInt` | 否 | 否 | - |
| `ssl` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch daemon redeploy`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Redeploy a daemon (with a specific image)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | - |
| `image` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch daemon rm`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Remove specific daemon(s)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `names` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch device ls`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：List devices on a host

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 否 | 是 | - |
| `format` | `CephString` | 否 | 否 | - |
| `refresh` | `CephBool` | 否 | 否 | - |
| `wide` | `CephBool` | 否 | 否 | - |
| `summary` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch device replace`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Perform all required operations in order to replace a device.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `device` | `CephString` | 是 | 是 | - |
| `clear` | `CephBool` | 否 | 否 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch device zap`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Zap (erase!) a device so it can be re-used

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `path` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch get-security-config`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
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

## `ceph orch hardware light`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Enable or Disable a device or chassis LED

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `light_type` | `CephString` | 是 | 是 | - |
| `action` | `CephString` | 是 | 是 | - |
| `hostname` | `CephString` | 是 | 是 | - |
| `device` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch hardware powercycle`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Reboot a host

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch hardware shutdown`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Shutdown a host

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch hardware status`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Display hardware status summary :param hostname: hostname

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 否 | 是 | - |
| `category` | `CephString` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host add`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Add a host

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `addr` | `CephString` | 否 | 是 | - |
| `labels` | `CephString` | 否 | 是 | - |
| `maintenance` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host drain`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：drain all daemons from a host

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `keep_conf_keyring` | `CephBool` | 否 | 否 | - |
| `zap_osd_devices` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host drain stop`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：drain all daemons from a host

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host label add`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Add a host label

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `label` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host label rm`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Remove a host label

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `label` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host ls`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：List high level host information

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `format` | `CephString` | 否 | 否 | - |
| `host_pattern` | `CephString` | 否 | 否 | - |
| `label` | `CephString` | 否 | 否 | - |
| `host_status` | `CephString` | 否 | 否 | - |
| `detail` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host maintenance enter`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Prepare a host for maintenance by shutting down and disabling all Ceph daemons (cephadm only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host maintenance exit`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Return a host from maintenance, restarting all Ceph daemons (cephadm only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `offline` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host ok-to-stop`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Check if the specified host can be safely stopped without reducing availability

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host rescan`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Perform a disk rescan on a host

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `with_summary` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host rm`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Remove a host

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `offline` | `CephBool` | 否 | 否 | - |
| `rm_crush_entry` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch host set-addr`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Update a host address

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 是 | 是 | - |
| `addr` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch ls`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：List services known to orchestrator

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `service_type` | `CephString` | 否 | 是 | - |
| `service_name` | `CephString` | 否 | 是 | - |
| `export` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `refresh` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch osd rm`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Remove OSD daemons

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `osd_id` | `CephString` | 是 | 是 | - |
| `replace` | `CephBool` | 否 | 否 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `zap` | `CephBool` | 否 | 否 | - |
| `no_destroy` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch osd rm status`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Status of OSD removal operation

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch osd rm stop`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Cancel ongoing OSD removal operation

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `osd_id` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch osd set-spec-affinity`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Set service spec affinity for osd

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `service_name` | `CephString` | 是 | 是 | - |
| `osd_id` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch pause`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Pause orchestrator background work

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch prometheus get-credentials`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
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

## `ceph orch prometheus remove-target`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `url` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch prometheus set-credentials`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `username` | `CephString` | 否 | 是 | - |
| `password` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch prometheus set-custom-alerts`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
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

## `ceph orch prometheus set-target`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `url` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch ps`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：List daemons known to orchestrator

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `hostname` | `CephString` | 否 | 是 | - |
| `service_name` | `CephString` | 否 | 否 | - |
| `daemon_type` | `CephString` | 否 | 否 | - |
| `daemon_id` | `CephString` | 否 | 否 | - |
| `sort_by` | `CephString` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |
| `refresh` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch resume`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Resume orchestrator background work (if paused)

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch rm`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Remove a service

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `service_name` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch set backend`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Select orchestrator module backend

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `module_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch set-managed`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Set 'unmanaged: false' for the given service name

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `service_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch set-unmanaged`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Set 'unmanaged: true' for the given service name

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `service_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch status`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Report configured backend and its status

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `detail` | `CephBool` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch tuned-profile add-setting`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `profile_name` | `CephString` | 是 | 是 | - |
| `setting` | `CephString` | 是 | 是 | - |
| `value` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch tuned-profile add-settings`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `profile_name` | `CephString` | 是 | 是 | - |
| `settings` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch tuned-profile apply`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Add or update a tuned profile

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `profile_name` | `CephString` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `settings` | `CephString` | 否 | 是 | - |
| `no_overwrite` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch tuned-profile ls`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch tuned-profile rm`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `profile_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch tuned-profile rm-setting`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `profile_name` | `CephString` | 是 | 是 | - |
| `setting` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch tuned-profile rm-settings`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：-

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `profile_name` | `CephString` | 是 | 是 | - |
| `settings` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch unpause`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Alias to orch resume

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch update service`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Update image for non-ceph image daemon

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `service_type` | `CephString` | 是 | 是 | - |
| `image` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch upgrade check`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Check service versions vs available and target containers

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `image` | `CephString` | 否 | 是 | - |
| `ceph_version` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch upgrade ls`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Check for available versions (or tags) we can upgrade to

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `image` | `CephString` | 否 | 是 | - |
| `tags` | `CephBool` | 否 | 否 | - |
| `show_all_versions` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch upgrade pause`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Pause an in-progress upgrade

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch upgrade resume`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Resume paused upgrade

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch upgrade start`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Initiate upgrade

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `image` | `CephString` | 否 | 是 | - |
| `daemon_types` | `CephString` | 否 | 否 | - |
| `hosts` | `CephString` | 否 | 否 | - |
| `services` | `CephString` | 否 | 否 | - |
| `limit` | `CephInt` | 否 | 否 | - |
| `ceph_version` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch upgrade status`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Check the status of any potential ongoing upgrade operation

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `format` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph orch upgrade stop`

- 组件/模块：`mgr module` / `orchestrator`
- 支持版本：20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/orchestrator/module.py`
- 标志：`mgr`
- 含义：Stop an in-progress upgrade

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
