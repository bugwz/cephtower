# Ceph 20.2.2 Dashboard API - 反馈

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/feedback` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/feedback`](#get-api-feedback) - GET /api/feedback
- [`POST /api/feedback`](#post-api-feedback) - POST /api/feedback
- [`GET /api/feedback/api_key`](#get-api-feedback-api-key) - GET /api/feedback/api_key
- [`POST /api/feedback/api_key`](#post-api-feedback-api-key) - POST /api/feedback/api_key
- [`DELETE /api/feedback/api_key`](#delete-api-feedback-api-key) - DELETE /api/feedback/api_key

## 接口详情

### `GET /api/feedback`

- 摘要：GET /api/feedback
- Tags：`Report`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v17.2.9
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

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


### `POST /api/feedback`

- 摘要：POST /api/feedback
- Tags：`Report`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v17.2.9
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: api_key, description, project, subject, tracker; required: project, tracker, subject, description

```yaml
properties:
  api_key:
    type: string
  description:
    type: string
  project:
    type: string
  subject:
    type: string
  tracker:
    type: string
required:
- project
- tracker
- subject
- description
type: object
```


#### 返回消息

#### `201`

Resource created.

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


### `GET /api/feedback/api_key`

- 摘要：GET /api/feedback/api_key
- Tags：`Report`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v17.2.9
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

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


### `POST /api/feedback/api_key`

- 摘要：POST /api/feedback/api_key
- Tags：`Report`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v17.2.9
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: api_key; required: api_key

```yaml
properties:
  api_key:
    type: string
required:
- api_key
type: object
```


#### 返回消息

#### `201`

Resource created.

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


### `DELETE /api/feedback/api_key`

- 摘要：DELETE /api/feedback/api_key
- Tags：`Report`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v17.2.9
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

无请求体。


#### 返回消息

#### `202`

Operation is still executing. Please check the task queue.

- Content-Type: `application/vnd.ceph.api.v0.1+json`
- Schema: object

```yaml
type: object
```

#### `204`

Resource deleted.

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

