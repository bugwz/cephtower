# Osd

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph osd add-nodown`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：mark osd(s) <id> [<id>...] as nodown, or use <all|any> to mark all osds as nodown

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd add-noin`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：mark osd(s) <id> [<id>...] as noin, or use <all|any> to mark all osds as noin

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd add-noout`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：mark osd(s) <id> [<id>...] as noout, or use <all|any> to mark all osds as noout

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd add-noup`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：mark osd(s) <id> [<id>...] as noup, or use <all|any> to mark all osds as noup

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd blacklist`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：add (optionally until <expire> seconds from now) or remove <addr> from blacklist

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `blacklistop` | `CephChoices` | 是 | 是 | 可选值: add, rm |
| `addr` | `CephEntityAddr` | 是 | 是 | - |
| `expire` | `CephFloat` | 否 | 是 | 范围: 0.0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd blacklist clear`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：clear all blacklisted clients

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd blacklist ls`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：show blacklisted clients

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd blocked-by`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：print histogram of which OSDs are blocking their peers

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd blocklist`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：add (optionally until <expire> seconds from now) or remove <addr> from blocklist

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `range` | `CephString` | 否 | 是 | 字符集: `[range]` |
| `blocklistop` | `CephChoices` | 是 | 是 | 可选值: add, rm |
| `addr` | `CephEntityAddr` | 是 | 是 | - |
| `expire` | `CephFloat` | 否 | 是 | 范围: 0.0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd blocklist clear`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：clear all blocklisted clients

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd blocklist ls`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：show blocklisted clients

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd count-metadata`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：count OSDs by metadata field property

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `property` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd create`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：create new osd (with optional UUID and ID)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `uuid` | `CephUUID` | 否 | 是 | - |
| `id` | `CephOsdName` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush add`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：add or update crushmap position and weight for <name> with <weight> and location <args>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 是 | 是 | - |
| `weight` | `CephFloat` | 是 | 是 | 范围: 0.0 |
| `args` | `CephString` | 是 | 是 | 可重复; 字符集: `[A-Za-z0-9-_.=]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush add-bucket`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：add no-parent (probably root) crush bucket <name> of type <type> to location <args>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `type` | `CephString` | 是 | 是 | - |
| `args` | `CephString` | 否 | 是 | 可重复; 字符集: `[A-Za-z0-9-_.=]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush class create`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：create crush device class <class>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | - |

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
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_]` |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_]` |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_]` |

## `ceph osd crush class ls`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：list all crush device classes

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush class ls-osd`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：list all osds belonging to the specific <class>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | - |

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
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_]` |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_]` |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_]` |

## `ceph osd crush class rename`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：rename crush device class <srcname> to <dstname>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `srcname` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_]` |
| `dstname` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush class rm`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove crush device class <class>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | - |

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
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_]` |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_]` |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_]` |

## `ceph osd crush create-or-move`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：create entry or move existing entry for <name> <weight> at/to location <args>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 是 | 是 | - |
| `weight` | `CephFloat` | 是 | 是 | 范围: 0.0 |
| `args` | `CephString` | 是 | 是 | 可重复; 字符集: `[A-Za-z0-9-_.=]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush dump`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：dump crush map

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush get-device-class`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：get classes of specified osd(s) <id> [<id>...]

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush get-tunable`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：get crush tunable <tunable>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `tunable` | `CephChoices` | 是 | 是 | 可选值: straw_calc_version |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush link`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：link existing entry for <name> under location <args>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | - |
| `args` | `CephString` | 是 | 是 | 可重复; 字符集: `[A-Za-z0-9-_.=]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush ls`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：list items beneath a node in the CRUSH tree

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `node` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush move`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：move existing entry for <name> to location <args>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `args` | `CephString` | 是 | 是 | 可重复; 字符集: `[A-Za-z0-9-_.=]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush remove`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：remove <name> from crush map (everywhere, or just at <ancestor>)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `ancestor` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush rename-bucket`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：rename bucket <srcname> to <dstname>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `srcname` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `dstname` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush reweight`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：change <name>'s weight to <weight> in crush map

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `weight` | `CephFloat` | 是 | 是 | 范围: 0.0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush reweight-all`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：recalculate the weights for the tree to ensure they sum correctly

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush reweight-subtree`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：change all leaf items beneath <name> to <weight> in crush map

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `weight` | `CephFloat` | 是 | 是 | 范围: 0.0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush rm`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove <name> from crush map (everywhere, or just at <ancestor>)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `ancestor` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush rm-device-class`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove class of the osd(s) <id> [<id>...],or use <all|any> to remove all.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush rule create-erasure`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：create crush rule <name> for erasure coded pool created with <profile> (default default)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `profile` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.=]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush rule create-replicated`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：create crush rule <name> for replicated pool to start from <root>, replicate across buckets of type <type>, use devices of type <class> (ssd or hdd)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `root` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `type` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `class` | `CephString` | 否 | 是 | - |

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
| `root` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `type` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `class` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `root` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `type` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `class` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `root` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `type` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `class` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |

## `ceph osd crush rule create-simple`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：create crush rule <name> to start from <root>, replicate across buckets of type <type>, using a choose mode of <firstn|indep> (default firstn; indep best for erasure pools)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `root` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `type` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `mode` | `CephChoices` | 否 | 是 | 可选值: firstn, indep |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush rule dump`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：dump crush rule <name> (default all)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush rule list`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：list crush rules

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush rule ls`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：list crush rules

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush rule ls-by-class`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：list all crush rules that reference the same <class>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | - |

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
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |

