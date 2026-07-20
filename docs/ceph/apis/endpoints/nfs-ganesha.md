# Ceph 20.2.2 Dashboard API - NFS Ganesha

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/nfs-ganesha` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/nfs-ganesha/cluster`](#get-api-nfs-ganesha-cluster) - GET /api/nfs-ganesha/cluster
- [`GET /api/nfs-ganesha/export`](#get-api-nfs-ganesha-export) - List all or cluster specific NFS-Ganesha exports
- [`POST /api/nfs-ganesha/export`](#post-api-nfs-ganesha-export) - Creates a new NFS-Ganesha export
- [`GET /api/nfs-ganesha/export/{cluster_id}/{export_id}`](#get-api-nfs-ganesha-export-cluster-id-export-id) - Get an NFS-Ganesha export
- [`PUT /api/nfs-ganesha/export/{cluster_id}/{export_id}`](#put-api-nfs-ganesha-export-cluster-id-export-id) - Updates an NFS-Ganesha export
- [`DELETE /api/nfs-ganesha/export/{cluster_id}/{export_id}`](#delete-api-nfs-ganesha-export-cluster-id-export-id) - Deletes an NFS-Ganesha export

## 接口详情

### `GET /api/nfs-ganesha/cluster`

- 摘要：GET /api/nfs-ganesha/cluster
- Tags：`NFS-Ganesha`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| info | query | 否 | boolean | false |  |


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


### `GET /api/nfs-ganesha/export`

- 摘要：List all or cluster specific NFS-Ganesha exports
- Tags：`NFS-Ganesha`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| cluster_id | query | 否 | string |  |  |


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
    access_type:
      description: Export access type
      type: string
    clients:
      description: List of client configurations
      items:
        properties:
          access_type:
            description: Client access type
            type: string
          addresses:
            description: list of IP addresses
            items:
              type: string
            type: array
          squash:
            description: Client squash policy
            type: string
        required:
        - addresses
        - access_type
        - squash
        type: object
      type: array
    cluster_id:
      description: Cluster identifier
      type: string
    export_id:
      description: Export ID
      type: integer
    fsal:
      description: FSAL configuration
      properties:
        fs_name:
          description: CephFS filesystem name
          type: string
        name:
          description: name of FSAL
          type: string
        sec_label_xattr:
          description: Name of xattr for security label
          type: string
        user_id:
          description: User id
          type: string
      required:
      - name
      type: object
    path:
      description: Export path
      type: string
    protocols:
      description: List of protocol types
      items:
        type: integer
      type: array
    pseudo:
      description: Pseudo FS path
      type: string
    security_label:
      description: Security label
      type: string
    squash:
      description: Export squash policy
      type: string
    transports:
      description: List of transport types
      items:
        type: string
      type: array
  type: object
required:
- export_id
- path
- cluster_id
- pseudo
- access_type
- squash
- security_label
- protocols
- transports
- fsal
- clients
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


### `POST /api/nfs-ganesha/export`

- 摘要：Creates a new NFS-Ganesha export
- Tags：`NFS-Ganesha`
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
- Schema: object; fields: access_type, clients, cluster_id, fsal, path, protocols, pseudo, security_label, squash, transports; required: path, cluster_id, pseudo, access_type, squash, security_label, protocols, transports, fsal, clients

```yaml
properties:
  access_type:
    description: Export access type
    type: string
  clients:
    description: List of client configurations
    items:
      properties:
        access_type:
          description: Client access type
          type: string
        addresses:
          description: list of IP addresses
          items:
            type: string
          type: array
        squash:
          description: Client squash policy
          type: string
      required:
      - addresses
      - access_type
      - squash
      type: object
    type: array
  cluster_id:
    description: Cluster identifier
    type: string
  fsal:
    description: FSAL configuration
    properties:
      fs_name:
        description: CephFS filesystem name
        type: string
      name:
        description: name of FSAL
        type: string
      sec_label_xattr:
        description: Name of xattr for security label
        type: string
    required:
    - name
    type: object
  path:
    description: Export path
    type: string
  protocols:
    description: List of protocol types
    items:
      type: integer
    type: array
  pseudo:
    description: Pseudo FS path
    type: string
  security_label:
    description: Security label
    type: string
  squash:
    description: Export squash policy
    type: string
  transports:
    description: List of transport types
    items:
      type: string
    type: array
required:
- path
- cluster_id
- pseudo
- access_type
- squash
- security_label
- protocols
- transports
- fsal
- clients
type: object
```


#### 返回消息

#### `201`

Resource created.

- Content-Type: `application/vnd.ceph.api.v2.0+json`
- Schema: object; fields: access_type, clients, cluster_id, export_id, fsal, path, protocols, pseudo, security_label, squash, transports; required: export_id, path, cluster_id, pseudo, access_type, squash, security_label, protocols, transports, fsal, clients

```yaml
properties:
  access_type:
    description: Export access type
    type: string
  clients:
    description: List of client configurations
    items:
      properties:
        access_type:
          description: Client access type
          type: string
        addresses:
          description: list of IP addresses
          items:
            type: string
          type: array
        squash:
          description: Client squash policy
          type: string
      required:
      - addresses
      - access_type
      - squash
      type: object
    type: array
  cluster_id:
    description: Cluster identifier
    type: string
  export_id:
    description: Export ID
    type: integer
  fsal:
    description: FSAL configuration
    properties:
      fs_name:
        description: CephFS filesystem name
        type: string
      name:
        description: name of FSAL
        type: string
      sec_label_xattr:
        description: Name of xattr for security label
        type: string
      user_id:
        description: User id
        type: string
    required:
    - name
    type: object
  path:
    description: Export path
    type: string
  protocols:
    description: List of protocol types
    items:
      type: integer
    type: array
  pseudo:
    description: Pseudo FS path
    type: string
  security_label:
    description: Security label
    type: string
  squash:
    description: Export squash policy
    type: string
  transports:
    description: List of transport types
    items:
      type: string
    type: array
required:
- export_id
- path
- cluster_id
- pseudo
- access_type
- squash
- security_label
- protocols
- transports
- fsal
- clients
type: object
```

#### `202`

Operation is still executing. Please check the task queue.

- Content-Type: `application/vnd.ceph.api.v2.0+json`
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


### `GET /api/nfs-ganesha/export/{cluster_id}/{export_id}`

- 摘要：Get an NFS-Ganesha export
- Tags：`NFS-Ganesha`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| cluster_id | path | 是 | string |  | Cluster identifier |
| export_id | path | 是 | string |  | Export ID |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: access_type, clients, cluster_id, export_id, fsal, path, protocols, pseudo, security_label, squash, transports; required: export_id, path, cluster_id, pseudo, access_type, squash, security_label, protocols, transports, fsal, clients

```yaml
properties:
  access_type:
    description: Export access type
    type: string
  clients:
    description: List of client configurations
    items:
      properties:
        access_type:
          description: Client access type
          type: string
        addresses:
          description: list of IP addresses
          items:
            type: string
          type: array
        squash:
          description: Client squash policy
          type: string
      required:
      - addresses
      - access_type
      - squash
      type: object
    type: array
  cluster_id:
    description: Cluster identifier
    type: string
  export_id:
    description: Export ID
    type: integer
  fsal:
    description: FSAL configuration
    properties:
      fs_name:
        description: CephFS filesystem name
        type: string
      name:
        description: name of FSAL
        type: string
      sec_label_xattr:
        description: Name of xattr for security label
        type: string
      user_id:
        description: User id
        type: string
    required:
    - name
    type: object
  path:
    description: Export path
    type: string
  protocols:
    description: List of protocol types
    items:
      type: integer
    type: array
  pseudo:
    description: Pseudo FS path
    type: string
  security_label:
    description: Security label
    type: string
  squash:
    description: Export squash policy
    type: string
  transports:
    description: List of transport types
    items:
      type: string
    type: array
required:
- export_id
- path
- cluster_id
- pseudo
- access_type
- squash
- security_label
- protocols
- transports
- fsal
- clients
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


### `PUT /api/nfs-ganesha/export/{cluster_id}/{export_id}`

- 摘要：Updates an NFS-Ganesha export
- Tags：`NFS-Ganesha`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| cluster_id | path | 是 | string |  | Cluster identifier |
| export_id | path | 是 | integer |  | Export ID |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: access_type, clients, fsal, path, protocols, pseudo, security_label, squash, transports; required: path, pseudo, access_type, squash, security_label, protocols, transports, fsal, clients

```yaml
properties:
  access_type:
    description: Export access type
    type: string
  clients:
    description: List of client configurations
    items:
      properties:
        access_type:
          description: Client access type
          type: string
        addresses:
          description: list of IP addresses
          items:
            type: string
          type: array
        squash:
          description: Client squash policy
          type: string
      required:
      - addresses
      - access_type
      - squash
      type: object
    type: array
  fsal:
    description: FSAL configuration
    properties:
      fs_name:
        description: CephFS filesystem name
        type: string
      name:
        description: name of FSAL
        type: string
      sec_label_xattr:
        description: Name of xattr for security label
        type: string
    required:
    - name
    type: object
  path:
    description: Export path
    type: string
  protocols:
    description: List of protocol types
    items:
      type: integer
    type: array
  pseudo:
    description: Pseudo FS path
    type: string
  security_label:
    description: Security label
    type: string
  squash:
    description: Export squash policy
    type: string
  transports:
    description: List of transport types
    items:
      type: string
    type: array
required:
- path
- pseudo
- access_type
- squash
- security_label
- protocols
- transports
- fsal
- clients
type: object
```


#### 返回消息

#### `200`

Resource updated.

- Content-Type: `application/vnd.ceph.api.v2.0+json`
- Schema: object; fields: access_type, clients, cluster_id, export_id, fsal, path, protocols, pseudo, security_label, squash, transports; required: export_id, path, cluster_id, pseudo, access_type, squash, security_label, protocols, transports, fsal, clients

```yaml
properties:
  access_type:
    description: Export access type
    type: string
  clients:
    description: List of client configurations
    items:
      properties:
        access_type:
          description: Client access type
          type: string
        addresses:
          description: list of IP addresses
          items:
            type: string
          type: array
        squash:
          description: Client squash policy
          type: string
      required:
      - addresses
      - access_type
      - squash
      type: object
    type: array
  cluster_id:
    description: Cluster identifier
    type: string
  export_id:
    description: Export ID
    type: integer
  fsal:
    description: FSAL configuration
    properties:
      fs_name:
        description: CephFS filesystem name
        type: string
      name:
        description: name of FSAL
        type: string
      sec_label_xattr:
        description: Name of xattr for security label
        type: string
      user_id:
        description: User id
        type: string
    required:
    - name
    type: object
  path:
    description: Export path
    type: string
  protocols:
    description: List of protocol types
    items:
      type: integer
    type: array
  pseudo:
    description: Pseudo FS path
    type: string
  security_label:
    description: Security label
    type: string
  squash:
    description: Export squash policy
    type: string
  transports:
    description: List of transport types
    items:
      type: string
    type: array
required:
- export_id
- path
- cluster_id
- pseudo
- access_type
- squash
- security_label
- protocols
- transports
- fsal
- clients
type: object
```

#### `202`

Operation is still executing. Please check the task queue.

- Content-Type: `application/vnd.ceph.api.v2.0+json`
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


### `DELETE /api/nfs-ganesha/export/{cluster_id}/{export_id}`

- 摘要：Deletes an NFS-Ganesha export
- Tags：`NFS-Ganesha`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| cluster_id | path | 是 | string |  | Cluster identifier |
| export_id | path | 是 | integer |  | Export ID |


#### 请求体

无请求体。


#### 返回消息

#### `202`

Operation is still executing. Please check the task queue.

- Content-Type: `application/vnd.ceph.api.v2.0+json`
- Schema: object

```yaml
type: object
```

#### `204`

Resource deleted.

- Content-Type: `application/vnd.ceph.api.v2.0+json`
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

