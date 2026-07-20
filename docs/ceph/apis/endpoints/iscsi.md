# Ceph 20.2.2 Dashboard API - iSCSI

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/iscsi` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/iscsi/discoveryauth`](#get-api-iscsi-discoveryauth) - Get Iscsi discoveryauth Details
- [`PUT /api/iscsi/discoveryauth`](#put-api-iscsi-discoveryauth) - Set Iscsi discoveryauth
- [`GET /api/iscsi/target`](#get-api-iscsi-target) - GET /api/iscsi/target
- [`POST /api/iscsi/target`](#post-api-iscsi-target) - POST /api/iscsi/target
- [`GET /api/iscsi/target/{target_iqn}`](#get-api-iscsi-target-target-iqn) - GET /api/iscsi/target/{target_iqn}
- [`PUT /api/iscsi/target/{target_iqn}`](#put-api-iscsi-target-target-iqn) - PUT /api/iscsi/target/{target_iqn}
- [`DELETE /api/iscsi/target/{target_iqn}`](#delete-api-iscsi-target-target-iqn) - DELETE /api/iscsi/target/{target_iqn}

## 接口详情

### `GET /api/iscsi/discoveryauth`

- 摘要：Get Iscsi discoveryauth Details
- Tags：`Iscsi`
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
    mutual_password:
      description: ''
      type: string
    mutual_user:
      description: ''
      type: string
    password:
      description: password
      type: string
    user:
      description: username
      type: string
  type: object
required:
- user
- password
- mutual_user
- mutual_password
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


### `PUT /api/iscsi/discoveryauth`

- 摘要：Set Iscsi discoveryauth
- Tags：`Iscsi`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| user | query | 是 | string |  | Username |
| password | query | 是 | string |  | Password |
| mutual_user | query | 是 | string |  | Mutual UserName |
| mutual_password | query | 是 | string |  | Mutual Password |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: mutual_password, mutual_user, password, user; required: user, password, mutual_user, mutual_password

```yaml
properties:
  mutual_password:
    description: Mutual Password
    type: string
  mutual_user:
    description: Mutual UserName
    type: string
  password:
    description: Password
    type: string
  user:
    description: Username
    type: string
required:
- user
- password
- mutual_user
- mutual_password
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


### `GET /api/iscsi/target`

- 摘要：GET /api/iscsi/target
- Tags：`IscsiTarget`
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


### `POST /api/iscsi/target`

- 摘要：POST /api/iscsi/target
- Tags：`IscsiTarget`
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
- Schema: object; fields: acl_enabled, auth, clients, disks, groups, portals, target_controls, target_iqn; required: 无

```yaml
properties:
  acl_enabled:
    type: string
  auth:
    type: string
  clients:
    type: string
  disks:
    type: string
  groups:
    type: string
  portals:
    type: string
  target_controls:
    type: string
  target_iqn:
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


### `GET /api/iscsi/target/{target_iqn}`

- 摘要：GET /api/iscsi/target/{target_iqn}
- Tags：`IscsiTarget`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| target_iqn | path | 是 | string |  |  |


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


### `PUT /api/iscsi/target/{target_iqn}`

- 摘要：PUT /api/iscsi/target/{target_iqn}
- Tags：`IscsiTarget`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| target_iqn | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: acl_enabled, auth, clients, disks, groups, new_target_iqn, portals, target_controls; required: 无

```yaml
properties:
  acl_enabled:
    type: string
  auth:
    type: string
  clients:
    type: string
  disks:
    type: string
  groups:
    type: string
  new_target_iqn:
    type: string
  portals:
    type: string
  target_controls:
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


### `DELETE /api/iscsi/target/{target_iqn}`

- 摘要：DELETE /api/iscsi/target/{target_iqn}
- Tags：`IscsiTarget`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| target_iqn | path | 是 | string |  |  |


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

