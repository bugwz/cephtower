# Ceph 20.2.2 Dashboard API - Daemon

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/daemon` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/daemon`](#get-api-daemon) - GET /api/daemon
- [`PUT /api/daemon/{daemon_name}`](#put-api-daemon-daemon-name) - PUT /api/daemon/{daemon_name}

## 接口详情

### `GET /api/daemon`

- 摘要：GET /api/daemon
- Tags：`Daemon`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| daemon_types | query | 否 | string |  |  |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object

```yaml
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


### `PUT /api/daemon/{daemon_name}`

- 摘要：PUT /api/daemon/{daemon_name}
- Tags：`Daemon`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| daemon_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: action, container_image, force; required: 无

```yaml
properties:
  action:
    default: ''
    type: string
  container_image:
    type: string
  force:
    default: false
    description: When true, force stops/restarts (bypasses ok-to-stop warnings; e.g.
      RGW, NFS, SMB, NVMe-oF, monitoring daemons).
    type: boolean
type: object
```


#### 返回消息

#### `200`

Resource updated.

- Content-Type: `application/vnd.ceph.api.v0.1+json`
- Schema: object

```yaml
type: object
```

#### `202`

Operation is still executing. Please check the task queue.

- Content-Type: `application/vnd.ceph.api.v0.1+json`
- Schema: object

```yaml
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

