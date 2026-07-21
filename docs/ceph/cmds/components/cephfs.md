# Cephfs

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph fs add_data_pool`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：add data pool <pool>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `pool` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs authorize`

- 组件/模块：`ceph mon` / `auth`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rwx`
- 来源：`src/mon/MonCommands.h`
- 含义：add auth for <entity> to access file system <filesystem> based on following directory and permissions pairs

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `filesystem` | `CephString` | 是 | 是 | - |
| `entity` | `CephString` | 是 | 是 | - |
| `caps` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs clone cancel`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Cancel an pending or ongoing clone operation.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `clone_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs clone status`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get status on a cloned subvolume.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `clone_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs compat`

- 组件/模块：`ceph mon` / `fs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：manipulate compat settings

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `subop` | `CephChoices` | 是 | 是 | 可选值: rm_compat, rm_incompat, add_compat, add_incompat |
| `feature` | `CephInt` | 是 | 是 | - |
| `feature_str` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs compat show`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：show fs compatibility settings

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs dump`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：dump all CephFS status, optionally from epoch

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `epoch` | `CephInt` | 否 | 是 | 范围: 0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs fail`

- 组件/模块：`ceph mon` / `fs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：bring the file system down and all of its ranks

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

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
| `fs_name` | `CephString` | 是 | 是 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

## `ceph fs feature ls`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：list available cephfs features to be set/unset

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs flag set`

- 组件/模块：`ceph mon` / `fs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：Set a global CephFS flag

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `flag_name` | `CephChoices` | 是 | 是 | 可选值: enable_multiple |
| `val` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs get`

- 组件/模块：`ceph mon` / `fs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：get info about one filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs ls`

- 组件/模块：`ceph mon` / `fs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：list filesystems

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs lsflags`

- 组件/模块：`ceph mon` / `fs`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：list the flags set on a ceph filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs mirror disable`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：disable mirroring for a ceph filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs mirror enable`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：enable mirroring for a ceph filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs mirror peer_add`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：add a mirror peer for a ceph filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `uuid` | `CephString` | 是 | 是 | - |
| `remote_cluster_spec` | `CephString` | 是 | 是 | - |
| `remote_fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs mirror peer_remove`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove a mirror peer for a ceph filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `uuid` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs new`

- 组件/模块：`ceph mon` / `fs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：make new filesystem using named pools <metadata> and <data>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `metadata` | `CephString` | 是 | 是 | - |
| `data` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `allow_dangerous_metadata_overlay` | `CephBool` | 否 | 否 | - |
| `fscid` | `CephInt` | 否 | 否 | 范围: 0 |
| `recover` | `CephBool` | 否 | 否 | - |
| `yes_i_really_really_mean_it` | `CephBool` | 否 | 否 | - |
| `set` | `CephString` | 否 | 否 | 可重复 |

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
| `fs_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `metadata` | `CephString` | 是 | 是 | - |
| `data` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `allow_dangerous_metadata_overlay` | `CephBool` | 否 | 否 | - |
| `fscid` | `CephInt` | 否 | 否 | 范围: 0 |
| `recover` | `CephBool` | 否 | 否 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `metadata` | `CephString` | 是 | 是 | - |
| `data` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `allow_dangerous_metadata_overlay` | `CephBool` | 否 | 否 | - |
| `fscid` | `CephInt` | 否 | 否 | 范围: 0 |
| `recover` | `CephBool` | 否 | 否 | - |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `metadata` | `CephString` | 是 | 是 | - |
| `data` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `allow_dangerous_metadata_overlay` | `CephBool` | 否 | 否 | - |
| `fscid` | `CephInt` | 否 | 否 | 范围: 0 |
| `recover` | `CephBool` | 否 | 否 | - |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `metadata` | `CephString` | 是 | 是 | - |
| `data` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `allow_dangerous_metadata_overlay` | `CephBool` | 否 | 否 | - |
| `fscid` | `CephInt` | 否 | 否 | 范围: 0 |
| `recover` | `CephBool` | 否 | 否 | - |

