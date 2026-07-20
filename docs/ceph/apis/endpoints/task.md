# Ceph 20.2.2 Dashboard API - 任务

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/task` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/task`](#get-api-task) - Display Tasks

## 接口详情

### `GET /api/task`

- 摘要：Display Tasks
- Tags：`Task`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| name | query | 否 | string |  | Task Name |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: executing_tasks, finished_tasks; required: executing_tasks, finished_tasks

```yaml
properties:
  executing_tasks:
    description: ongoing executing tasks
    type: string
  finished_tasks:
    description: ''
    items:
      properties:
        begin_time:
          description: Task begin time
          type: string
        duration:
          description: ''
          type: integer
        end_time:
          description: Task end time
          type: string
        exception:
          description: ''
          type: boolean
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
          description: finished tasks name
          type: string
        progress:
          description: Progress of tasks
          type: integer
        ret_value:
          description: ''
          type: boolean
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
required:
- executing_tasks
- finished_tasks
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