## `ceph osd crush rule rename`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：rename crush rule <srcname> to <dstname>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `srcname` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `dstname` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush rule rm`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove crush rule <name>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush set`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set crush map from input file

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 是 | 是 | - |
| `weight` | `CephFloat` | 是 | 是 | 范围: 0.0 |
| `args` | `CephString` | 是 | 是 | 可重复; 字符集: `[A-Za-z0-9-_.=]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush set-all-straw-buckets-to-straw2`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：convert all CRUSH current straw buckets to use the straw2 algorithm

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush set-device-class`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set the <class> of the osd(s) <id> [<id>...],or use <all|any> to set all.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `class` | `CephString` | 是 | 是 | - |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush set-tunable`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set crush tunable <tunable> to <value>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `tunable` | `CephChoices` | 是 | 是 | 可选值: straw_calc_version |
| `value` | `CephInt` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush show-tunables`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：show current crush tunables

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush swap-bucket`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：swap existing bucket contents from (orphan) bucket <source> and <target>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `source` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `dest` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush tree`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：dump crush buckets and items in a tree view

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `show_shadow` | `CephBool` | 否 | 否 | - |

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
| `shadow` | `CephChoices` | 否 | 是 | 可选值: --show-shadow |

## `ceph osd crush tunables`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set crush tunables values to <profile>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `profile` | `CephChoices` | 是 | 是 | 可选值: legacy, argonaut, bobtail, firefly, hammer, jewel, optimal, default |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush unlink`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：unlink <name> from crush map (everywhere, or just at <ancestor>)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `ancestor` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush weight-set create`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：create a weight-set for a given pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `mode` | `CephChoices` | 是 | 是 | 可选值: flat, positional |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush weight-set create-compat`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：create a default backward-compatible weight-set

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush weight-set dump`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：dump crush weight sets

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush weight-set ls`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：list crush weight sets

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush weight-set reweight`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set weight for an item (bucket or osd) in a pool's weight-set

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `item` | `CephString` | 是 | 是 | - |
| `weight` | `CephFloat` | 是 | 是 | 可重复; 范围: 0.0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush weight-set reweight-compat`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set weight for an item (bucket or osd) in the backward-compatible weight-set

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `item` | `CephString` | 是 | 是 | - |
| `weight` | `CephFloat` | 是 | 是 | 可重复; 范围: 0.0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush weight-set rm`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove the weight-set for a given pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd crush weight-set rm-compat`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove the backward-compatible weight-set

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd deep-scrub`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：initiate deep scrub on osd <who>, or use <all|any> to deep scrub all

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

## `ceph osd destroy`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：mark osd as being destroyed. Keeps the ID intact (allowing reuse), but removes cephx keys, config-key data and lockbox keys, rendering data permanently unreadable.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd df`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：show OSD utilization

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `output_method` | `CephChoices` | 否 | 是 | 可选值: plain, tree |
| `filter_by` | `CephChoices` | 否 | 是 | 可选值: class, name |
| `filter` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd down`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set osd(s) <id> [<id>...] down, or use <any|all> to set all osds down

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |
| `definitely_dead` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd dump`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：print summary of OSD map

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

## `ceph osd erasure-code-profile get`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：get erasure code profile <name>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd erasure-code-profile ls`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：list all erasure code profiles

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd erasure-code-profile rm`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove erasure code profile <name>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd erasure-code-profile set`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：create erasure code profile <name> with [<key[=value]> ...] pairs. Add a --force at the end to override an existing profile (VERY DANGEROUS)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `profile` | `CephString` | 否 | 是 | 可重复 |
| `force` | `CephBool` | 否 | 否 | - |
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
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `profile` | `CephString` | 否 | 是 | 可重复 |
| `force` | `CephBool` | 否 | 否 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `profile` | `CephString` | 否 | 是 | 可重复 |
| `force` | `CephBool` | 否 | 否 | - |

## `ceph osd find`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：find osd <id> in the CRUSH map and show its location

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd force-create-pg`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：force creation of pg <pgid>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd force_healthy_stretch_mode`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：force a healthy stretch mode, requiring the full number of CRUSH buckets to peer and letting all non-tiebreaker monitors be elected leader

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd force_recovery_stretch_mode`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：try and force a recovery stretch mode, increasing the pool size to its non-failure value if currently degraded and all monitor buckets are up

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd get-require-min-compat-client`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：get the minimum client version we will maintain compatibility with

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd getcrushmap`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：get CRUSH map

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `epoch` | `CephInt` | 否 | 是 | 范围: 0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout 或 `-o <file>`; map 类命令通常输出二进制 map 数据，JSON 不适用。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：否，通常为二进制 map 输出。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd getmap`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：get OSD map

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `epoch` | `CephInt` | 否 | 是 | 范围: 0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout 或 `-o <file>`; map 类命令通常输出二进制 map 数据，JSON 不适用。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：否，通常为二进制 map 输出。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd getmaxosd`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：show largest OSD id

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd in`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set osd(s) <id> [<id>...] in, can use <any|all> to automatically set all previously out osds in

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd info`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：print osd's {id} information (instead of all osds from map)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd last-stat-seq`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：get the last pg stats sequence number reported for this osd

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd lost`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：mark osd as permanently lost. THIS DESTROYS DATA IF NO MORE REPLICAS EXIST, BE CAREFUL

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd ls`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：show all OSD ids

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

