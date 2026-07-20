# Ceph 20.2.2 Dashboard API - 概览

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/summary` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/summary`](#get-api-summary) - Display Summary

## 接口详情

### `GET /api/summary`

- 摘要：Display Summary
- Tags：`Summary`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: executing_tasks, finished_tasks, have_mon_connection, health_status, mgr_host, mgr_id, rbd_mirroring, version; required: health_status, mgr_id, mgr_host, have_mon_connection, executing_tasks, finished_tasks, version, rbd_mirroring

```yaml
properties:
  executing_tasks:
    description: ''
    items:
      type: string
    type: array
  finished_tasks:
    description: ''
    items:
      properties:
        begin_time:
          description: ''
          type: string
        duration:
          description: ''
          type: integer
        end_time:
          description: ''
          type: string
        exception:
          description: ''
          type: string
        metadata:
          description: ''
          properties:
            pool:
              description: ''
              type: integer
          required:
          - pool
          type: object
        name:
          description: ''
          type: string
        progress:
          description: ''
          type: integer
        ret_value:
          description: ''
          type: string
        success:
          description: ''
          type: boolean
      required:
      - name
      - metadata
      - begin_time
      - end_time
      - duration
      - progress
      - success
      - ret_value
      - exception
      type: object
    type: array
  have_mon_connection:
    description: ''
    type: string
  health_status:
    description: ''
    type: string
  mgr_host:
    description: ''
    type: string
  mgr_id:
    description: ''
    type: string
  rbd_mirroring:
    description: ''
    properties:
      errors:
        description: ''
        type: integer
      warnings:
        description: ''
        type: integer
    required:
    - warnings
    - errors
    type: object
  version:
    description: ''
    type: string
required:
- health_status
- mgr_id
- mgr_host
- have_mon_connection
- executing_tasks
- finished_tasks
- version
- rbd_mirroring
type: object
```

#### `400`

Operation exception. Please check the response body for details.

无响应体 schema。

#### `401`

Unauthenticated access. Please login first.

无响应体 schema。

#### `403`

Unauthorized access. Please check your permissions.

无响应体 schema。

#### `500`

Unexpected error. Please check the response body for the stack trace.

无响应体 schema。

