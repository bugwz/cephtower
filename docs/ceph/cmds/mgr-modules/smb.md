# Smb

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph smb apply`

- 组件/模块：`mgr module` / `smb`
- 支持版本：20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/smb/module.py`
- 标志：`mgr`
- 含义：Create, update, or remove smb configuration resources based on YAML or JSON specs

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `password_filter` | `CephString` | 否 | 是 | - |
| `password_filter_out` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph smb cluster create`

- 组件/模块：`mgr module` / `smb`
- 支持版本：20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/smb/module.py`
- 标志：`mgr`
- 含义：Create an smb cluster

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `auth_mode` | `CephString` | 是 | 是 | - |
| `domain_realm` | `CephString` | 否 | 是 | - |
| `domain_join_ref` | `CephString` | 否 | 是 | - |
| `domain_join_user_pass` | `CephString` | 否 | 是 | - |
| `user_group_ref` | `CephString` | 否 | 是 | - |
| `define_user_pass` | `CephString` | 否 | 是 | - |
| `custom_dns` | `CephString` | 否 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `clustering` | `CephString` | 否 | 是 | - |
| `public_addrs` | `CephString` | 否 | 是 | - |
| `password_filter` | `CephString` | 否 | 是 | - |
| `password_filter_out` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph smb cluster ls`

- 组件/模块：`mgr module` / `smb`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/smb/module.py`
- 标志：`mgr`
- 含义：List smb clusters by ID

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph smb cluster rm`

- 组件/模块：`mgr module` / `smb`
- 支持版本：20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/smb/module.py`
- 标志：`mgr`
- 含义：Remove an smb cluster

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `password_filter` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph smb share create`

- 组件/模块：`mgr module` / `smb`
- 支持版本：20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/smb/module.py`
- 标志：`mgr`
- 含义：Create an smb share

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `share_id` | `CephString` | 是 | 是 | - |
| `cephfs_volume` | `CephString` | 是 | 是 | - |
| `path` | `CephString` | 是 | 是 | - |
| `share_name` | `CephString` | 否 | 是 | - |
| `subvolume` | `CephString` | 否 | 是 | - |
| `readonly` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph smb share ls`

- 组件/模块：`mgr module` / `smb`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/smb/module.py`
- 标志：`mgr`
- 含义：List smb shares in a cluster by ID

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph smb share rm`

- 组件/模块：`mgr module` / `smb`
- 支持版本：20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/smb/module.py`
- 标志：`mgr`
- 含义：Remove an smb share

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `share_id` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph smb show`

- 组件/模块：`mgr module` / `smb`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/smb/module.py`
- 标志：`mgr`
- 含义：Show resources fetched from the local config store based on resource type or resource type and id(s).

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `resource_names` | `CephString` | 否 | 是 | - |
| `results` | `CephString` | 否 | 是 | - |
| `password_filter` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
