# Ceph 20.2.2 Dashboard API - 块存储 RBD

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/block` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/block/image`](#get-api-block-image) - Display Rbd Images
- [`POST /api/block/image`](#post-api-block-image) - POST /api/block/image
- [`GET /api/block/image/clone_format_version`](#get-api-block-image-clone-format-version) - GET /api/block/image/clone_format_version
- [`GET /api/block/image/default_features`](#get-api-block-image-default-features) - GET /api/block/image/default_features
- [`GET /api/block/image/trash`](#get-api-block-image-trash) - Get RBD Trash Details by pool name
- [`POST /api/block/image/trash/purge`](#post-api-block-image-trash-purge) - POST /api/block/image/trash/purge
- [`DELETE /api/block/image/trash/{image_id_spec}`](#delete-api-block-image-trash-image-id-spec) - DELETE /api/block/image/trash/{image_id_spec}
- [`POST /api/block/image/trash/{image_id_spec}/restore`](#post-api-block-image-trash-image-id-spec-restore) - POST /api/block/image/trash/{image_id_spec}/restore
- [`GET /api/block/image/{image_spec}`](#get-api-block-image-image-spec) - Get Rbd Image Info
- [`PUT /api/block/image/{image_spec}`](#put-api-block-image-image-spec) - PUT /api/block/image/{image_spec}
- [`DELETE /api/block/image/{image_spec}`](#delete-api-block-image-image-spec) - DELETE /api/block/image/{image_spec}
- [`POST /api/block/image/{image_spec}/copy`](#post-api-block-image-image-spec-copy) - POST /api/block/image/{image_spec}/copy
- [`POST /api/block/image/{image_spec}/flatten`](#post-api-block-image-image-spec-flatten) - POST /api/block/image/{image_spec}/flatten
- [`POST /api/block/image/{image_spec}/move_trash`](#post-api-block-image-image-spec-move-trash) - POST /api/block/image/{image_spec}/move_trash
- [`POST /api/block/image/{image_spec}/snap`](#post-api-block-image-image-spec-snap) - POST /api/block/image/{image_spec}/snap
- [`PUT /api/block/image/{image_spec}/snap/{snapshot_name}`](#put-api-block-image-image-spec-snap-snapshot-name) - PUT /api/block/image/{image_spec}/snap/{snapshot_name}
- [`DELETE /api/block/image/{image_spec}/snap/{snapshot_name}`](#delete-api-block-image-image-spec-snap-snapshot-name) - DELETE /api/block/image/{image_spec}/snap/{snapshot_name}
- [`POST /api/block/image/{image_spec}/snap/{snapshot_name}/clone`](#post-api-block-image-image-spec-snap-snapshot-name-clone) - POST /api/block/image/{image_spec}/snap/{snapshot_name}/clone
- [`POST /api/block/image/{image_spec}/snap/{snapshot_name}/rollback`](#post-api-block-image-image-spec-snap-snapshot-name-rollback) - POST /api/block/image/{image_spec}/snap/{snapshot_name}/rollback
- [`GET /api/block/mirroring/pool/{pool_name}`](#get-api-block-mirroring-pool-pool-name) - Display Rbd Mirroring Summary
- [`PUT /api/block/mirroring/pool/{pool_name}`](#put-api-block-mirroring-pool-pool-name) - PUT /api/block/mirroring/pool/{pool_name}
- [`POST /api/block/mirroring/pool/{pool_name}/bootstrap/peer`](#post-api-block-mirroring-pool-pool-name-bootstrap-peer) - POST /api/block/mirroring/pool/{pool_name}/bootstrap/peer
- [`POST /api/block/mirroring/pool/{pool_name}/bootstrap/token`](#post-api-block-mirroring-pool-pool-name-bootstrap-token) - POST /api/block/mirroring/pool/{pool_name}/bootstrap/token
- [`GET /api/block/mirroring/pool/{pool_name}/peer`](#get-api-block-mirroring-pool-pool-name-peer) - GET /api/block/mirroring/pool/{pool_name}/peer
- [`POST /api/block/mirroring/pool/{pool_name}/peer`](#post-api-block-mirroring-pool-pool-name-peer) - POST /api/block/mirroring/pool/{pool_name}/peer
- [`GET /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}`](#get-api-block-mirroring-pool-pool-name-peer-peer-uuid) - GET /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}
- [`PUT /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}`](#put-api-block-mirroring-pool-pool-name-peer-peer-uuid) - PUT /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}
- [`DELETE /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}`](#delete-api-block-mirroring-pool-pool-name-peer-peer-uuid) - DELETE /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}
- [`GET /api/block/mirroring/site_name`](#get-api-block-mirroring-site-name) - Display Rbd Mirroring sitename
- [`PUT /api/block/mirroring/site_name`](#put-api-block-mirroring-site-name) - PUT /api/block/mirroring/site_name
- [`GET /api/block/mirroring/summary`](#get-api-block-mirroring-summary) - Display Rbd Mirroring Summary
- [`GET /api/block/mirroring/{pool_name}/{image_name}/summary`](#get-api-block-mirroring-pool-name-image-name-summary) - GET /api/block/mirroring/{pool_name}/{image_name}/summary
- [`GET /api/block/pool/{pool_name}/group`](#get-api-block-pool-pool-name-group) - List groups by pool name
- [`POST /api/block/pool/{pool_name}/group`](#post-api-block-pool-pool-name-group) - Create a group
- [`GET /api/block/pool/{pool_name}/group/{group_name}`](#get-api-block-pool-pool-name-group-group-name) - Get the list of images in a group
- [`PUT /api/block/pool/{pool_name}/group/{group_name}`](#put-api-block-pool-pool-name-group-group-name) - Update a group (rename)
- [`DELETE /api/block/pool/{pool_name}/group/{group_name}`](#delete-api-block-pool-pool-name-group-group-name) - Delete a group
- [`POST /api/block/pool/{pool_name}/group/{group_name}/image`](#post-api-block-pool-pool-name-group-group-name-image) - Add image to a group
- [`DELETE /api/block/pool/{pool_name}/group/{group_name}/image`](#delete-api-block-pool-pool-name-group-group-name-image) - Remove image from a group
- [`GET /api/block/pool/{pool_name}/group/{group_name}/snap`](#get-api-block-pool-pool-name-group-group-name-snap) - List group snapshots
- [`POST /api/block/pool/{pool_name}/group/{group_name}/snap`](#post-api-block-pool-pool-name-group-group-name-snap) - Create a group snapshot
- [`GET /api/block/pool/{pool_name}/group/{group_name}/snap/{snapshot_name}`](#get-api-block-pool-pool-name-group-group-name-snap-snapshot-name) - Get group snapshot information
- [`PUT /api/block/pool/{pool_name}/group/{group_name}/snap/{snapshot_name}`](#put-api-block-pool-pool-name-group-group-name-snap-snapshot-name) - Update a group snapshot
- [`DELETE /api/block/pool/{pool_name}/group/{group_name}/snap/{snapshot_name}`](#delete-api-block-pool-pool-name-group-group-name-snap-snapshot-name) - Delete a group snapshot
- [`POST /api/block/pool/{pool_name}/group/{group_name}/snap/{snapshot_name}/rollback`](#post-api-block-pool-pool-name-group-group-name-snap-snapshot-name-rollback) - Rollback group to snapshot
- [`GET /api/block/pool/{pool_name}/namespace`](#get-api-block-pool-pool-name-namespace) - GET /api/block/pool/{pool_name}/namespace
- [`POST /api/block/pool/{pool_name}/namespace`](#post-api-block-pool-pool-name-namespace) - POST /api/block/pool/{pool_name}/namespace
- [`DELETE /api/block/pool/{pool_name}/namespace/{namespace}`](#delete-api-block-pool-pool-name-namespace-namespace) - DELETE /api/block/pool/{pool_name}/namespace/{namespace}

## 接口详情

### `GET /api/block/image`

- 摘要：Display Rbd Images
- Tags：`Rbd`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | query | 否 | string |  | Pool Name |
| offset | query | 否 | integer | 0 | offset |
| limit | query | 否 | integer | 5 | limit |
| search | query | 否 | string | "" |  |
| sort | query | 否 | string | "" |  |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v2.0+json`
- Schema: array<object>

```yaml
items:
  properties:
    pool_name:
      description: pool name
      type: string
    value:
      description: ''
      items:
        type: string
      type: array
  type: object
required:
- value
- pool_name
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


### `POST /api/block/image`

- 摘要：POST /api/block/image
- Tags：`Rbd`
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
- Schema: object; fields: configuration, data_pool, features, metadata, mirror_mode, name, namespace, obj_size, pool_name, schedule_interval, size, stripe_count, stripe_unit; required: name, pool_name, size

```yaml
properties:
  configuration:
    type: string
  data_pool:
    type: string
  features:
    type: string
  metadata:
    type: string
  mirror_mode:
    type: string
  name:
    type: string
  namespace:
    type: string
  obj_size:
    type: integer
  pool_name:
    type: string
  schedule_interval:
    default: ''
    type: string
  size:
    type: integer
  stripe_count:
    type: integer
  stripe_unit:
    type: string
required:
- name
- pool_name
- size
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


### `GET /api/block/image/clone_format_version`

- 摘要：GET /api/block/image/clone_format_version
- Tags：`Rbd`
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


### `GET /api/block/image/default_features`

- 摘要：GET /api/block/image/default_features
- Tags：`Rbd`
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


### `GET /api/block/image/trash`

- 摘要：Get RBD Trash Details by pool name
- Tags：`RbdTrash`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | query | 否 | string |  | Name of the pool |


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
    pool_name:
      description: pool name
      type: string
    status:
      description: ''
      type: integer
    value:
      description: ''
      items:
        type: string
      type: array
  type: object
required:
- status
- value
- pool_name
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


### `POST /api/block/image/trash/purge`

- 摘要：POST /api/block/image/trash/purge
- Tags：`RbdTrash`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | query | 否 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: pool_name; required: 无

```yaml
properties:
  pool_name:
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


### `DELETE /api/block/image/trash/{image_id_spec}`

- 摘要：DELETE /api/block/image/trash/{image_id_spec}
- Tags：`RbdTrash`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_id_spec | path | 是 | string |  |  |
| force | query | 否 | boolean | false |  |


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


### `POST /api/block/image/trash/{image_id_spec}/restore`

- 摘要：POST /api/block/image/trash/{image_id_spec}/restore
- Tags：`RbdTrash`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_id_spec | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: new_image_name; required: new_image_name

```yaml
properties:
  new_image_name:
    type: string
required:
- new_image_name
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


### `GET /api/block/image/{image_spec}`

- 摘要：Get Rbd Image Info
- Tags：`Rbd`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_spec | path | 是 | string |  | URL-encoded "pool/rbd_name". e.g. "rbd%2Ffoo" |
| omit_usage | query | 否 | boolean | false | When true, usage information is not returned |


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
    pool_name:
      description: pool name
      type: string
    value:
      description: ''
      items:
        type: string
      type: array
  type: object
required:
- value
- pool_name
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


### `PUT /api/block/image/{image_spec}`

- 摘要：PUT /api/block/image/{image_spec}
- Tags：`Rbd`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_spec | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: configuration, enable_mirror, features, force, image_mirror_mode, metadata, mirror_mode, name, primary, remove_scheduling, resync, schedule_interval, size; required: 无

```yaml
properties:
  configuration:
    type: string
  enable_mirror:
    type: string
  features:
    type: string
  force:
    default: false
    type: boolean
  image_mirror_mode:
    type: string
  metadata:
    type: string
  mirror_mode:
    type: string
  name:
    type: string
  primary:
    type: string
  remove_scheduling:
    default: false
    type: boolean
  resync:
    default: false
    type: boolean
  schedule_interval:
    default: ''
    type: string
  size:
    type: integer
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


### `DELETE /api/block/image/{image_spec}`

- 摘要：DELETE /api/block/image/{image_spec}
- Tags：`Rbd`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_spec | path | 是 | string |  |  |


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


### `POST /api/block/image/{image_spec}/copy`

- 摘要：POST /api/block/image/{image_spec}/copy
- Tags：`Rbd`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_spec | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: configuration, data_pool, dest_image_name, dest_namespace, dest_pool_name, features, metadata, obj_size, snapshot_name, stripe_count, stripe_unit; required: dest_pool_name, dest_namespace, dest_image_name

```yaml
properties:
  configuration:
    type: string
  data_pool:
    type: string
  dest_image_name:
    type: string
  dest_namespace:
    type: string
  dest_pool_name:
    type: string
  features:
    type: string
  metadata:
    type: string
  obj_size:
    type: integer
  snapshot_name:
    type: string
  stripe_count:
    type: integer
  stripe_unit:
    type: string
required:
- dest_pool_name
- dest_namespace
- dest_image_name
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


### `POST /api/block/image/{image_spec}/flatten`

- 摘要：POST /api/block/image/{image_spec}/flatten
- Tags：`Rbd`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_spec | path | 是 | string |  |  |


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


### `POST /api/block/image/{image_spec}/move_trash`

- 摘要：POST /api/block/image/{image_spec}/move_trash
- Tags：`Rbd`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_spec | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: delay; required: 无

```yaml
properties:
  delay:
    default: 0
    type: integer
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


### `POST /api/block/image/{image_spec}/snap`

- 摘要：POST /api/block/image/{image_spec}/snap
- Tags：`RbdSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_spec | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: mirrorImageSnapshot, snapshot_name; required: snapshot_name, mirrorImageSnapshot

```yaml
properties:
  mirrorImageSnapshot:
    type: string
  snapshot_name:
    type: string
required:
- snapshot_name
- mirrorImageSnapshot
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


### `PUT /api/block/image/{image_spec}/snap/{snapshot_name}`

- 摘要：PUT /api/block/image/{image_spec}/snap/{snapshot_name}
- Tags：`RbdSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_spec | path | 是 | string |  |  |
| snapshot_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: is_protected, new_snap_name; required: 无

```yaml
properties:
  is_protected:
    type: boolean
  new_snap_name:
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


### `DELETE /api/block/image/{image_spec}/snap/{snapshot_name}`

- 摘要：DELETE /api/block/image/{image_spec}/snap/{snapshot_name}
- Tags：`RbdSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_spec | path | 是 | string |  |  |
| snapshot_name | path | 是 | string |  |  |


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


### `POST /api/block/image/{image_spec}/snap/{snapshot_name}/clone`

- 摘要：POST /api/block/image/{image_spec}/snap/{snapshot_name}/clone
- Tags：`RbdSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_spec | path | 是 | string |  |  |
| snapshot_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: child_image_name, child_namespace, child_pool_name, configuration, data_pool, features, metadata, obj_size, stripe_count, stripe_unit; required: child_pool_name, child_image_name

```yaml
properties:
  child_image_name:
    type: string
  child_namespace:
    type: string
  child_pool_name:
    type: string
  configuration:
    type: string
  data_pool:
    type: string
  features:
    type: string
  metadata:
    type: string
  obj_size:
    type: integer
  stripe_count:
    type: integer
  stripe_unit:
    type: string
required:
- child_pool_name
- child_image_name
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


### `POST /api/block/image/{image_spec}/snap/{snapshot_name}/rollback`

- 摘要：POST /api/block/image/{image_spec}/snap/{snapshot_name}/rollback
- Tags：`RbdSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| image_spec | path | 是 | string |  |  |
| snapshot_name | path | 是 | string |  |  |


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


### `GET /api/block/mirroring/pool/{pool_name}`

- 摘要：Display Rbd Mirroring Summary
- Tags：`RbdMirroringPoolMode`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Pool Name |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: mirror_mode; required: mirror_mode

```yaml
properties:
  mirror_mode:
    description: Mirror Mode
    type: string
required:
- mirror_mode
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


### `PUT /api/block/mirroring/pool/{pool_name}`

- 摘要：PUT /api/block/mirroring/pool/{pool_name}
- Tags：`RbdMirroringPoolMode`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: mirror_mode; required: 无

```yaml
properties:
  mirror_mode:
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


### `POST /api/block/mirroring/pool/{pool_name}/bootstrap/peer`

- 摘要：POST /api/block/mirroring/pool/{pool_name}/bootstrap/peer
- Tags：`RbdMirroringPoolBootstrap`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: direction, token; required: direction, token

```yaml
properties:
  direction:
    type: string
  token:
    type: string
required:
- direction
- token
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


### `POST /api/block/mirroring/pool/{pool_name}/bootstrap/token`

- 摘要：POST /api/block/mirroring/pool/{pool_name}/bootstrap/token
- Tags：`RbdMirroringPoolBootstrap`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |


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


### `GET /api/block/mirroring/pool/{pool_name}/peer`

- 摘要：GET /api/block/mirroring/pool/{pool_name}/peer
- Tags：`RbdMirroringPoolPeer`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |


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


### `POST /api/block/mirroring/pool/{pool_name}/peer`

- 摘要：POST /api/block/mirroring/pool/{pool_name}/peer
- Tags：`RbdMirroringPoolPeer`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: client_id, cluster_name, key, mon_host; required: cluster_name, client_id

```yaml
properties:
  client_id:
    type: string
  cluster_name:
    type: string
  key:
    type: string
  mon_host:
    type: string
required:
- cluster_name
- client_id
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


### `GET /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}`

- 摘要：GET /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}
- Tags：`RbdMirroringPoolPeer`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |
| peer_uuid | path | 是 | string |  |  |


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


### `PUT /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}`

- 摘要：PUT /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}
- Tags：`RbdMirroringPoolPeer`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |
| peer_uuid | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: client_id, cluster_name, key, mon_host; required: 无

```yaml
properties:
  client_id:
    type: string
  cluster_name:
    type: string
  key:
    type: string
  mon_host:
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


### `DELETE /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}`

- 摘要：DELETE /api/block/mirroring/pool/{pool_name}/peer/{peer_uuid}
- Tags：`RbdMirroringPoolPeer`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |
| peer_uuid | path | 是 | string |  |  |


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


### `GET /api/block/mirroring/site_name`

- 摘要：Display Rbd Mirroring sitename
- Tags：`RbdMirroring`
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
- Schema: object; fields: site_name; required: site_name

```yaml
properties:
  site_name:
    description: Site Name
    type: string
required:
- site_name
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


### `PUT /api/block/mirroring/site_name`

- 摘要：PUT /api/block/mirroring/site_name
- Tags：`RbdMirroring`
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
- Schema: object; fields: site_name; required: site_name

```yaml
properties:
  site_name:
    type: string
required:
- site_name
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


### `GET /api/block/mirroring/summary`

- 摘要：Display Rbd Mirroring Summary
- Tags：`RbdMirroringSummary`
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
- Schema: object; fields: content_data, site_name, status; required: site_name, status, content_data

```yaml
properties:
  content_data:
    description: ''
    properties:
      daemons:
        description: ''
        items:
          type: string
        type: array
      image_error:
        description: ''
        items:
          type: string
        type: array
      image_ready:
        description: ''
        items:
          type: string
        type: array
      image_syncing:
        description: ''
        items:
          type: string
        type: array
      pools:
        description: Pools
        items:
          properties:
            health:
              description: pool health
              type: string
            health_color:
              description: ''
              type: string
            mirror_mode:
              description: status
              type: string
            name:
              description: Pool name
              type: string
            peer_uuids:
              description: ''
              items:
                type: string
              type: array
          required:
          - name
          - health_color
          - health
          - mirror_mode
          - peer_uuids
          type: object
        type: array
    required:
    - daemons
    - pools
    - image_error
    - image_syncing
    - image_ready
    type: object
  site_name:
    description: site name
    type: string
  status:
    description: ''
    type: integer
required:
- site_name
- status
- content_data
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


### `GET /api/block/mirroring/{pool_name}/{image_name}/summary`

- 摘要：GET /api/block/mirroring/{pool_name}/{image_name}/summary
- Tags：`RbdMirroringSummary`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |
| image_name | path | 是 | string |  |  |


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


### `GET /api/block/pool/{pool_name}/group`

- 摘要：List groups by pool name
- Tags：`RbdGroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| namespace | query | 否 | string |  |  |


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
    group:
      description: group name
      type: string
    num_images:
      description: ''
      type: integer
  type: object
required:
- group
- num_images
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


### `POST /api/block/pool/{pool_name}/group`

- 摘要：Create a group
- Tags：`RbdGroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: name, namespace; required: name

```yaml
properties:
  name:
    description: Name of the group
    type: string
  namespace:
    type: string
required:
- name
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


### `GET /api/block/pool/{pool_name}/group/{group_name}`

- 摘要：Get the list of images in a group
- Tags：`RbdGroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| group_name | path | 是 | string |  | Name of the group |
| namespace | query | 否 | string |  |  |


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
    group:
      description: group name
      type: string
    images:
      description: ''
      items:
        type: string
      type: array
  type: object
required:
- group
- images
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


### `PUT /api/block/pool/{pool_name}/group/{group_name}`

- 摘要：Update a group (rename)
- Tags：`RbdGroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| group_name | path | 是 | string |  | Name of the group |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: namespace, new_name; required: new_name

```yaml
properties:
  namespace:
    type: string
  new_name:
    description: New name for the group
    type: string
required:
- new_name
type: object
```


#### 返回消息

#### `200`

Resource updated.

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object

```yaml
properties: {}
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


### `DELETE /api/block/pool/{pool_name}/group/{group_name}`

- 摘要：Delete a group
- Tags：`RbdGroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| group_name | path | 是 | string |  | Name of the group |
| namespace | query | 否 | string |  |  |


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


### `POST /api/block/pool/{pool_name}/group/{group_name}/image`

- 摘要：Add image to a group
- Tags：`RbdGroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| group_name | path | 是 | string |  | Name of the group |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: image_name, namespace; required: image_name

```yaml
properties:
  image_name:
    description: Name of the image
    type: string
  namespace:
    type: string
required:
- image_name
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


### `DELETE /api/block/pool/{pool_name}/group/{group_name}/image`

- 摘要：Remove image from a group
- Tags：`RbdGroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| group_name | path | 是 | string |  | Name of the group |
| image_name | query | 是 | string |  | Name of the image |
| namespace | query | 否 | string |  |  |


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


### `GET /api/block/pool/{pool_name}/group/{group_name}/snap`

- 摘要：List group snapshots
- Tags：`RbdGroupSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| group_name | path | 是 | string |  | Name of the group |
| namespace | query | 否 | string |  |  |


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
    id:
      description: snapshot id
      type: string
    name:
      description: snapshot name
      type: string
    namespace_type:
      description: namespace type
      type: integer
    state:
      description: snapshot state
      type: integer
  type: object
required:
- id
- name
- state
- namespace_type
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


### `POST /api/block/pool/{pool_name}/group/{group_name}/snap`

- 摘要：Create a group snapshot
- Tags：`RbdGroupSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| group_name | path | 是 | string |  | Name of the group |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: flags, namespace, snapshot_name; required: snapshot_name

```yaml
properties:
  flags:
    default: 0
    description: Snapshot creation flags
    type: integer
  namespace:
    type: string
  snapshot_name:
    description: Name of the snapshot
    type: string
required:
- snapshot_name
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


### `GET /api/block/pool/{pool_name}/group/{group_name}/snap/{snapshot_name}`

- 摘要：Get group snapshot information
- Tags：`RbdGroupSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| group_name | path | 是 | string |  | Name of the group |
| snapshot_name | path | 是 | string |  | Name of the snapshot |
| namespace | query | 否 | string |  |  |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: id, image_snap_name, image_snaps, name, namespace_type, state; required: id, name, state, namespace_type, image_snap_name, image_snaps

```yaml
properties:
  id:
    description: snapshot id
    type: string
  image_snap_name:
    description: image snapshot name
    type: string
  image_snaps:
    description: image snapshots
    items:
      type: object
    type: array
  name:
    description: snapshot name
    type: string
  namespace_type:
    description: namespace type
    type: integer
  state:
    description: snapshot state
    type: integer
required:
- id
- name
- state
- namespace_type
- image_snap_name
- image_snaps
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


### `PUT /api/block/pool/{pool_name}/group/{group_name}/snap/{snapshot_name}`

- 摘要：Update a group snapshot
- Tags：`RbdGroupSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| group_name | path | 是 | string |  | Name of the group |
| snapshot_name | path | 是 | string |  | Current name of the snapshot |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: namespace, new_snap_name; required: 无

```yaml
properties:
  namespace:
    type: string
  new_snap_name:
    description: New name for the snapshot
    type: string
type: object
```


#### 返回消息

#### `200`

Resource updated.

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object

```yaml
properties: {}
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


### `DELETE /api/block/pool/{pool_name}/group/{group_name}/snap/{snapshot_name}`

- 摘要：Delete a group snapshot
- Tags：`RbdGroupSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| group_name | path | 是 | string |  | Name of the group |
| snapshot_name | path | 是 | string |  | Name of the snapshot |
| namespace | query | 否 | string |  |  |


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


### `POST /api/block/pool/{pool_name}/group/{group_name}/snap/{snapshot_name}/rollback`

- 摘要：Rollback group to snapshot
- Tags：`RbdGroupSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  | Name of the pool |
| group_name | path | 是 | string |  | Name of the group |
| snapshot_name | path | 是 | string |  | Name of the snapshot |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: namespace; required: 无

```yaml
properties:
  namespace:
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


### `GET /api/block/pool/{pool_name}/namespace`

- 摘要：GET /api/block/pool/{pool_name}/namespace
- Tags：`RbdNamespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |


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


### `POST /api/block/pool/{pool_name}/namespace`

- 摘要：POST /api/block/pool/{pool_name}/namespace
- Tags：`RbdNamespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: namespace; required: namespace

```yaml
properties:
  namespace:
    type: string
required:
- namespace
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


### `DELETE /api/block/pool/{pool_name}/namespace/{namespace}`

- 摘要：DELETE /api/block/pool/{pool_name}/namespace/{namespace}
- Tags：`RbdNamespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |
| namespace | path | 是 | string |  |  |


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