## `ceph osd ls-tree`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：show OSD ids under bucket <name> in the CRUSH map

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `epoch` | `CephInt` | 否 | 是 | 范围: 0 |
| `name` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd lspools`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：list pools

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd map`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：find pg for <object> in <pool> with [namespace]

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `object` | `CephObjectname` | 是 | 是 | - |
| `nspace` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd metadata`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：fetch metadata for osd {id} (default all)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd new`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：Create a new OSD. If supplied, the `id` to be replaced needs to exist and have been previously destroyed. Reads secrets from JSON file via `-i <file>` (see man page).

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `uuid` | `CephUUID` | 是 | 是 | - |
| `id` | `CephOsdName` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd numa-status`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：show NUMA status of OSDs

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd ok-to-stop`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：check whether osd(s) can be safely stopped without reducing immediate data availability

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |
| `max` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd ok-to-upgrade`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：determine a safe number of osd(s) subject to a maximum(if specified) within the provided CRUSH bucket that can be safely upgraded without reducing immediate data availability

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `crush_bucket` | `CephString` | 是 | 是 | - |
| `ceph_version` | `CephString` | 是 | 是 | - |
| `max` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd out`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set osd(s) <id> [<id>...] out, or use <any|all> to set all osds out

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pause`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：pause osd

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd perf`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：print dump of OSD perf summary stats

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd perf counters get`

- 组件/模块：`mgr module` / `osd_perf_query`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/osd_perf_query/module.py`
- 标志：`mgr`
- 含义：fetch osd perf counters

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `query_id` | `CephInt` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd perf query add`

- 组件/模块：`mgr module` / `osd_perf_query`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/osd_perf_query/module.py`
- 标志：`mgr`
- 含义：add osd perf query

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `query` | `CephChoices` | 是 | 是 | 可选值: client_id, rbd_image_id, all_subkeys |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd perf query remove`

- 组件/模块：`mgr module` / `osd_perf_query`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/osd_perf_query/module.py`
- 标志：`mgr`
- 含义：remove osd perf query

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `query_id` | `CephInt` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pg-temp`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set pg_temp mapping pgid:[<id> [<id>...]] (developers only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |
| `id` | `CephOsdName` | 否 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pg-upmap`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set pg_upmap mapping <pgid>:[<id> [<id>...]] (developers only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |
| `id` | `CephOsdName` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pg-upmap-items`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set pg_upmap_items mapping <pgid>:{<id> to <id>, [...]} (developers only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |
| `id` | `CephOsdName` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pg-upmap-primary`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set pg primary osd <pgid>:<id> (id (osd) must be part of pgid)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |
| `id` | `CephOsdName` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool application disable`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：disables use of an application <app> on pool <poolname>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `app` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool application enable`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：enable use of an application <app> [cephfs,rbd,rgw] on pool <poolname>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `app` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool application get`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：get value of key <key> of application <app> on pool <poolname>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `app` | `CephString` | 否 | 是 | - |
| `key` | `CephString` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool application rm`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：removes application <app> metadata key <key> on pool <poolname>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `app` | `CephString` | 是 | 是 | - |
| `key` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool application set`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：sets application <app> metadata key <key> to <value> on pool <poolname>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `app` | `CephString` | 是 | 是 | - |
| `key` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `value` | `CephString` | 是 | 是 | 字符集: `[A-Za-z0-9-_.=]` |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool autoscale-status`

- 组件/模块：`mgr module` / `pg_autoscaler`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/pg_autoscaler/module.py`
- 标志：`mgr`
- 含义：report on pool pg_num sizing recommendation and intent

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

## `ceph osd pool availability-status`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：obtain availability stats from all pools

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool cancel-force-backfill`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：restore normal recovery priority of specified pool <who>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `who` | `CephPoolname` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool cancel-force-recovery`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：restore normal recovery priority of specified pool <who>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `who` | `CephPoolname` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool clear-availability-status`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：clear a pool's existing availability stats

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool create`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：create pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `pg_num` | `CephInt` | 否 | 是 | 范围: 0 |
| `pgp_num` | `CephInt` | 否 | 是 | 范围: 0 |
| `pool_type` | `CephChoices` | 否 | 是 | 可选值: replicated, erasure |
| `erasure_code_profile` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `rule` | `CephString` | 否 | 是 | - |
| `expected_num_objects` | `CephInt` | 否 | 是 | 范围: 0 |
| `size` | `CephInt` | 否 | 是 | 范围: 0 |
| `pg_num_min` | `CephInt` | 否 | 是 | 范围: 0 |
| `pg_num_max` | `CephInt` | 否 | 是 | 范围: 0 |
| `autoscale_mode` | `CephChoices` | 否 | 是 | 可选值: on, off, warn |
| `bulk` | `CephBool` | 否 | 否 | - |
| `target_size_bytes` | `CephInt` | 否 | 否 | 范围: 0 |
| `target_size_ratio` | `CephFloat` | 否 | 否 | 范围: 0.0 |
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
| `pool` | `CephPoolname` | 是 | 是 | - |
| `pg_num` | `CephInt` | 否 | 是 | 范围: 0 |
| `pgp_num` | `CephInt` | 否 | 是 | 范围: 0 |
| `pool_type` | `CephChoices` | 否 | 是 | 可选值: replicated, erasure |
| `erasure_code_profile` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `rule` | `CephString` | 否 | 是 | - |
| `expected_num_objects` | `CephInt` | 否 | 是 | 范围: 0 |
| `size` | `CephInt` | 否 | 是 | 范围: 0 |
| `pg_num_min` | `CephInt` | 否 | 是 | 范围: 0 |
| `pg_num_max` | `CephInt` | 否 | 是 | 范围: 0 |
| `autoscale_mode` | `CephChoices` | 否 | 是 | 可选值: on, off, warn |
| `bulk` | `CephBool` | 否 | 否 | - |
| `target_size_bytes` | `CephInt` | 否 | 否 | 范围: 0 |
| `target_size_ratio` | `CephFloat` | 否 | 否 | 范围: 0.0 |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `pg_num` | `CephInt` | 否 | 是 | 范围: 0 |
| `pgp_num` | `CephInt` | 否 | 是 | 范围: 0 |
| `pool_type` | `CephChoices` | 否 | 是 | 可选值: replicated, erasure |
| `erasure_code_profile` | `CephString` | 否 | 是 | 字符集: `[A-Za-z0-9-_.]` |
| `rule` | `CephString` | 否 | 是 | - |
| `expected_num_objects` | `CephInt` | 否 | 是 | 范围: 0 |
| `size` | `CephInt` | 否 | 是 | 范围: 0 |
| `pg_num_min` | `CephInt` | 否 | 是 | 范围: 0 |
| `pg_num_max` | `CephInt` | 否 | 是 | 范围: 0 |
| `autoscale_mode` | `CephChoices` | 否 | 是 | 可选值: on, off, warn |
| `bulk` | `CephBool` | 否 | 否 | - |
| `target_size_bytes` | `CephInt` | 否 | 否 | 范围: 0 |
| `target_size_ratio` | `CephFloat` | 否 | 否 | 范围: 0.0 |

## `ceph osd pool deep-scrub`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：initiate deep-scrub on pool <who>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `who` | `CephPoolname` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool delete`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：delete pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `pool2` | `CephPoolname` | 否 | 是 | - |
| `yes_i_really_really_mean_it` | `CephBool` | 否 | 否 | - |
| `yes_i_really_really_mean_it_not_faking` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool force-backfill`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：force backfill of specified pool <who> first

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `who` | `CephPoolname` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool force-recovery`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：force recovery of specified pool <who> first

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `who` | `CephPoolname` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool force-remove-snap`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：Forces removal of snapshots in the range [lower_snapid_bound, upper_snapid_bound) on pool <pool> in order to cause OSDs to re-trim them.

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `lower_snapid_bound` | `CephInt` | 否 | 是 | 范围: 0 |
| `upper_snapid_bound` | `CephInt` | 否 | 是 | 范围: 0 |
| `dry_run` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool get`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：get pool parameter <var>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: size, min_size, pg_num, pgp_num, crush_rule, hashpspool, nodelete, nopgchange, nosizechange, write_fadvise_dontneed, noscrub, nodeep-scrub, hit_set_type, hit_set_period, hit_set_count, hit_set_fpp, use_gmt_hitset, target_max_objects, target_max_bytes, cache_target_dirty_ratio, cache_target_dirty_high_ratio, cache_target_full_ratio, cache_min_flush_age, cache_min_evict_age, erasure_code_profile, min_read_recency_for_promote, all, min_write_recency_for_promote, fast_read, hit_set_grade_decay_rate, hit_set_search_last_n, scrub_min_interval, scrub_max_interval, deep_scrub_interval, recovery_priority, recovery_op_priority, scrub_priority, compression_mode, compression_algorithm, compression_required_ratio, compression_max_blob_size, compression_min_blob_size, csum_type, csum_min_block, csum_max_block, allow_ec_overwrites, fingerprint_algorithm, pg_autoscale_mode, pg_autoscale_bias, pg_num_min, pg_num_max, target_size_bytes, target_size_ratio, dedup_tier, dedup_chunk_algorithm, dedup_cdc_chunk_size, eio, bulk, read_ratio, pct_update_delay, allow_ec_optimizations |

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
| `pool` | `CephPoolname` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: size, min_size, pg_num, pgp_num, crush_rule, hashpspool, nodelete, nopgchange, nosizechange, write_fadvise_dontneed, noscrub, nodeep-scrub, hit_set_type, hit_set_period, hit_set_count, hit_set_fpp, use_gmt_hitset, target_max_objects, target_max_bytes, cache_target_dirty_ratio, cache_target_dirty_high_ratio, cache_target_full_ratio, cache_min_flush_age, cache_min_evict_age, erasure_code_profile, min_read_recency_for_promote, all, min_write_recency_for_promote, fast_read, hit_set_grade_decay_rate, hit_set_search_last_n, scrub_min_interval, scrub_max_interval, deep_scrub_interval, recovery_priority, recovery_op_priority, scrub_priority, compression_mode, compression_algorithm, compression_required_ratio, compression_max_blob_size, compression_min_blob_size, csum_type, csum_min_block, csum_max_block, allow_ec_overwrites, fingerprint_algorithm, pg_autoscale_mode, pg_autoscale_bias, pg_num_min, pg_num_max, target_size_bytes, target_size_ratio, dedup_tier, dedup_chunk_algorithm, dedup_cdc_chunk_size, bulk |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: size, min_size, pg_num, pgp_num, crush_rule, hashpspool, nodelete, nopgchange, nosizechange, write_fadvise_dontneed, noscrub, nodeep-scrub, hit_set_type, hit_set_period, hit_set_count, hit_set_fpp, use_gmt_hitset, target_max_objects, target_max_bytes, cache_target_dirty_ratio, cache_target_dirty_high_ratio, cache_target_full_ratio, cache_min_flush_age, cache_min_evict_age, erasure_code_profile, min_read_recency_for_promote, all, min_write_recency_for_promote, fast_read, hit_set_grade_decay_rate, hit_set_search_last_n, scrub_min_interval, scrub_max_interval, deep_scrub_interval, recovery_priority, recovery_op_priority, scrub_priority, compression_mode, compression_algorithm, compression_required_ratio, compression_max_blob_size, compression_min_blob_size, csum_type, csum_min_block, csum_max_block, allow_ec_overwrites, fingerprint_algorithm, pg_autoscale_mode, pg_autoscale_bias, pg_num_min, pg_num_max, target_size_bytes, target_size_ratio, dedup_tier, dedup_chunk_algorithm, dedup_cdc_chunk_size, eio, bulk |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: size, min_size, pg_num, pgp_num, crush_rule, hashpspool, nodelete, nopgchange, nosizechange, write_fadvise_dontneed, noscrub, nodeep-scrub, hit_set_type, hit_set_period, hit_set_count, hit_set_fpp, use_gmt_hitset, target_max_objects, target_max_bytes, cache_target_dirty_ratio, cache_target_dirty_high_ratio, cache_target_full_ratio, cache_min_flush_age, cache_min_evict_age, erasure_code_profile, min_read_recency_for_promote, all, min_write_recency_for_promote, fast_read, hit_set_grade_decay_rate, hit_set_search_last_n, scrub_min_interval, scrub_max_interval, deep_scrub_interval, recovery_priority, recovery_op_priority, scrub_priority, compression_mode, compression_algorithm, compression_required_ratio, compression_max_blob_size, compression_min_blob_size, csum_type, csum_min_block, csum_max_block, allow_ec_overwrites, fingerprint_algorithm, pg_autoscale_mode, pg_autoscale_bias, pg_num_min, pg_num_max, target_size_bytes, target_size_ratio, dedup_tier, dedup_chunk_algorithm, dedup_cdc_chunk_size, eio, bulk |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: size, min_size, pg_num, pgp_num, crush_rule, hashpspool, nodelete, nopgchange, nosizechange, write_fadvise_dontneed, noscrub, nodeep-scrub, hit_set_type, hit_set_period, hit_set_count, hit_set_fpp, use_gmt_hitset, target_max_objects, target_max_bytes, cache_target_dirty_ratio, cache_target_dirty_high_ratio, cache_target_full_ratio, cache_min_flush_age, cache_min_evict_age, erasure_code_profile, min_read_recency_for_promote, all, min_write_recency_for_promote, fast_read, hit_set_grade_decay_rate, hit_set_search_last_n, scrub_min_interval, scrub_max_interval, deep_scrub_interval, recovery_priority, recovery_op_priority, scrub_priority, compression_mode, compression_algorithm, compression_required_ratio, compression_max_blob_size, compression_min_blob_size, csum_type, csum_min_block, csum_max_block, allow_ec_overwrites, fingerprint_algorithm, pg_autoscale_mode, pg_autoscale_bias, pg_num_min, pg_num_max, target_size_bytes, target_size_ratio, dedup_tier, dedup_chunk_algorithm, dedup_cdc_chunk_size, eio, bulk, read_ratio |

## `ceph osd pool get noautoscale`

- 组件/模块：`mgr module` / `pg_autoscaler`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/pg_autoscaler/module.py`
- 标志：`mgr`
- 含义：Get the noautoscale flag to see if all pools are setting the autoscaler on or off as well as newly created pools in the future.

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool get-quota`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：obtain object or byte limits for pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool ls`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：list pools

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `detail` | `CephChoices` | 否 | 是 | 可选值: detail |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool mksnap`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：make snapshot <snap> in <pool>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `snap` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool rename`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：rename <srcpool> to <destpool>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `srcpool` | `CephPoolname` | 是 | 是 | - |
| `destpool` | `CephPoolname` | 是 | 是 | - |
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
| `srcpool` | `CephPoolname` | 是 | 是 | - |
| `destpool` | `CephPoolname` | 是 | 是 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `srcpool` | `CephPoolname` | 是 | 是 | - |
| `destpool` | `CephPoolname` | 是 | 是 | - |

## `ceph osd pool repair`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：initiate repair on pool <who>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `who` | `CephPoolname` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool rm`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `pool2` | `CephPoolname` | 否 | 是 | - |
| `yes_i_really_really_mean_it` | `CephBool` | 否 | 否 | - |
| `yes_i_really_really_mean_it_not_faking` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool rmsnap`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove snapshot <snap> from <pool>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `snap` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool scrub`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：initiate scrub on pool <who>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `who` | `CephPoolname` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool set`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set pool parameter <var> to <val>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: size, min_size, pg_num, pgp_num, pgp_num_actual, crush_rule, hashpspool, nodelete, nopgchange, nosizechange, write_fadvise_dontneed, noscrub, nodeep-scrub, hit_set_type, hit_set_period, hit_set_count, hit_set_fpp, use_gmt_hitset, target_max_bytes, target_max_objects, cache_target_dirty_ratio, cache_target_dirty_high_ratio, cache_target_full_ratio, cache_min_flush_age, cache_min_evict_age, min_read_recency_for_promote, min_write_recency_for_promote, fast_read, hit_set_grade_decay_rate, hit_set_search_last_n, scrub_min_interval, scrub_max_interval, deep_scrub_interval, recovery_priority, recovery_op_priority, scrub_priority, compression_mode, compression_algorithm, compression_required_ratio, compression_max_blob_size, compression_min_blob_size, csum_type, csum_min_block, csum_max_block, allow_ec_overwrites, fingerprint_algorithm, pg_autoscale_mode, pg_autoscale_bias, pg_num_min, pg_num_max, target_size_bytes, target_size_ratio, dedup_tier, dedup_chunk_algorithm, dedup_cdc_chunk_size, eio, bulk, read_ratio, pct_update_delay, allow_ec_optimizations |
| `val` | `CephString` | 是 | 是 | - |
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
| `pool` | `CephPoolname` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: size, min_size, pg_num, pgp_num, pgp_num_actual, crush_rule, hashpspool, nodelete, nopgchange, nosizechange, write_fadvise_dontneed, noscrub, nodeep-scrub, hit_set_type, hit_set_period, hit_set_count, hit_set_fpp, use_gmt_hitset, target_max_bytes, target_max_objects, cache_target_dirty_ratio, cache_target_dirty_high_ratio, cache_target_full_ratio, cache_min_flush_age, cache_min_evict_age, min_read_recency_for_promote, min_write_recency_for_promote, fast_read, hit_set_grade_decay_rate, hit_set_search_last_n, scrub_min_interval, scrub_max_interval, deep_scrub_interval, recovery_priority, recovery_op_priority, scrub_priority, compression_mode, compression_algorithm, compression_required_ratio, compression_max_blob_size, compression_min_blob_size, csum_type, csum_min_block, csum_max_block, allow_ec_overwrites, fingerprint_algorithm, pg_autoscale_mode, pg_autoscale_bias, pg_num_min, pg_num_max, target_size_bytes, target_size_ratio, dedup_tier, dedup_chunk_algorithm, dedup_cdc_chunk_size, bulk |
| `val` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: size, min_size, pg_num, pgp_num, pgp_num_actual, crush_rule, hashpspool, nodelete, nopgchange, nosizechange, write_fadvise_dontneed, noscrub, nodeep-scrub, hit_set_type, hit_set_period, hit_set_count, hit_set_fpp, use_gmt_hitset, target_max_bytes, target_max_objects, cache_target_dirty_ratio, cache_target_dirty_high_ratio, cache_target_full_ratio, cache_min_flush_age, cache_min_evict_age, min_read_recency_for_promote, min_write_recency_for_promote, fast_read, hit_set_grade_decay_rate, hit_set_search_last_n, scrub_min_interval, scrub_max_interval, deep_scrub_interval, recovery_priority, recovery_op_priority, scrub_priority, compression_mode, compression_algorithm, compression_required_ratio, compression_max_blob_size, compression_min_blob_size, csum_type, csum_min_block, csum_max_block, allow_ec_overwrites, fingerprint_algorithm, pg_autoscale_mode, pg_autoscale_bias, pg_num_min, pg_num_max, target_size_bytes, target_size_ratio, dedup_tier, dedup_chunk_algorithm, dedup_cdc_chunk_size, eio, bulk |
| `val` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: size, min_size, pg_num, pgp_num, pgp_num_actual, crush_rule, hashpspool, nodelete, nopgchange, nosizechange, write_fadvise_dontneed, noscrub, nodeep-scrub, hit_set_type, hit_set_period, hit_set_count, hit_set_fpp, use_gmt_hitset, target_max_bytes, target_max_objects, cache_target_dirty_ratio, cache_target_dirty_high_ratio, cache_target_full_ratio, cache_min_flush_age, cache_min_evict_age, min_read_recency_for_promote, min_write_recency_for_promote, fast_read, hit_set_grade_decay_rate, hit_set_search_last_n, scrub_min_interval, scrub_max_interval, deep_scrub_interval, recovery_priority, recovery_op_priority, scrub_priority, compression_mode, compression_algorithm, compression_required_ratio, compression_max_blob_size, compression_min_blob_size, csum_type, csum_min_block, csum_max_block, allow_ec_overwrites, fingerprint_algorithm, pg_autoscale_mode, pg_autoscale_bias, pg_num_min, pg_num_max, target_size_bytes, target_size_ratio, dedup_tier, dedup_chunk_algorithm, dedup_cdc_chunk_size, eio, bulk |
| `val` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `var` | `CephChoices` | 是 | 是 | 可选值: size, min_size, pg_num, pgp_num, pgp_num_actual, crush_rule, hashpspool, nodelete, nopgchange, nosizechange, write_fadvise_dontneed, noscrub, nodeep-scrub, hit_set_type, hit_set_period, hit_set_count, hit_set_fpp, use_gmt_hitset, target_max_bytes, target_max_objects, cache_target_dirty_ratio, cache_target_dirty_high_ratio, cache_target_full_ratio, cache_min_flush_age, cache_min_evict_age, min_read_recency_for_promote, min_write_recency_for_promote, fast_read, hit_set_grade_decay_rate, hit_set_search_last_n, scrub_min_interval, scrub_max_interval, deep_scrub_interval, recovery_priority, recovery_op_priority, scrub_priority, compression_mode, compression_algorithm, compression_required_ratio, compression_max_blob_size, compression_min_blob_size, csum_type, csum_min_block, csum_max_block, allow_ec_overwrites, fingerprint_algorithm, pg_autoscale_mode, pg_autoscale_bias, pg_num_min, pg_num_max, target_size_bytes, target_size_ratio, dedup_tier, dedup_chunk_algorithm, dedup_cdc_chunk_size, eio, bulk, read_ratio |
| `val` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

## `ceph osd pool set noautoscale`

- 组件/模块：`mgr module` / `pg_autoscaler`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/pg_autoscaler/module.py`
- 标志：`mgr`
- 含义：set the noautoscale for all pools (including newly created pools in the future) and complete all on-going progress events regarding PG-autoscaling.

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool set threshold`

- 组件/模块：`mgr module` / `pg_autoscaler`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/pg_autoscaler/module.py`
- 标志：`mgr`
- 含义：set the autoscaler threshold A.K.A. the factor by which the new PG_NUM must vary from the existing PG_NUM

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `num` | `CephFloat` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool set-quota`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set object or byte limit on pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `field` | `CephChoices` | 是 | 是 | 可选值: max_objects, max_bytes |
| `val` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool stats`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：obtain stats from all pools, or from specified pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool_name` | `CephPoolname` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool stretch set`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：make the pool stretched across the specified number of CRUSH buckets

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `peering_crush_bucket_count` | `CephInt` | 是 | 是 | 范围: 0 |
| `peering_crush_bucket_target` | `CephInt` | 是 | 是 | 范围: 0 |
| `peering_crush_bucket_barrier` | `CephString` | 是 | 是 | - |
| `crush_rule` | `CephString` | 是 | 是 | - |
| `size` | `CephInt` | 是 | 是 | 范围: 0 |
| `min_size` | `CephInt` | 是 | 是 | 范围: 0 |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool stretch show`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：19.2.5, 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：show all the stretch related information for the pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd pool stretch unset`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：unset the stretch mode for the pool

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `crush_rule` | `CephString` | 是 | 是 | - |
| `size` | `CephInt` | 是 | 是 | 范围: 0 |
| `min_size` | `CephInt` | 是 | 是 | 范围: 0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

### 版本差异

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |

## `ceph osd pool unset noautoscale`

- 组件/模块：`mgr module` / `pg_autoscaler`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/pg_autoscaler/module.py`
- 标志：`mgr`
- 含义：Unset the noautoscale flag so all pools will have autoscale enabled (including newly created pools in the future).

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd primary-affinity`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：adjust osd primary-affinity from 0.0 <= <weight> <= 1.0

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 是 | 是 | - |
| `weight` | `CephFloat` | 是 | 是 | 范围: 0.0..1.0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd primary-temp`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set primary_temp mapping pgid:<id>|-1 (developers only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |
| `id` | `CephOsdName` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd purge`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：purge all osd data from the monitors including the OSD id and CRUSH position

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 是 | 是 | - |
| `force` | `CephBool` | 否 | 否 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd purge-new`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：purge all traces of an OSD that was partially created but never started

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd repair`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：initiate repair on osd <who>, or use <all|any> to repair all

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

## `ceph osd require-osd-release`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set the minimum allowed OSD release to participate in the cluster

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `release` | `CephChoices` | 是 | 是 | 可选值: octopus, pacific, quincy, reef, squid, tentacle |
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
| `release` | `CephChoices` | 是 | 是 | 可选值: luminous, mimic, nautilus, octopus, pacific |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `release` | `CephChoices` | 是 | 是 | 可选值: octopus, pacific, quincy |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

- `18.2.8` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `release` | `CephChoices` | 是 | 是 | 可选值: octopus, pacific, quincy, reef |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

- `19.2.5` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `release` | `CephChoices` | 是 | 是 | 可选值: octopus, pacific, quincy, reef, squid |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

## `ceph osd reweight`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：reweight osd to 0.0 < <weight> < 1.0

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `CephOsdName` | 是 | 是 | - |
| `weight` | `CephFloat` | 是 | 是 | 范围: 0.0..1.0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd reweight-by-pg`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：reweight OSDs by PG distribution [overload-percentage-for-consideration, default 120]

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `oload` | `CephInt` | 否 | 是 | - |
| `max_change` | `CephFloat` | 否 | 是 | - |
| `max_osds` | `CephInt` | 否 | 是 | - |
| `pools` | `CephPoolname` | 否 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd reweight-by-utilization`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：reweight OSDs by utilization [overload-percentage-for-consideration, default 120]

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `oload` | `CephInt` | 否 | 是 | - |
| `max_change` | `CephFloat` | 否 | 是 | - |
| `max_osds` | `CephInt` | 否 | 是 | - |
| `no_increasing` | `CephBool` | 否 | 否 | - |

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
| `oload` | `CephInt` | 否 | 是 | - |
| `max_change` | `CephFloat` | 否 | 是 | - |
| `max_osds` | `CephInt` | 否 | 是 | - |
| `no_increasing` | `CephChoices` | 否 | 是 | 可选值: --no-increasing |

## `ceph osd reweightn`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：reweight osds with {<id>: <weight>,...}

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `weights` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd rm`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：remove osd(s) <id> [<id>...], or use <any|all> to remove all osds

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd rm-nodown`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：allow osd(s) <id> [<id>...] to be marked down (if they are currently marked as nodown), can use <all|any> to automatically filter out all nodown osds

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd rm-noin`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：allow osd(s) <id> [<id>...] to be marked in (if they are currently marked as noin), can use <all|any> to automatically filter out all noin osds

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd rm-noout`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：allow osd(s) <id> [<id>...] to be marked out (if they are currently marked as noout), can use <all|any> to automatically filter out all noout osds

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd rm-noup`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：allow osd(s) <id> [<id>...] to be marked up (if they are currently marked as noup), can use <all|any> to automatically filter out all noup osds

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd rm-pg-upmap`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：clear pg_upmap mapping for <pgid> (developers only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd rm-pg-upmap-items`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：clear pg_upmap_items mapping for <pgid> (developers only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd rm-pg-upmap-primary`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：clear pg primary setting for <pgid>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd rm-pg-upmap-primary-all`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：clear all pg primary entries (developers only)

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd rm-primary-temp`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：clear primary_temp mapping pgid (developers only)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pgid` | `CephPgid` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd safe-to-destroy`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：check whether osd(s) can be safely destroyed without reducing data durability

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd scrub`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mgr/MgrCommands.h`
- 含义：initiate scrub on osd <who>, or use <all|any> to scrub all

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

## `ceph osd set`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set <key>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `key` | `CephChoices` | 是 | 是 | 可选值: full, pause, noup, nodown, noout, noin, nobackfill, norebalance, norecover, noscrub, nodeep-scrub, notieragent, nosnaptrim, pglog_hardlimit, noautoscale |
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
| `key` | `CephChoices` | 是 | 是 | 可选值: full, pause, noup, nodown, noout, noin, nobackfill, norebalance, norecover, noscrub, nodeep-scrub, notieragent, nosnaptrim, pglog_hardlimit |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

## `ceph osd set-allow-crimson`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：18.2.8, 19.2.5, 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：Allow crimson-osds to boot and join the cluster.  Note, crimson-osd is not yet considered stable and may crash or cause data loss -- should be avoided outside of testing and development.  This setting is irrevocable

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd set-backfillfull-ratio`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set usage ratio at which OSDs are marked too full to backfill

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ratio` | `CephFloat` | 是 | 是 | 范围: 0.0..1.0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd set-full-ratio`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set usage ratio at which OSDs are marked full

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ratio` | `CephFloat` | 是 | 是 | 范围: 0.0..1.0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd set-group`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set <flags> for batch osds or crush nodes, <flags> must be a comma-separated subset of {noup,nodown,noin,noout}

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `flags` | `CephString` | 是 | 是 | - |
| `who` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd set-nearfull-ratio`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set usage ratio at which OSDs are marked near-full

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ratio` | `CephFloat` | 是 | 是 | 范围: 0.0..1.0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd set-require-min-compat-client`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set the minimum client version we will maintain compatibility with

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `version` | `CephString` | 是 | 是 | - |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd setcrushmap`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set crush map from input file

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `prior_version` | `CephInt` | 否 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd setmaxosd`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set new maximum osd value

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `newmax` | `CephInt` | 是 | 是 | 范围: 0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd stat`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：print summary of OSD map

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd status`

- 组件/模块：`mgr module` / `status`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/pybind/mgr/status/module.py`
- 标志：`mgr`
- 含义：Show the status of OSDs within a bucket, or all

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `bucket` | `CephString` | 否 | 是 | - |
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
| `bucket` | `CephString` | 否 | 是 | - |

- `17.2.9` 参数：

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `bucket` | `CephString` | 否 | 是 | - |

## `ceph osd stop`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：stop the corresponding osd daemons and mark them as down

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `ids` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd test-reweight-by-pg`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：dry run of reweight OSDs by PG distribution [overload-percentage-for-consideration, default 120]

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `oload` | `CephInt` | 否 | 是 | - |
| `max_change` | `CephFloat` | 否 | 是 | - |
| `max_osds` | `CephInt` | 否 | 是 | - |
| `pools` | `CephPoolname` | 否 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd test-reweight-by-utilization`

- 组件/模块：`ceph mgr` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mgr/MgrCommands.h`
- 含义：dry run of reweight OSDs by utilization [overload-percentage-for-consideration, default 120]

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `oload` | `CephInt` | 否 | 是 | - |
| `max_change` | `CephFloat` | 否 | 是 | - |
| `max_osds` | `CephInt` | 否 | 是 | - |
| `no_increasing` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd tier add`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：add the tier <tierpool> (the second one) to base pool <pool> (the first one)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `tierpool` | `CephPoolname` | 是 | 是 | - |
| `force_nonempty` | `CephBool` | 否 | 否 | - |

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
| `pool` | `CephPoolname` | 是 | 是 | - |
| `tierpool` | `CephPoolname` | 是 | 是 | - |
| `force_nonempty` | `CephChoices` | 否 | 是 | 可选值: --force-nonempty |

## `ceph osd tier add-cache`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：add a cache <tierpool> (the second one) of size <size> to existing pool <pool> (the first one)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `tierpool` | `CephPoolname` | 是 | 是 | - |
| `size` | `CephInt` | 是 | 是 | 范围: 0 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd tier cache-mode`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：specify the caching mode for cache tier <pool>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `mode` | `CephChoices` | 是 | 是 | 可选值: writeback, proxy, readproxy, readonly, none |
| `yes_i_really_mean_it` | `CephBool` | 否 | 否 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd tier remove`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：remove the tier <tierpool> (the second one) from base pool <pool> (the first one)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `tierpool` | `CephPoolname` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd tier remove-overlay`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 标志：`deprecated`
- 含义：remove the overlay pool for base pool <pool>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd tier rm`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove the tier <tierpool> (the second one) from base pool <pool> (the first one)

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `tierpool` | `CephPoolname` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd tier rm-overlay`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：remove the overlay pool for base pool <pool>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd tier set-overlay`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：set the overlay pool for base pool <pool> to be <overlaypool>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `pool` | `CephPoolname` | 是 | 是 | - |
| `overlaypool` | `CephPoolname` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd tree`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：print OSD tree

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `epoch` | `CephInt` | 否 | 是 | 范围: 0 |
| `states` | `CephChoices` | 否 | 是 | 可重复; 可选值: up, down, in, out, destroyed |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd tree-from`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：print OSD tree in bucket

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `epoch` | `CephInt` | 否 | 是 | 范围: 0 |
| `bucket` | `CephString` | 是 | 是 | - |
| `states` | `CephChoices` | 否 | 是 | 可重复; 可选值: up, down, in, out, destroyed |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd unpause`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：unpause osd

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd unset`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：unset <key>

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `key` | `CephChoices` | 是 | 是 | 可选值: full, pause, noup, nodown, noout, noin, nobackfill, norebalance, norecover, noscrub, nodeep-scrub, notieragent, nosnaptrim, noautoscale |

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
| `key` | `CephChoices` | 是 | 是 | 可选值: full, pause, noup, nodown, noout, noin, nobackfill, norebalance, norecover, noscrub, nodeep-scrub, notieragent, nosnaptrim |

## `ceph osd unset-group`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`rw`
- 来源：`src/mon/MonCommands.h`
- 含义：unset <flags> for batch osds or crush nodes, <flags> must be a comma-separated subset of {noup,nodown,noin,noout}

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `flags` | `CephString` | 是 | 是 | - |
| `who` | `CephString` | 是 | 是 | 可重复 |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd utilization`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：get basic pg distribution stats

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph osd versions`

- 组件/模块：`ceph mon` / `osd`
- 支持版本：16.2.15 - 20.2.2
- 权限：`r`
- 来源：`src/mon/MonCommands.h`
- 含义：check running versions of OSDs

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
