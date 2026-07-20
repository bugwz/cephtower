# Ceph 20.2.2 Dashboard API - CephFS

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/cephfs` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/cephfs`](#get-api-cephfs) - GET /api/cephfs
- [`POST /api/cephfs`](#post-api-cephfs) - POST /api/cephfs
- [`PUT /api/cephfs/auth`](#put-api-cephfs-auth) - Set Ceph authentication capabilities for the specified user ID in the given path
- [`DELETE /api/cephfs/remove/{name}`](#delete-api-cephfs-remove-name) - Remove CephFS Volume
- [`PUT /api/cephfs/rename`](#put-api-cephfs-rename) - Rename CephFS Volume
- [`POST /api/cephfs/snapshot/schedule`](#post-api-cephfs-snapshot-schedule) - POST /api/cephfs/snapshot/schedule
- [`GET /api/cephfs/snapshot/schedule/{fs}`](#get-api-cephfs-snapshot-schedule-fs) - GET /api/cephfs/snapshot/schedule/{fs}
- [`PUT /api/cephfs/snapshot/schedule/{fs}/{path}`](#put-api-cephfs-snapshot-schedule-fs-path) - PUT /api/cephfs/snapshot/schedule/{fs}/{path}
- [`POST /api/cephfs/snapshot/schedule/{fs}/{path}/activate`](#post-api-cephfs-snapshot-schedule-fs-path-activate) - POST /api/cephfs/snapshot/schedule/{fs}/{path}/activate
- [`POST /api/cephfs/snapshot/schedule/{fs}/{path}/deactivate`](#post-api-cephfs-snapshot-schedule-fs-path-deactivate) - POST /api/cephfs/snapshot/schedule/{fs}/{path}/deactivate
- [`DELETE /api/cephfs/snapshot/schedule/{fs}/{path}/delete_snapshot`](#delete-api-cephfs-snapshot-schedule-fs-path-delete-snapshot) - DELETE /api/cephfs/snapshot/schedule/{fs}/{path}/delete_snapshot
- [`POST /api/cephfs/subvolume`](#post-api-cephfs-subvolume) - POST /api/cephfs/subvolume
- [`POST /api/cephfs/subvolume/group`](#post-api-cephfs-subvolume-group) - POST /api/cephfs/subvolume/group
- [`GET /api/cephfs/subvolume/group/{vol_name}`](#get-api-cephfs-subvolume-group-vol-name) - GET /api/cephfs/subvolume/group/{vol_name}
- [`PUT /api/cephfs/subvolume/group/{vol_name}`](#put-api-cephfs-subvolume-group-vol-name) - PUT /api/cephfs/subvolume/group/{vol_name}
- [`DELETE /api/cephfs/subvolume/group/{vol_name}`](#delete-api-cephfs-subvolume-group-vol-name) - DELETE /api/cephfs/subvolume/group/{vol_name}
- [`GET /api/cephfs/subvolume/group/{vol_name}/info`](#get-api-cephfs-subvolume-group-vol-name-info) - GET /api/cephfs/subvolume/group/{vol_name}/info
- [`POST /api/cephfs/subvolume/snapshot`](#post-api-cephfs-subvolume-snapshot) - POST /api/cephfs/subvolume/snapshot
- [`POST /api/cephfs/subvolume/snapshot/clone`](#post-api-cephfs-subvolume-snapshot-clone) - Create a clone of a subvolume snapshot
- [`GET /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}`](#get-api-cephfs-subvolume-snapshot-vol-name-subvol-name) - GET /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}
- [`DELETE /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}`](#delete-api-cephfs-subvolume-snapshot-vol-name-subvol-name) - DELETE /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}
- [`GET /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}/info`](#get-api-cephfs-subvolume-snapshot-vol-name-subvol-name-info) - GET /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}/info
- [`GET /api/cephfs/subvolume/{vol_name}`](#get-api-cephfs-subvolume-vol-name) - GET /api/cephfs/subvolume/{vol_name}
- [`PUT /api/cephfs/subvolume/{vol_name}`](#put-api-cephfs-subvolume-vol-name) - PUT /api/cephfs/subvolume/{vol_name}
- [`DELETE /api/cephfs/subvolume/{vol_name}`](#delete-api-cephfs-subvolume-vol-name) - DELETE /api/cephfs/subvolume/{vol_name}
- [`GET /api/cephfs/subvolume/{vol_name}/exists`](#get-api-cephfs-subvolume-vol-name-exists) - GET /api/cephfs/subvolume/{vol_name}/exists
- [`GET /api/cephfs/subvolume/{vol_name}/info`](#get-api-cephfs-subvolume-vol-name-info) - GET /api/cephfs/subvolume/{vol_name}/info
- [`GET /api/cephfs/subvolume/{vol_name}/snapshot-visibility`](#get-api-cephfs-subvolume-vol-name-snapshot-visibility) - GET /api/cephfs/subvolume/{vol_name}/snapshot-visibility
- [`PUT /api/cephfs/subvolume/{vol_name}/snapshot-visibility`](#put-api-cephfs-subvolume-vol-name-snapshot-visibility) - PUT /api/cephfs/subvolume/{vol_name}/snapshot-visibility
- [`GET /api/cephfs/{fs_id}`](#get-api-cephfs-fs-id) - GET /api/cephfs/{fs_id}
- [`DELETE /api/cephfs/{fs_id}/client/{client_id}`](#delete-api-cephfs-fs-id-client-client-id) - DELETE /api/cephfs/{fs_id}/client/{client_id}
- [`GET /api/cephfs/{fs_id}/clients`](#get-api-cephfs-fs-id-clients) - GET /api/cephfs/{fs_id}/clients
- [`GET /api/cephfs/{fs_id}/get_root_directory`](#get-api-cephfs-fs-id-get-root-directory) - GET /api/cephfs/{fs_id}/get_root_directory
- [`GET /api/cephfs/{fs_id}/ls_dir`](#get-api-cephfs-fs-id-ls-dir) - GET /api/cephfs/{fs_id}/ls_dir
- [`GET /api/cephfs/{fs_id}/mds_counters`](#get-api-cephfs-fs-id-mds-counters) - GET /api/cephfs/{fs_id}/mds_counters
- [`GET /api/cephfs/{fs_id}/quota`](#get-api-cephfs-fs-id-quota) - Get Cephfs Quotas of the specified path
- [`PUT /api/cephfs/{fs_id}/quota`](#put-api-cephfs-fs-id-quota) - PUT /api/cephfs/{fs_id}/quota
- [`PUT /api/cephfs/{fs_id}/rename-path`](#put-api-cephfs-fs-id-rename-path) - PUT /api/cephfs/{fs_id}/rename-path
- [`POST /api/cephfs/{fs_id}/snapshot`](#post-api-cephfs-fs-id-snapshot) - POST /api/cephfs/{fs_id}/snapshot
- [`DELETE /api/cephfs/{fs_id}/snapshot`](#delete-api-cephfs-fs-id-snapshot) - DELETE /api/cephfs/{fs_id}/snapshot
- [`GET /api/cephfs/{fs_id}/statfs`](#get-api-cephfs-fs-id-statfs) - Get Cephfs statfs of the specified path
- [`POST /api/cephfs/{fs_id}/tree`](#post-api-cephfs-fs-id-tree) - POST /api/cephfs/{fs_id}/tree
- [`DELETE /api/cephfs/{fs_id}/tree`](#delete-api-cephfs-fs-id-tree) - DELETE /api/cephfs/{fs_id}/tree
- [`DELETE /api/cephfs/{fs_id}/unlink`](#delete-api-cephfs-fs-id-unlink) - DELETE /api/cephfs/{fs_id}/unlink
- [`POST /api/cephfs/{fs_id}/write_to_file`](#post-api-cephfs-fs-id-write-to-file) - POST /api/cephfs/{fs_id}/write_to_file

## 接口详情

### `GET /api/cephfs`

- 摘要：GET /api/cephfs
- Tags：`Cephfs`
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


### `POST /api/cephfs`

- 摘要：POST /api/cephfs
- Tags：`Cephfs`
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
- Schema: object; fields: data_pool, metadata_pool, name, service_spec; required: name, service_spec

```yaml
properties:
  data_pool:
    type: string
  metadata_pool:
    type: string
  name:
    type: string
  service_spec:
    type: string
required:
- name
- service_spec
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


### `PUT /api/cephfs/auth`

- 摘要：Set Ceph authentication capabilities for the specified user ID in the given path
- Tags：`Cephfs`
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
- Schema: object; fields: caps, client_id, fs_name, root_squash; required: fs_name, client_id, caps, root_squash

```yaml
properties:
  caps:
    description: Path and given capabilities
    type: string
  client_id:
    description: Cephx user ID
    type: string
  fs_name:
    description: File system name
    type: string
  root_squash:
    description: File System Identifier
    type: string
required:
- fs_name
- client_id
- caps
- root_squash
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


### `DELETE /api/cephfs/remove/{name}`

- 摘要：Remove CephFS Volume
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| name | path | 是 | string |  | File System Name |


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


### `PUT /api/cephfs/rename`

- 摘要：Rename CephFS Volume
- Tags：`Cephfs`
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
- Schema: object; fields: name, new_name; required: name, new_name

```yaml
properties:
  name:
    description: Existing FS Name
    type: string
  new_name:
    description: New FS Name
    type: string
required:
- name
- new_name
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


### `POST /api/cephfs/snapshot/schedule`

- 摘要：POST /api/cephfs/snapshot/schedule
- Tags：`CephFSSnapshotSchedule`
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
- Schema: object; fields: fs, group, path, retention_policy, snap_schedule, start, subvol; required: fs, path, snap_schedule, start

```yaml
properties:
  fs:
    type: string
  group:
    type: string
  path:
    type: string
  retention_policy:
    type: string
  snap_schedule:
    type: string
  start:
    type: string
  subvol:
    type: string
required:
- fs
- path
- snap_schedule
- start
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


### `GET /api/cephfs/snapshot/schedule/{fs}`

- 摘要：GET /api/cephfs/snapshot/schedule/{fs}
- Tags：`CephFSSnapshotSchedule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs | path | 是 | string |  |  |
| path | query | 否 | string | "/" |  |
| recursive | query | 否 | boolean | true |  |


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


### `PUT /api/cephfs/snapshot/schedule/{fs}/{path}`

- 摘要：PUT /api/cephfs/snapshot/schedule/{fs}/{path}
- Tags：`CephFSSnapshotSchedule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs | path | 是 | string |  |  |
| path | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: group, retention_to_add, retention_to_remove, subvol; required: 无

```yaml
properties:
  group:
    type: string
  retention_to_add:
    type: string
  retention_to_remove:
    type: string
  subvol:
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


### `POST /api/cephfs/snapshot/schedule/{fs}/{path}/activate`

- 摘要：POST /api/cephfs/snapshot/schedule/{fs}/{path}/activate
- Tags：`CephFSSnapshotSchedule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs | path | 是 | string |  |  |
| path | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: group, schedule, start, subvol; required: schedule, start

```yaml
properties:
  group:
    type: string
  schedule:
    type: string
  start:
    type: string
  subvol:
    type: string
required:
- schedule
- start
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


### `POST /api/cephfs/snapshot/schedule/{fs}/{path}/deactivate`

- 摘要：POST /api/cephfs/snapshot/schedule/{fs}/{path}/deactivate
- Tags：`CephFSSnapshotSchedule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs | path | 是 | string |  |  |
| path | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: group, schedule, start, subvol; required: schedule, start

```yaml
properties:
  group:
    type: string
  schedule:
    type: string
  start:
    type: string
  subvol:
    type: string
required:
- schedule
- start
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


### `DELETE /api/cephfs/snapshot/schedule/{fs}/{path}/delete_snapshot`

- 摘要：DELETE /api/cephfs/snapshot/schedule/{fs}/{path}/delete_snapshot
- Tags：`CephFSSnapshotSchedule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs | path | 是 | string |  |  |
| path | path | 是 | string |  |  |
| schedule | query | 是 | string |  |  |
| start | query | 是 | string |  |  |
| retention_policy | query | 否 | string |  |  |
| subvol | query | 否 | string |  |  |
| group | query | 否 | string |  |  |


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


### `POST /api/cephfs/subvolume`

- 摘要：POST /api/cephfs/subvolume
- Tags：`CephFSSubvolume`
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
- Schema: object; fields: subvol_name, vol_name; required: vol_name, subvol_name

```yaml
properties:
  subvol_name:
    type: string
  vol_name:
    type: string
required:
- vol_name
- subvol_name
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


### `POST /api/cephfs/subvolume/group`

- 摘要：POST /api/cephfs/subvolume/group
- Tags：`CephfsSubvolumeGroup`
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
- Schema: object; fields: group_name, vol_name; required: vol_name, group_name

```yaml
properties:
  group_name:
    type: string
  vol_name:
    type: string
required:
- vol_name
- group_name
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


### `GET /api/cephfs/subvolume/group/{vol_name}`

- 摘要：GET /api/cephfs/subvolume/group/{vol_name}
- Tags：`CephfsSubvolumeGroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |
| info | query | 否 | boolean | true |  |


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


### `PUT /api/cephfs/subvolume/group/{vol_name}`

- 摘要：PUT /api/cephfs/subvolume/group/{vol_name}
- Tags：`CephfsSubvolumeGroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: group_name, size; required: group_name, size

```yaml
properties:
  group_name:
    type: string
  size:
    type: integer
required:
- group_name
- size
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


### `DELETE /api/cephfs/subvolume/group/{vol_name}`

- 摘要：DELETE /api/cephfs/subvolume/group/{vol_name}
- Tags：`CephfsSubvolumeGroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |
| group_name | query | 是 | string |  |  |


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


### `GET /api/cephfs/subvolume/group/{vol_name}/info`

- 摘要：GET /api/cephfs/subvolume/group/{vol_name}/info
- Tags：`CephfsSubvolumeGroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |
| group_name | query | 是 | string |  |  |


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


### `POST /api/cephfs/subvolume/snapshot`

- 摘要：POST /api/cephfs/subvolume/snapshot
- Tags：`CephfsSubvolumeSnapshot`
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
- Schema: object; fields: group_name, snap_name, subvol_name, vol_name; required: vol_name, subvol_name, snap_name

```yaml
properties:
  group_name:
    default: ''
    type: string
  snap_name:
    type: string
  subvol_name:
    type: string
  vol_name:
    type: string
required:
- vol_name
- subvol_name
- snap_name
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


### `POST /api/cephfs/subvolume/snapshot/clone`

- 摘要：Create a clone of a subvolume snapshot
- Tags：`CephfsSnapshotClone`
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
- Schema: object; fields: clone_name, group_name, snap_name, subvol_name, target_group_name, vol_name; required: vol_name, subvol_name, snap_name, clone_name

```yaml
properties:
  clone_name:
    type: string
  group_name:
    default: ''
    type: string
  snap_name:
    type: string
  subvol_name:
    type: string
  target_group_name:
    default: ''
    type: string
  vol_name:
    type: string
required:
- vol_name
- subvol_name
- snap_name
- clone_name
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


### `GET /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}`

- 摘要：GET /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}
- Tags：`CephfsSubvolumeSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |
| subvol_name | path | 是 | string |  |  |
| group_name | query | 否 | string | "" |  |
| info | query | 否 | boolean | true |  |


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


### `DELETE /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}`

- 摘要：DELETE /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}
- Tags：`CephfsSubvolumeSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |
| subvol_name | path | 是 | string |  |  |
| snap_name | query | 是 | string |  |  |
| group_name | query | 否 | string | "" |  |
| force | query | 否 | boolean | true |  |


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


### `GET /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}/info`

- 摘要：GET /api/cephfs/subvolume/snapshot/{vol_name}/{subvol_name}/info
- Tags：`CephfsSubvolumeSnapshot`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |
| subvol_name | path | 是 | string |  |  |
| snap_name | query | 是 | string |  |  |
| group_name | query | 否 | string | "" |  |


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


### `GET /api/cephfs/subvolume/{vol_name}`

- 摘要：GET /api/cephfs/subvolume/{vol_name}
- Tags：`CephFSSubvolume`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |
| group_name | query | 否 | string | "" |  |
| info | query | 否 | boolean | true |  |


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


### `PUT /api/cephfs/subvolume/{vol_name}`

- 摘要：PUT /api/cephfs/subvolume/{vol_name}
- Tags：`CephFSSubvolume`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: group_name, size, subvol_name; required: subvol_name, size

```yaml
properties:
  group_name:
    default: ''
    type: string
  size:
    type: integer
  subvol_name:
    type: string
required:
- subvol_name
- size
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


### `DELETE /api/cephfs/subvolume/{vol_name}`

- 摘要：DELETE /api/cephfs/subvolume/{vol_name}
- Tags：`CephFSSubvolume`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |
| subvol_name | query | 是 | string |  |  |
| group_name | query | 否 | string | "" |  |
| retain_snapshots | query | 否 | boolean | false |  |


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


### `GET /api/cephfs/subvolume/{vol_name}/exists`

- 摘要：GET /api/cephfs/subvolume/{vol_name}/exists
- Tags：`CephFSSubvolume`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |
| group_name | query | 否 | string | "" |  |


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


### `GET /api/cephfs/subvolume/{vol_name}/info`

- 摘要：GET /api/cephfs/subvolume/{vol_name}/info
- Tags：`CephFSSubvolume`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |
| subvol_name | query | 是 | string |  |  |
| group_name | query | 否 | string | "" |  |


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


### `GET /api/cephfs/subvolume/{vol_name}/snapshot-visibility`

- 摘要：GET /api/cephfs/subvolume/{vol_name}/snapshot-visibility
- Tags：`CephFSSubvolume`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |
| subvol_name | query | 是 | string |  |  |
| group_name | query | 否 | string | "" |  |


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


### `PUT /api/cephfs/subvolume/{vol_name}/snapshot-visibility`

- 摘要：PUT /api/cephfs/subvolume/{vol_name}/snapshot-visibility
- Tags：`CephFSSubvolume`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| vol_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: group_name, subvol_name, value; required: subvol_name, value

```yaml
properties:
  group_name:
    default: ''
    type: string
  subvol_name:
    type: string
  value:
    type: string
required:
- subvol_name
- value
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


### `GET /api/cephfs/{fs_id}`

- 摘要：GET /api/cephfs/{fs_id}
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |


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


### `DELETE /api/cephfs/{fs_id}/client/{client_id}`

- 摘要：DELETE /api/cephfs/{fs_id}/client/{client_id}
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |
| client_id | path | 是 | string |  |  |


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


### `GET /api/cephfs/{fs_id}/clients`

- 摘要：GET /api/cephfs/{fs_id}/clients
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |


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


### `GET /api/cephfs/{fs_id}/get_root_directory`

- 摘要：GET /api/cephfs/{fs_id}/get_root_directory
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |


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


### `GET /api/cephfs/{fs_id}/ls_dir`

- 摘要：GET /api/cephfs/{fs_id}/ls_dir
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |
| path | query | 否 | string |  |  |
| depth | query | 否 | integer | 1 |  |


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


### `GET /api/cephfs/{fs_id}/mds_counters`

- 摘要：GET /api/cephfs/{fs_id}/mds_counters
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |
| counters | query | 否 | integer |  |  |


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


### `GET /api/cephfs/{fs_id}/quota`

- 摘要：Get Cephfs Quotas of the specified path
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  | File System Identifier |
| path | query | 是 | string |  | File System Path |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: max_bytes, max_files; required: max_bytes, max_files

```yaml
properties:
  max_bytes:
    description: ''
    type: integer
  max_files:
    description: ''
    type: integer
required:
- max_bytes
- max_files
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


### `PUT /api/cephfs/{fs_id}/quota`

- 摘要：PUT /api/cephfs/{fs_id}/quota
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: max_bytes, max_files, path; required: path

```yaml
properties:
  max_bytes:
    type: string
  max_files:
    type: string
  path:
    type: string
required:
- path
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


### `PUT /api/cephfs/{fs_id}/rename-path`

- 摘要：PUT /api/cephfs/{fs_id}/rename-path
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: dst_path, src_path; required: src_path, dst_path

```yaml
properties:
  dst_path:
    type: string
  src_path:
    type: string
required:
- src_path
- dst_path
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


### `POST /api/cephfs/{fs_id}/snapshot`

- 摘要：POST /api/cephfs/{fs_id}/snapshot
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: name, path; required: path

```yaml
properties:
  name:
    type: string
  path:
    type: string
required:
- path
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


### `DELETE /api/cephfs/{fs_id}/snapshot`

- 摘要：DELETE /api/cephfs/{fs_id}/snapshot
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |
| path | query | 是 | string |  |  |
| name | query | 是 | string |  |  |


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


### `GET /api/cephfs/{fs_id}/statfs`

- 摘要：Get Cephfs statfs of the specified path
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  | File System Identifier |
| path | query | 是 | string |  | File System Path |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: bytes, files, subdirs; required: bytes, files, subdirs

```yaml
properties:
  bytes:
    description: ''
    type: integer
  files:
    description: ''
    type: integer
  subdirs:
    description: ''
    type: integer
required:
- bytes
- files
- subdirs
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


### `POST /api/cephfs/{fs_id}/tree`

- 摘要：POST /api/cephfs/{fs_id}/tree
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: path; required: path

```yaml
properties:
  path:
    type: string
required:
- path
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


### `DELETE /api/cephfs/{fs_id}/tree`

- 摘要：DELETE /api/cephfs/{fs_id}/tree
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |
| path | query | 是 | string |  |  |


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


### `DELETE /api/cephfs/{fs_id}/unlink`

- 摘要：DELETE /api/cephfs/{fs_id}/unlink
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |
| path | query | 是 | string |  |  |


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


### `POST /api/cephfs/{fs_id}/write_to_file`

- 摘要：POST /api/cephfs/{fs_id}/write_to_file
- Tags：`Cephfs`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| fs_id | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: buf, path; required: path, buf

```yaml
properties:
  buf:
    type: string
  path:
    type: string
required:
- path
- buf
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