## `ceph fs perf stats`

- 组件/模块：`mgr module` / `stats`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/stats/module.py`
- 标志：`mgr`
- 含义：retrieve ceph fs performance stats

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `mds_rank` | `CephString` | 否 | 是 | - |
| `client_id` | `CephString` | 否 | 是 | - |
| `client_ip` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs quiesce`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Manage quiesce sets of subvolumes

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `members` | `CephString` | 否 | 是 | 可重复 |
| `set_id` | `CephString` | 否 | 否 | - |
| `timeout` | `CephFloat` | 否 | 否 | 范围: 0 |
| `expiration` | `CephFloat` | 否 | 否 | 范围: 0 |
| `await_for` | `CephFloat` | 否 | 否 | 范围: 0 |
| `await` | `CephBool` | 否 | 否 | - |
| `if_version` | `CephInt` | 否 | 否 | 范围: 0 |
| `include` | `CephBool` | 否 | 否 | - |
| `exclude` | `CephBool` | 否 | 否 | - |
| `reset` | `CephBool` | 否 | 否 | - |
| `release` | `CephBool` | 否 | 否 | - |
| `query` | `CephBool` | 否 | 否 | - |
| `all` | `CephBool` | 否 | 否 | - |
| `cancel` | `CephBool` | 否 | 否 | - |
| `group_name` | `CephString` | 否 | 否 | - |
| `leader` | `CephBool` | 否 | 否 | - |
| `with_leader` | `CephInt` | 否 | 否 | 范围: 0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs rename`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：rename a ceph file system

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `new_fs_name` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

### 版本差异

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `new_fs_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `new_fs_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

## `ceph fs required_client_features`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：add/remove required features of clients

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `subop` | `CephChoices` | 是 | 是 | 可选值: add, rm |
| `val` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs reset`

- 组件/模块：`ceph mon` / `fs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：disaster recovery only: reset to a single-MDS map

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs rm`

- 组件/模块：`ceph mon` / `fs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：disable the named filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs rm_data_pool`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove data pool <pool>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `pool` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs set`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set fs parameter <var> to <val>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: max_mds, allow_dirfrags, allow_new_snaps, allow_standby_replay, bal_rank_mask, balance_automate, balancer, cluster_down, down, inline_data, joinable, max_file_size, max_xattr_size, min_compat_client, refuse_client_session, refuse_standby_for_another_fs, session_autoclose, session_timeout, standby_count_wanted |
| `val` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |
| `yes_i_really_really_mean_it` | `CephBool` | 否 | 否 | - |

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
| `fs_name` | `CephString` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: max_mds, max_file_size, allow_new_snaps, inline_data, cluster_down, allow_dirfrags, balancer, standby_count_wanted, session_timeout, session_autoclose, allow_standby_replay, down, joinable, min_compat_client |
| `val` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |
| `yes_i_really_really_mean_it` | `CephBool` | 否 | 否 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: max_mds, max_file_size, allow_new_snaps, inline_data, cluster_down, allow_dirfrags, balancer, standby_count_wanted, session_timeout, session_autoclose, allow_standby_replay, down, joinable, min_compat_client |
| `val` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |
| `yes_i_really_really_mean_it` | `CephBool` | 否 | 否 | - |

## `ceph fs set-default`

- 组件/模块：`ceph mon` / `fs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set the default to the named filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs set_default`

- 组件/模块：`ceph mon` / `fs`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：set the default to the named filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snap-schedule activate`

