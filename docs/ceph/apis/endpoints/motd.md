# Ceph 20.2.2 Dashboard API - MOTD

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/motd` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`POST /api/motd`](#post-api-motd) - POST /api/motd
- [`DELETE /api/motd/clear`](#delete-api-motd-clear) - DELETE /api/motd/clear

## 接口详情

### `POST /api/motd`

- 摘要：POST /api/motd
- Tags：`Motd`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: expires, message, severity; required: severity, expires, message

```yaml
properties:
  expires:
    type: string
  message:
    type: string
  severity:
    type: string
required:
- severity
- expires
- message
type: object
```


#### 返回消息

#### `201`

Resource created.

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object

```yaml
type: object
```

#### `202`

Operation is still executing. Please check the task queue.

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


### `DELETE /api/motd/clear`

- 摘要：DELETE /api/motd/clear
- Tags：`Motd`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

无请求体。


#### 返回消息

#### `202`

Operation is still executing. Please check the task queue.

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object

```yaml
type: object
```

#### `204`

Resource deleted.

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

