# Ceph 20.2.2 Dashboard API - 存储池

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/pool` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/pool`](#get-api-pool) - Display Pool List
- [`POST /api/pool`](#post-api-pool) - POST /api/pool
- [`GET /api/pool/{pool_name}`](#get-api-pool-pool-name) - GET /api/pool/{pool_name}
- [`PUT /api/pool/{pool_name}`](#put-api-pool-pool-name) - PUT /api/pool/{pool_name}
- [`DELETE /api/pool/{pool_name}`](#delete-api-pool-pool-name) - DELETE /api/pool/{pool_name}
- [`GET /api/pool/{pool_name}/configuration`](#get-api-pool-pool-name-configuration) - GET /api/pool/{pool_name}/configuration

## 接口详情

### `GET /api/pool`

- 摘要：Display Pool List
- Tags：`Pool`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| attrs | query | 否 | string |  | Pool Attributes |
| stats | query | 否 | boolean | false | Pool Stats |


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
    application_metadata:
      description: ''
      items:
        type: string
      type: array
    auid:
      description: ''
      type: integer
    cache_min_evict_age:
      description: ''
      type: integer
    cache_min_flush_age:
      description: ''
      type: integer
    cache_mode:
      description: ''
      type: string
    cache_target_dirty_high_ratio_micro:
      description: ''
      type: integer
    cache_target_dirty_ratio_micro:
      description: ''
      type: integer
    cache_target_full_ratio_micro:
      description: ''
      type: integer
    create_time:
      description: ''
      type: string
    crush_rule:
      description: ''
      type: string
    erasure_code_profile:
      description: ''
      type: string
    expected_num_objects:
      description: ''
      type: integer
    fast_read:
      description: ''
      type: boolean
    flags:
      description: ''
      type: integer
    flags_names:
      description: flags name
      type: string
    grade_table:
      description: ''
      items:
        type: string
      type: array
    hit_set_count:
      description: ''
      type: integer
    hit_set_grade_decay_rate:
      description: ''
      type: integer
    hit_set_params:
      description: ''
      properties:
        type:
          description: ''
          type: string
      required:
      - type
      type: object
    hit_set_period:
      description: ''
      type: integer
    hit_set_search_last_n:
      description: ''
      type: integer
    last_change:
      description: ''
      type: string
    last_force_op_resend:
      description: ''
      type: string
    last_force_op_resend_preluminous:
      description: ''
      type: string
    last_force_op_resend_prenautilus:
      description: ''
      type: string
    last_pg_merge_meta:
      description: ''
      properties:
        last_epoch_clean:
          description: ''
          type: integer
        last_epoch_started:
          description: ''
          type: integer
        ready_epoch:
          description: ''
          type: integer
        source_pgid:
          description: ''
          type: string
        source_version:
          description: ''
          type: string
        target_version:
          description: ''
          type: string
      required:
      - ready_epoch
      - last_epoch_started
      - last_epoch_clean
      - source_pgid
      - source_version
      - target_version
      type: object
    min_read_recency_for_promote:
      description: ''
      type: integer
    min_size:
      description: ''
      type: integer
    min_write_recency_for_promote:
      description: ''
      type: integer
    object_hash:
      description: ''
      type: integer
    options:
      description: ''
      properties:
        pg_num_max:
          description: ''
          type: integer
        pg_num_min:
          description: ''
          type: integer
      required:
      - pg_num_min
      - pg_num_max
      type: object
    pg_autoscale_mode:
      description: ''
      type: string
    pg_num:
      description: ''
      type: integer
    pg_num_pending:
      description: ''
      type: integer
    pg_num_target:
      description: ''
      type: integer
    pg_placement_num:
      description: ''
      type: integer
    pg_placement_num_target:
      description: ''
      type: integer
    pool:
      description: pool id
      type: integer
    pool_name:
      description: pool name
      type: string
    pool_snaps:
      description: ''
      items:
        type: string
      type: array
    quota_max_bytes:
      description: ''
      type: integer
    quota_max_objects:
      description: ''
      type: integer
    read_tier:
      description: ''
      type: integer
    removed_snaps:
      description: ''
      items:
        type: string
      type: array
    size:
      description: pool size
      type: integer
    snap_epoch:
      description: ''
      type: integer
    snap_mode:
      description: ''
      type: string
    snap_seq:
      description: ''
      type: integer
    stripe_width:
      description: ''
      type: integer
    target_max_bytes:
      description: ''
      type: integer
    target_max_objects:
      description: ''
      type: integer
    tier_of:
      description: ''
      type: integer
    tiers:
      description: ''
      items:
        type: string
      type: array
    type:
      description: type of pool
      type: string
    use_gmt_hitset:
      description: ''
      type: boolean
    write_tier:
      description: ''
      type: integer
  type: object
required:
- pool
- pool_name
- flags
- flags_names
- type
- size
- min_size
- crush_rule
- object_hash
- pg_autoscale_mode
- pg_num
- pg_placement_num
- pg_placement_num_target
- pg_num_target
- pg_num_pending
- last_pg_merge_meta
- auid
- snap_mode
- snap_seq
- snap_epoch
- pool_snaps
- quota_max_bytes
- quota_max_objects
- tiers
- tier_of
- read_tier
- write_tier
- cache_mode
- target_max_bytes
- target_max_objects
- cache_target_dirty_ratio_micro
- cache_target_dirty_high_ratio_micro
- cache_target_full_ratio_micro
- cache_min_flush_age
- cache_min_evict_age
- erasure_code_profile
- hit_set_params
- hit_set_period
- hit_set_count
- use_gmt_hitset
- min_read_recency_for_promote
- min_write_recency_for_promote
- hit_set_grade_decay_rate
- hit_set_search_last_n
- grade_table
- stripe_width
- expected_num_objects
- fast_read
- options
- application_metadata
- create_time
- last_change
- last_force_op_resend
- last_force_op_resend_prenautilus
- last_force_op_resend_preluminous
- removed_snaps
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


### `POST /api/pool`

- 摘要：POST /api/pool
- Tags：`Pool`
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
- Schema: object; fields: pool; required: 无

```yaml
properties:
  pool:
    default: rbd-mirror
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


### `GET /api/pool/{pool_name}`

- 摘要：GET /api/pool/{pool_name}
- Tags：`Pool`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| pool_name | path | 是 | string |  |  |
| attrs | query | 否 | string |  |  |
| stats | query | 否 | boolean | false |  |


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


### `PUT /api/pool/{pool_name}`

- 摘要：PUT /api/pool/{pool_name}
- Tags：`Pool`
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
- Schema: object; fields: application_metadata, configuration, flags, rbd_mirroring; required: 无

```yaml
properties:
  application_metadata:
    type: string
  configuration:
    type: string
  flags:
    type: string
  rbd_mirroring:
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


### `DELETE /api/pool/{pool_name}`

- 摘要：DELETE /api/pool/{pool_name}
- Tags：`Pool`
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


### `GET /api/pool/{pool_name}/configuration`

- 摘要：GET /api/pool/{pool_name}/configuration
- Tags：`Pool`
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