- 组件/模块：`mgr module` / `snap_schedule`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/snap_schedule/module.py`
- 标志：`mgr`
- 含义：Activate a snapshot schedule for <path>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `repeat` | `CephString` | 否 | 是 | - |
| `start` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |
| `subvol` | `CephString` | 否 | 是 | - |
| `group` | `CephString` | 否 | 是 | - |

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
| `path` | `CephString` | 是 | 是 | - |
| `repeat` | `CephString` | 否 | 是 | - |
| `start` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |

## `ceph fs snap-schedule add`

- 组件/模块：`mgr module` / `snap_schedule`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/snap_schedule/module.py`
- 标志：`mgr`
- 含义：Set a snapshot schedule for <path>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `snap_schedule` | `CephString` | 是 | 是 | - |
| `start` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |
| `subvol` | `CephString` | 否 | 是 | - |
| `group` | `CephString` | 否 | 是 | - |

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
| `path` | `CephString` | 是 | 是 | - |
| `snap_schedule` | `CephString` | 是 | 是 | - |
| `start` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |

## `ceph fs snap-schedule deactivate`

- 组件/模块：`mgr module` / `snap_schedule`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/snap_schedule/module.py`
- 标志：`mgr`
- 含义：Deactivate a snapshot schedule for <path>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `repeat` | `CephString` | 否 | 是 | - |
| `start` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |
| `subvol` | `CephString` | 否 | 是 | - |
| `group` | `CephString` | 否 | 是 | - |

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
| `path` | `CephString` | 是 | 是 | - |
| `repeat` | `CephString` | 否 | 是 | - |
| `start` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |

## `ceph fs snap-schedule list`

- 组件/模块：`mgr module` / `snap_schedule`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/snap_schedule/module.py`
- 标志：`mgr`
- 含义：Get current snapshot schedule for <path>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `recursive` | `CephBool` | 否 | 否 | - |
| `fs` | `CephString` | 否 | 否 | - |
| `subvol` | `CephString` | 否 | 否 | - |
| `group` | `CephString` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |

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
| `path` | `CephString` | 是 | 是 | - |
| `recursive` | `CephBool` | 否 | 否 | - |
| `fs` | `CephString` | 否 | 否 | - |
| `format` | `CephString` | 否 | 否 | - |

## `ceph fs snap-schedule remove`

- 组件/模块：`mgr module` / `snap_schedule`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/snap_schedule/module.py`
- 标志：`mgr`
- 含义：Remove a snapshot schedule for <path>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `repeat` | `CephString` | 否 | 是 | - |
| `start` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |
| `subvol` | `CephString` | 否 | 是 | - |
| `group` | `CephString` | 否 | 是 | - |

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
| `path` | `CephString` | 是 | 是 | - |
| `repeat` | `CephString` | 否 | 是 | - |
| `start` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |

## `ceph fs snap-schedule retention add`

- 组件/模块：`mgr module` / `snap_schedule`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/snap_schedule/module.py`
- 标志：`mgr`
- 含义：Set a retention specification for <path>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `retention_spec_or_period` | `CephString` | 是 | 是 | - |
| `retention_count` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |
| `subvol` | `CephString` | 否 | 是 | - |
| `group` | `CephString` | 否 | 是 | - |

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
| `path` | `CephString` | 是 | 是 | - |
| `retention_spec_or_period` | `CephString` | 是 | 是 | - |
| `retention_count` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |

## `ceph fs snap-schedule retention remove`

- 组件/模块：`mgr module` / `snap_schedule`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/snap_schedule/module.py`
- 标志：`mgr`
- 含义：Remove a retention specification for <path>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 是 | 是 | - |
| `retention_spec_or_period` | `CephString` | 是 | 是 | - |
| `retention_count` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |
| `subvol` | `CephString` | 否 | 是 | - |
| `group` | `CephString` | 否 | 是 | - |

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
| `path` | `CephString` | 是 | 是 | - |
| `retention_spec_or_period` | `CephString` | 是 | 是 | - |
| `retention_count` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |

## `ceph fs snap-schedule status`

- 组件/模块：`mgr module` / `snap_schedule`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/snap_schedule/module.py`
- 标志：`mgr`
- 含义：List current snapshot schedules

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `path` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |
| `subvol` | `CephString` | 否 | 是 | - |
| `group` | `CephString` | 否 | 是 | - |
| `format` | `CephString` | 否 | 否 | - |

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
| `path` | `CephString` | 否 | 是 | - |
| `fs` | `CephString` | 否 | 是 | - |
| `format` | `CephString` | 否 | 否 | - |

