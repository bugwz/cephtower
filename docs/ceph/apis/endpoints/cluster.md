# Ceph 20.2.2 Dashboard API - 集群

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/cluster` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/cluster`](#get-api-cluster) - Get the cluster status
- [`PUT /api/cluster`](#put-api-cluster) - Update the cluster status
- [`GET /api/cluster/upgrade`](#get-api-cluster-upgrade) - Get the available versions to upgrade
- [`PUT /api/cluster/upgrade/pause`](#put-api-cluster-upgrade-pause) - Pause the cluster upgrade
- [`PUT /api/cluster/upgrade/resume`](#put-api-cluster-upgrade-resume) - Resume the cluster upgrade
- [`POST /api/cluster/upgrade/start`](#post-api-cluster-upgrade-start) - Start the cluster upgrade
- [`GET /api/cluster/upgrade/status`](#get-api-cluster-upgrade-status) - Get the cluster upgrade status
- [`PUT /api/cluster/upgrade/stop`](#put-api-cluster-upgrade-stop) - Stop the cluster upgrade
- [`GET /api/cluster/user`](#get-api-cluster-user) - Get list of ceph users
- [`POST /api/cluster/user`](#post-api-cluster-user) - Create Ceph User
- [`PUT /api/cluster/user`](#put-api-cluster-user) - Edit Ceph User Capabilities
- [`POST /api/cluster/user/export`](#post-api-cluster-user-export) - Export Ceph Users
- [`DELETE /api/cluster/user/{user_entities}`](#delete-api-cluster-user-user-entities) - Delete one or more Ceph Users

## 接口详情

### `GET /api/cluster`

- 摘要：Get the cluster status
- Tags：`Cluster`
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


### `PUT /api/cluster`

- 摘要：Update the cluster status
- Tags：`Cluster`
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
- Schema: object; fields: status; required: status

```yaml
properties:
  status:
    description: Cluster Status
    type: string
required:
- status
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


### `GET /api/cluster/upgrade`

- 摘要：Get the available versions to upgrade
- Tags：`Upgrade`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| tags | query | 否 | boolean | false | Show all image tags |
| image | query | 否 | string |  | Ceph Image |
| show_all_versions | query | 否 | boolean | false | Show all available versions |


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


### `PUT /api/cluster/upgrade/pause`

- 摘要：Pause the cluster upgrade
- Tags：`Upgrade`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

无请求体。


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


### `PUT /api/cluster/upgrade/resume`

- 摘要：Resume the cluster upgrade
- Tags：`Upgrade`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

无请求体。


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


### `POST /api/cluster/upgrade/start`

- 摘要：Start the cluster upgrade
- Tags：`Upgrade`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: daemon_types, host_placement, image, limit, services, version; required: 无

```yaml
properties:
  daemon_types:
    type: string
  host_placement:
    type: string
  image:
    type: string
  limit:
    type: string
  services:
    type: string
  version:
    type: string
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


### `GET /api/cluster/upgrade/status`

- 摘要：Get the cluster upgrade status
- Tags：`Upgrade`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
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


### `PUT /api/cluster/upgrade/stop`

- 摘要：Stop the cluster upgrade
- Tags：`Upgrade`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

无请求体。


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


### `GET /api/cluster/user`

- 摘要：Get list of ceph users
- Tags：`Cluster`
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


### `POST /api/cluster/user`

- 摘要：Create Ceph User
- Tags：`Cluster`
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
- Schema: object; fields: capabilities, import_data, user_entity; required: 无

```yaml
properties:
  capabilities:
    description: List of capabilities to add to user_entity
    items:
      properties:
        cap:
          description: Capability to add; eg. allow *
          type: string
        entity:
          description: Entity to add
          type: string
      required:
      - entity
      - cap
      type: object
    type: array
  import_data:
    default: ''
    type: string
  user_entity:
    default: ''
    description: Entity to add
    type: string
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


### `PUT /api/cluster/user`

- 摘要：Edit Ceph User Capabilities
- Tags：`Cluster`
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
- Schema: object; fields: capabilities, user_entity; required: 无

```yaml
properties:
  capabilities:
    description: List of updated capabilities to user_entity
    items:
      properties:
        cap:
          description: Capability to edit; eg. allow *
          type: string
        entity:
          description: Entity to edit
          type: string
      required:
      - entity
      - cap
      type: object
    type: array
  user_entity:
    default: ''
    description: Entity to edit
    type: string
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


### `POST /api/cluster/user/export`

- 摘要：Export Ceph Users
- Tags：`Cluster`
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
- Schema: object; fields: entities; required: entities

```yaml
properties:
  entities:
    description: List of entities to export
    items:
      type: string
    type: array
required:
- entities
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


### `DELETE /api/cluster/user/{user_entities}`

- 摘要：Delete one or more Ceph Users
- Tags：`Cluster`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| user_entities | path | 是 | string |  |  |


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

