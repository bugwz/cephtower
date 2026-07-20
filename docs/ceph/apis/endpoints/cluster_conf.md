# Ceph 20.2.2 Dashboard API - 集群配置

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/cluster_conf` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/cluster_conf`](#get-api-cluster-conf) - GET /api/cluster_conf
- [`POST /api/cluster_conf`](#post-api-cluster-conf) - Create/Update Cluster Configuration
- [`PUT /api/cluster_conf`](#put-api-cluster-conf) - PUT /api/cluster_conf
- [`GET /api/cluster_conf/filter`](#get-api-cluster-conf-filter) - Get Cluster Configuration by name
- [`GET /api/cluster_conf/{name}`](#get-api-cluster-conf-name) - GET /api/cluster_conf/{name}
- [`DELETE /api/cluster_conf/{name}`](#delete-api-cluster-conf-name) - DELETE /api/cluster_conf/{name}

## 接口详情

### `GET /api/cluster_conf`

- 摘要：GET /api/cluster_conf
- Tags：`ClusterConfiguration`
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


### `POST /api/cluster_conf`

- 摘要：Create/Update Cluster Configuration
- Tags：`ClusterConfiguration`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: force_update, name, value; required: name, value

```yaml
properties:
  force_update:
    description: Force update the config option
    type: boolean
  name:
    description: Config option name
    type: string
  value:
    description: Section and Value of the config option
    items:
      properties:
        section:
          description: Section/Client where config needs to be updated
          type: string
        value:
          description: Value of the config option
          type: string
      required:
      - section
      - value
      type: object
    type: array
required:
- name
- value
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


### `PUT /api/cluster_conf`

- 摘要：PUT /api/cluster_conf
- Tags：`ClusterConfiguration`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: options; required: options

```yaml
properties:
  options:
    type: string
required:
- options
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


### `GET /api/cluster_conf/filter`

- 摘要：Get Cluster Configuration by name
- Tags：`ClusterConfiguration`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| names | query | 否 | string |  | Config option names |


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
    can_update_at_runtime:
      description: Check if can update at runtime
      type: boolean
    daemon_default:
      description: Daemon specific default value
      type: string
    default:
      description: Default value for the config option
      type: string
    desc:
      description: Description of the configuration
      type: string
    enum_values:
      description: List of enums allowed
      items:
        type: string
      type: array
    flags:
      description: List of flags associated
      items:
        type: string
      type: array
    level:
      description: Config option level
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
      description: Name of the config option
      type: string
    see_also:
      description: Related config options
      items:
        type: string
      type: array
    services:
      description: Services associated with the config option
      items:
        type: string
      type: array
    tags:
      description: Tags associated with the cluster
      items:
        type: string
      type: array
    type:
      description: Config option type
      type: string
  type: object
required:
- name
- type
- level
- desc
- long_desc
- default
- daemon_default
- tags
- services
- see_also
- enum_values
- min
- max
- can_update_at_runtime
- flags
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


### `GET /api/cluster_conf/{name}`

- 摘要：GET /api/cluster_conf/{name}
- Tags：`ClusterConfiguration`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| name | path | 是 | string |  |  |


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


### `DELETE /api/cluster_conf/{name}`

- 摘要：DELETE /api/cluster_conf/{name}
- Tags：`ClusterConfiguration`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| name | path | 是 | string |  |  |
| section | query | 是 | string |  |  |


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