## `ceph fs snapshot mirror add`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：Add a directory for snapshot mirroring

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `path` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror daemon status`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：Get mirror daemon status

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror dirmap`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：Get current mirror instance map for a directory

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `path` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror disable`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：Disable snapshot mirroring for a filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror enable`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：Enable snapshot mirroring for a filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror ls`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：List the snapshot mirrored directories

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror peer_add`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：Add a remote filesystem peer

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `remote_cluster_spec` | `CephString` | 是 | 是 | - |
| `remote_fs_name` | `CephString` | 否 | 是 | - |
| `remote_mon_host` | `CephString` | 否 | 是 | - |
| `cephx_key` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror peer_bootstrap create`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：Bootstrap a filesystem peer

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `client_name` | `CephString` | 是 | 是 | - |
| `site_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror peer_bootstrap import`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：Import a bootstrap token

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `token` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror peer_list`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：List configured peers for a file system

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror peer_remove`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：Remove a filesystem peer

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `peer_uuid` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror remove`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：Remove a snapshot mirrored directory

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |
| `path` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs snapshot mirror show distribution`

- 组件/模块：`mgr module` / `mirroring`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/mirroring/module.py`
- 标志：`mgr`
- 含义：Get current instance to directory map for a filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs status`

- 组件/模块：`mgr module` / `status`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/status/module.py`
- 标志：`mgr`
- 含义：Show the status of a CephFS filesystem

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs` | `CephString` | 否 | 是 | - |
| `format` | `CephString` | 否 | 否 | - |

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
| `fs` | `CephString` | 否 | 是 | - |

## `ceph fs subvolume authorize`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Allow a cephx auth ID access to a subvolume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `auth_id` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `access_level` | `CephString` | 否 | 是 | - |
| `tenant_id` | `CephString` | 否 | 是 | - |
| `allow_existing_id` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume authorized_list`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：List auth IDs that have access to a subvolume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume charmap get`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get charmap settings for subvolumegroup

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `setting` | `CephChoices` | 否 | 是 | 可选值: casesensitive, normalization, encoding |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume charmap rm`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Remove charmap settings for subvolume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume charmap set`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Set charmap settings for subvolumegroup

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `setting` | `CephChoices` | 是 | 是 | 可选值: casesensitive, normalization, encoding |
| `value` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume create`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Create a CephFS subvolume in a volume, and optionally, with a specific size (in bytes), a specific data pool layout, a specific mode, in a specific subvolume group and in separate RADOS namespace

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `size` | `CephInt` | 否 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `pool_layout` | `CephString` | 否 | 是 | - |
| `uid` | `CephInt` | 否 | 是 | - |
| `gid` | `CephInt` | 否 | 是 | - |
| `mode` | `CephString` | 否 | 是 | - |
| `namespace_isolated` | `CephBool` | 否 | 否 | - |
| `earmark` | `CephString` | 否 | 否 | - |
| `normalization` | `CephChoices` | 否 | 否 | 可选值: nfd, nfc, nfkd, nfkc |
| `casesensitive` | `CephBool` | 否 | 否 | - |

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
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `size` | `CephInt` | 否 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `pool_layout` | `CephString` | 否 | 是 | - |
| `uid` | `CephInt` | 否 | 是 | - |
| `gid` | `CephInt` | 否 | 是 | - |
| `mode` | `CephString` | 否 | 是 | - |
| `namespace_isolated` | `CephBool` | 否 | 否 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `size` | `CephInt` | 否 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `pool_layout` | `CephString` | 否 | 是 | - |
| `uid` | `CephInt` | 否 | 是 | - |
| `gid` | `CephInt` | 否 | 是 | - |
| `mode` | `CephString` | 否 | 是 | - |
| `namespace_isolated` | `CephBool` | 否 | 否 | - |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `size` | `CephInt` | 否 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `pool_layout` | `CephString` | 否 | 是 | - |
| `uid` | `CephInt` | 否 | 是 | - |
| `gid` | `CephInt` | 否 | 是 | - |
| `mode` | `CephString` | 否 | 是 | - |
| `namespace_isolated` | `CephBool` | 否 | 否 | - |

