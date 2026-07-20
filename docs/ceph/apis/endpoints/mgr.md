# Ceph 20.2.2 Dashboard API - Mgr 模块

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/mgr` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/mgr/module`](#get-api-mgr-module) - List Mgr modules
- [`GET /api/mgr/module/{module_name}`](#get-api-mgr-module-module-name) - GET /api/mgr/module/{module_name}
- [`PUT /api/mgr/module/{module_name}`](#put-api-mgr-module-module-name) - PUT /api/mgr/module/{module_name}
- [`POST /api/mgr/module/{module_name}/disable`](#post-api-mgr-module-module-name-disable) - POST /api/mgr/module/{module_name}/disable
- [`POST /api/mgr/module/{module_name}/enable`](#post-api-mgr-module-module-name-enable) - POST /api/mgr/module/{module_name}/enable
- [`GET /api/mgr/module/{module_name}/options`](#get-api-mgr-module-module-name-options) - GET /api/mgr/module/{module_name}/options

## 接口详情

### `GET /api/mgr/module`

- 摘要：List Mgr modules
- Tags：`MgrModule`
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
- Schema: array<object>

```yaml
items:
  properties:
    always_on:
      description: Is it an always on module?
      type: boolean
    enabled:
      description: Is Module Enabled
      type: boolean
    name:
      description: Module Name
      type: string
    options:
      description: Module Options
      properties:
        Option_name:
          description: Options
          properties:
            default_value:
              description: Default value for the option
              type: integer
            desc:
              description: Description of the option
              type: string
            enum_allowed:
              description: ''
              items:
                type: string
              type: array
            flags:
              description: List of flags associated
              type: integer
            level:
              description: Option level
              type: string
            long_desc:
              description: Elaborated description
              type: string
            max:
              description: Maximum value
              type: string
            min:
              description: Minimum value
              type: string
            name:
              description: Name of the option
              type: string
            see_also:
              description: Related options
              items:
                type: string
              type: array
            tags:
              description: Tags associated with the option
              items:
                type: string
              type: array
            type:
              description: Type of the option
              type: string
          required:
          - name
          - type
          - level
          - flags
          - default_value
          - min
          - max
          - enum_allowed
          - desc
          - long_desc
          - tags
          - see_also
          type: object
      required:
      - Option_name
      type: object
  type: object
required:
- name
- enabled
- always_on
- options
type: array
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


### `GET /api/mgr/module/{module_name}`

- 摘要：GET /api/mgr/module/{module_name}
- Tags：`MgrModule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| module_name | path | 是 | string |  |  |


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


### `PUT /api/mgr/module/{module_name}`

- 摘要：PUT /api/mgr/module/{module_name}
- Tags：`MgrModule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| module_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: config; required: config

```yaml
properties:
  config:
    type: string
required:
- config
type: object
```


#### 返回消息

#### `200`

Resource updated.

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


### `POST /api/mgr/module/{module_name}/disable`

- 摘要：POST /api/mgr/module/{module_name}/disable
- Tags：`MgrModule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| module_name | path | 是 | string |  |  |


#### 请求体

无请求体。


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


### `POST /api/mgr/module/{module_name}/enable`

- 摘要：POST /api/mgr/module/{module_name}/enable
- Tags：`MgrModule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| module_name | path | 是 | string |  |  |


#### 请求体

无请求体。


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


### `GET /api/mgr/module/{module_name}/options`

- 摘要：GET /api/mgr/module/{module_name}/options
- Tags：`MgrModule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| module_name | path | 是 | string |  |  |


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

