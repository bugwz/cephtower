# Ceph Dashboard API 版本兼容性

本文档汇总 v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2 的 Dashboard API operation 支持情况。

## 版本统计

| 版本 | 路径数 | 接口操作数 |
| --- | --- | --- |
| v16.2.15 | 134 | 195 |
| v17.2.9 | 147 | 214 |
| v18.2.8 | 195 | 278 |
| v19.2.5 | 220 | 314 |
| v20.2.2 | 281 | 403 |

## v20.2.2 未包含的历史接口

| 接口 | 支持版本 |
| --- | --- |
| DELETE /api/cluster/user/{user_entity} | v17.2.9, v18.2.8, v19.2.5 |
| DELETE /api/nvmeof/subsystem/{nqn}/listener/{host_name}/{traddr} | v19.2.5 |