## `ceph fs subvolume deauthorize`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Deny a cephx auth ID access to a subvolume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `auth_id` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume earmark get`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：19.2.5, 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get earmark for a subvolume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume earmark rm`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Remove earmark from a subvolume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume earmark set`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Set earmark for a subvolume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `earmark` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume evict`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Evict clients based on auth IDs and subvolume mounted

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `auth_id` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume exist`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Check a volume for the existence of a subvolume, optionally in a specified subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume getpath`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get the mountpath of a CephFS subvolume in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume info`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get the information of a CephFS subvolume in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume ls`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：List subvolumes

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume metadata get`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get custom metadata associated with the key of a CephFS subvolume in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `key_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume metadata ls`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：List custom metadata (key-value pairs) of a CephFS subvolume in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume metadata rm`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Remove custom metadata (key-value) associated with the key of a CephFS subvolume in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `key_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume metadata set`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Set custom metadata (key-value) for a CephFS subvolume in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `key_name` | `CephString` | 是 | 是 | - |
| `value` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume pin`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Set MDS pinning policy for subvolume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `pin_type` | `CephChoices` | 是 | 是 | 可选值: export, distributed, random |
| `pin_setting` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume resize`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Resize a CephFS subvolume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `new_size` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `no_shrink` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume rm`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Delete a CephFS subvolume in a volume, and optionally, in a specific subvolume group, force deleting a cancelled or failed clone, and retaining existing subvolume snapshots

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `retain_snapshots` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot clone`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Clone a snapshot to target subvolume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `target_sub_name` | `CephString` | 是 | 是 | - |
| `pool_layout` | `CephString` | 否 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `target_group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot create`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Create a snapshot of a CephFS subvolume in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot getpath`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get path for a snapshot of a CephFS subvolume located in a specific volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot info`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get the information of a CephFS subvolume snapshot and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot ls`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：List subvolume snapshots

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot metadata get`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get custom metadata associated with the key of a CephFS subvolume snapshot in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `key_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot metadata ls`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：List custom metadata (key-value pairs) of a CephFS subvolume snapshot in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot metadata rm`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Remove custom metadata (key-value) associated with the key of a CephFS subvolume snapshot in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `key_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot metadata set`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Set custom metadata (key-value) for a CephFS subvolume snapshot in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `key_name` | `CephString` | 是 | 是 | - |
| `value` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot protect`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：(deprecated) Protect snapshot of a CephFS subvolume in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot rm`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Delete a snapshot of a CephFS subvolume in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolume snapshot unprotect`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：(deprecated) Unprotect a snapshot of a CephFS subvolume in a volume, and optionally, in a specific subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `sub_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup charmap get`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get charmap settings for subvolumegroup

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | - |
| `setting` | `CephChoices` | 否 | 是 | 可选值: casesensitive, normalization, encoding |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup charmap rm`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Remove charmap settings for subvolumegroup

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup charmap set`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Set charmap settings for subvolumegroup

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | - |
| `setting` | `CephChoices` | 是 | 是 | 可选值: casesensitive, normalization, encoding |
| `value` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup create`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Create a CephFS subvolume group in a volume, and optionally, with a specific data pool layout, and a specific numeric mode

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `size` | `CephInt` | 否 | 是 | - |
| `pool_layout` | `CephString` | 否 | 是 | - |
| `uid` | `CephInt` | 否 | 是 | - |
| `gid` | `CephInt` | 否 | 是 | - |
| `mode` | `CephString` | 否 | 是 | - |
| `normalization` | `CephChoices` | 否 | 是 | 可选值: nfd, nfc, nfkd, nfkc |
| `casesensitive` | `CephBool` | 否 | 否 | - |

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
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `size` | `CephInt` | 否 | 是 | - |
| `pool_layout` | `CephString` | 否 | 是 | - |
| `uid` | `CephInt` | 否 | 是 | - |
| `gid` | `CephInt` | 否 | 是 | - |
| `mode` | `CephString` | 否 | 是 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `size` | `CephInt` | 否 | 是 | - |
| `pool_layout` | `CephString` | 否 | 是 | - |
| `uid` | `CephInt` | 否 | 是 | - |
| `gid` | `CephInt` | 否 | 是 | - |
| `mode` | `CephString` | 否 | 是 | - |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `size` | `CephInt` | 否 | 是 | - |
| `pool_layout` | `CephString` | 否 | 是 | - |
| `uid` | `CephInt` | 否 | 是 | - |
| `gid` | `CephInt` | 否 | 是 | - |
| `mode` | `CephString` | 否 | 是 | - |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `size` | `CephInt` | 否 | 是 | - |
| `pool_layout` | `CephString` | 否 | 是 | - |
| `uid` | `CephInt` | 否 | 是 | - |
| `gid` | `CephInt` | 否 | 是 | - |
| `mode` | `CephString` | 否 | 是 | - |

## `ceph fs subvolumegroup exist`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Check a volume for the existence of subvolumegroup

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup getpath`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get the mountpath of a CephFS subvolume group in a volume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup info`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get the metadata of a CephFS subvolume group in a volume,

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup ls`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：List subvolumegroups

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup pin`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Set MDS pinning policy for subvolumegroup

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | - |
| `pin_type` | `CephChoices` | 是 | 是 | 可选值: export, distributed, random |
| `pin_setting` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup resize`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Resize a CephFS subvolume group

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | - |
| `new_size` | `CephString` | 是 | 是 | - |
| `no_shrink` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup rm`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Delete a CephFS subvolume group in a volume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup snapshot create`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Create a snapshot of a CephFS subvolume group in a volume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup snapshot ls`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：List subvolumegroup snapshots

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs subvolumegroup snapshot rm`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Delete a snapshot of a CephFS subvolume group in a volume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `group_name` | `CephString` | 是 | 是 | - |
| `snap_name` | `CephString` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs swap`

- 组件/模块：`ceph mon` / `mds`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：swap ceph file system names

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `fs1_name` | `CephString` | 是 | 是 | - |
| `fs1_id` | `CephInt` | 是 | 是 | 范围: 0 |
| `fs2_name` | `CephString` | 是 | 是 | - |
| `fs2_id` | `CephInt` | 是 | 是 | 范围: 0 |
| `swap_fscids` | `CephChoices` | 是 | 是 | 可选值: yes, no |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs volume create`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Create a CephFS volume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `placement` | `CephString` | 否 | 是 | - |
| `meta_pool` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `data_pool` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |

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
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `placement` | `CephString` | 否 | 是 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `placement` | `CephString` | 否 | 是 | - |

## `ceph fs volume info`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Get the information of a CephFS volume

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `human_readable` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs volume ls`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：List volumes

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs volume rename`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Rename a CephFS volume by passing --yes-i-really-mean-it flag

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `new_vol_name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph fs volume rm`

- 组件/模块：`mgr module` / `volumes`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/pybind/mgr/volumes/module.py`
- 标志：`mgr`
- 含义：Delete a FS volume by passing --yes-i-really-mean-it flag

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `vol_name` | `CephString` | 是 | 是 | - |
| `yes-i-really-mean-it` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
