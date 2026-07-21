# Feedback

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph feedback delete api-key`

- 组件/模块：`mgr module` / `feedback`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/feedback/module.py`
- 标志：`mgr`
- 含义：Delete Ceph Issue Tracker API key

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph feedback get api-key`

- 组件/模块：`mgr module` / `feedback`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/feedback/module.py`
- 标志：`mgr`
- 含义：Get Ceph Issue Tracker API key

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph feedback issue list`

- 组件/模块：`mgr module` / `feedback`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/feedback/module.py`
- 标志：`mgr`
- 含义：Fetch issue list

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph feedback issue report`

- 组件/模块：`mgr module` / `feedback`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/feedback/module.py`
- 标志：`mgr`
- 含义：Create an issue

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `project` | `CephString` | 是 | 是 | - |
| `tracker` | `CephString` | 是 | 是 | - |
| `subject` | `CephString` | 是 | 是 | - |
| `description` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。

## `ceph feedback set api-key`

- 组件/模块：`mgr module` / `feedback`
- 支持版本：17.2.9, 18.2.8, 19.2.5, 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/feedback/module.py`
- 标志：`mgr`
- 含义：Set Ceph Issue Tracker API key

### 参数

| 参数 | 类型 | 必填 | 位置参数 | 说明 |
| --- | --- | --- | --- | --- |
| `key` | `CephString` | 是 | 是 | - |

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
