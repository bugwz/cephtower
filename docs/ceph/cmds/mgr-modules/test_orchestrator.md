# Test_Orchestrator

> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。

## `ceph test_orchestrator load_data`

- 组件/模块：`mgr module` / `test_orchestrator`
- 支持版本：16.2.15 - 20.2.2
- 权限：`w`
- 来源：`src/pybind/mgr/test_orchestrator/module.py`
- 标志：`mgr`
- 含义：load dummy data into test orchestrator

### 参数

无。

### 返回信息

- 成功：退出码 `0`；输出通道：stdout; 错误信息通常在 stderr。
- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。
- JSON：通常不需要；写操作多返回确认文本或空 stdout。
- JSON 返回形态：未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema。

#### 返回字段

未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。
