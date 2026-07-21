# Nfs

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph nfs cluster config get`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Fetch NFS-Ganesha config

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

## `ceph nfs cluster config reset`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Reset NFS-Ganesha Config to default

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

## `ceph nfs cluster config set`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Set NFS-Ganesha config by `-i <config_file>`

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

## `ceph nfs cluster create`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Create an NFS Cluster

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `ingress` | `CephBool` | 否 | 否 | - |
| `virtual_ip` | `CephString` | 否 | 否 | - |
| `ingress_mode` | `CephString` | 否 | 否 | - |
| `port` | `CephInt` | 否 | 否 | - |

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
| `cluster_id` | `CephString` | 是 | 是 | - |
| `placement` | `CephString` | 否 | 是 | - |
| `ingress` | `CephBool` | 否 | 否 | - |
| `virtual_ip` | `CephString` | 否 | 否 | - |
| `port` | `CephInt` | 否 | 否 | - |

## `ceph nfs cluster delete`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Removes an NFS Cluster (DEPRECATED)

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

## `ceph nfs cluster info`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Displays NFS Cluster info

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nfs cluster ls`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：List NFS Clusters

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nfs cluster rm`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Removes an NFS Cluster

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

## `ceph nfs export apply`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Create or update an export by `-i <json_or_ganesha_export_file>`

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

## `ceph nfs export create cephfs`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Create a CephFS export

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `pseudo_path` | `CephString` | 是 | 是 | - |
| `fsname` | `CephString` | 是 | 是 | - |
| `path` | `CephString` | 否 | 是 | - |
| `readonly` | `CephBool` | 否 | 否 | - |
| `client_addr` | `CephString` | 否 | 否 | - |
| `squash` | `CephString` | 否 | 否 | - |
| `sectype` | `CephString` | 否 | 否 | - |
| `cmount_path` | `CephString` | 否 | 否 | - |

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
| `cluster_id` | `CephString` | 是 | 是 | - |
| `pseudo_path` | `CephString` | 是 | 是 | - |
| `fsname` | `CephString` | 是 | 是 | - |
| `path` | `CephString` | 否 | 是 | - |
| `readonly` | `CephBool` | 否 | 否 | - |
| `client_addr` | `CephString` | 否 | 否 | - |
| `squash` | `CephString` | 否 | 否 | - |
| `sectype` | `CephString` | 否 | 否 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `pseudo_path` | `CephString` | 是 | 是 | - |
| `fsname` | `CephString` | 是 | 是 | - |
| `path` | `CephString` | 否 | 是 | - |
| `readonly` | `CephBool` | 否 | 否 | - |
| `client_addr` | `CephString` | 否 | 否 | - |
| `squash` | `CephString` | 否 | 否 | - |
| `sectype` | `CephString` | 否 | 否 | - |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `pseudo_path` | `CephString` | 是 | 是 | - |
| `fsname` | `CephString` | 是 | 是 | - |
| `path` | `CephString` | 否 | 是 | - |
| `readonly` | `CephBool` | 否 | 否 | - |
| `client_addr` | `CephString` | 否 | 否 | - |
| `squash` | `CephString` | 否 | 否 | - |
| `sectype` | `CephString` | 否 | 否 | - |

## `ceph nfs export create rgw`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Create an RGW export

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `pseudo_path` | `CephString` | 是 | 是 | - |
| `bucket` | `CephString` | 否 | 是 | - |
| `user_id` | `CephString` | 否 | 是 | - |
| `readonly` | `CephBool` | 否 | 否 | - |
| `client_addr` | `CephString` | 否 | 否 | - |
| `squash` | `CephString` | 否 | 否 | - |
| `sectype` | `CephString` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nfs export delete`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Delete a cephfs export (DEPRECATED)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `pseudo_path` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nfs export get`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Fetch a export of a NFS cluster given the pseudo path/binding (DEPRECATED)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `pseudo_path` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nfs export info`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Fetch a export of a NFS cluster given the pseudo path/binding

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `pseudo_path` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nfs export ls`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：List exports of a NFS cluster

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `detailed` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph nfs export rm`

- 组件/模块：`mgr module` / `nfs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/nfs/module.py`
- 标志：`mgr`
- 含义：Remove a cephfs export

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `cluster_id` | `CephString` | 是 | 是 | - |
| `pseudo_path` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
