# Ceph 20.2.2 Dashboard API - SMB

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/smb` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/smb/cluster`](#get-api-smb-cluster) - List smb clusters
- [`POST /api/smb/cluster`](#post-api-smb-cluster) - Create smb cluster
- [`GET /api/smb/cluster/{cluster_id}`](#get-api-smb-cluster-cluster-id) - Get an smb cluster
- [`DELETE /api/smb/cluster/{cluster_id}`](#delete-api-smb-cluster-cluster-id) - Remove an smb cluster
- [`GET /api/smb/joinauth`](#get-api-smb-joinauth) - List smb join authorization resources
- [`POST /api/smb/joinauth`](#post-api-smb-joinauth) - Create smb join auth
- [`GET /api/smb/joinauth/{auth_id}`](#get-api-smb-joinauth-auth-id) - Get smb join authorization resource
- [`DELETE /api/smb/joinauth/{auth_id}`](#delete-api-smb-joinauth-auth-id) - Delete smb join auth
- [`GET /api/smb/share`](#get-api-smb-share) - List smb shares
- [`POST /api/smb/share`](#post-api-smb-share) - Create smb share
- [`GET /api/smb/share/{cluster_id}/{share_id}`](#get-api-smb-share-cluster-id-share-id) - Get an smb share
- [`DELETE /api/smb/share/{cluster_id}/{share_id}`](#delete-api-smb-share-cluster-id-share-id) - Remove an smb share
- [`GET /api/smb/usersgroups`](#get-api-smb-usersgroups) - List smb user resources
- [`POST /api/smb/usersgroups`](#post-api-smb-usersgroups) - Create smb usersgroups
- [`GET /api/smb/usersgroups/{users_groups_id}`](#get-api-smb-usersgroups-users-groups-id) - Get smb usersgroups authorization resource
- [`DELETE /api/smb/usersgroups/{users_groups_id}`](#delete-api-smb-usersgroups-users-groups-id) - Delete smb join auth

## 接口详情

### `GET /api/smb/cluster`

- 摘要：List smb clusters
- Tags：`SMB`
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

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: array<object>

```yaml
items:
  properties:
    auth_mode:
      description: Either 'active-directory' or 'user'
      type: string
    cluster_id:
      description: Unique identifier for the cluster
      type: string
    custom_dns:
      description: List of custom DNS server addresses
      items:
        type: string
      type: array
    domain_settings:
      description: Domain-specific settings for active-directory auth mode
      properties:
        join_sources:
          description: List of join auth sources for domain settings
          items:
            properties:
              ref:
                description: Reference identifier for the join auth resource
                type: string
              source_type:
                description: resource
                type: string
            required:
            - source_type
            - ref
            type: object
          type: array
        realm:
          description: Domain realm, e.g., 'DOMAIN1.SINK.TEST'
          type: string
      required:
      - realm
      - join_sources
      type: object
    intent:
      description: Desired state of the resource, e.g., 'present' or 'removed'
      type: string
    placement:
      description: Placement configuration for the resource
      properties:
        count:
          description: Number of instances to place
          type: integer
      required:
      - count
      type: object
    public_addrs:
      description: Public Address
      items:
        properties:
          address:
            description: This address will be assigned to one of the host's network
              devices
            type: string
          destination:
            description: Defines where the system will assign the managed IPs.
            type: string
        required:
        - address
        - destination
        type: object
      type: array
    resource_type:
      description: ceph.smb.cluster
      type: string
    user_group_settings:
      description: User group settings for user auth mode
      items:
        properties:
          ref:
            description: Reference identifier for the user group resource
            type: string
          source_type:
            description: resource
            type: string
        required:
        - source_type
        - ref
        type: object
      type: array
  type: object
required:
- resource_type
- cluster_id
- auth_mode
- intent
- domain_settings
- user_group_settings
- custom_dns
- public_addrs
- placement
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


### `POST /api/smb/cluster`

- 摘要：Create smb cluster
- Tags：`SMB`
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
- Schema: object; fields: cluster_resource; required: cluster_resource

```yaml
properties:
  cluster_resource:
    description: cluster_resource
    type: string
required:
- cluster_resource
type: object
```


#### 返回消息

#### `201`

Resource created.

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: results, success; required: results, success

```yaml
properties:
  results:
    description: List of operation results
    items:
      properties:
        resource:
          description: Resource
          properties:
            auth_mode:
              description: Either 'active-directory' or 'user'
              type: string
            cluster_id:
              description: Unique identifier for the cluster
              type: string
            custom_dns:
              description: List of custom DNS server addresses
              items:
                type: string
              type: array
            domain_settings:
              description: Domain-specific settings for active-directory auth mode
              properties:
                join_sources:
                  description: List of join auth sources for domain settings
                  items:
                    properties:
                      ref:
                        description: Reference identifier for the join auth resource
                        type: string
                      source_type:
                        description: resource
                        type: string
                    required:
                    - source_type
                    - ref
                    type: object
                  type: array
                realm:
                  description: Domain realm, e.g., 'DOMAIN1.SINK.TEST'
                  type: string
              required:
              - realm
              - join_sources
              type: object
            intent:
              description: Desired state of the resource, e.g., 'present' or 'removed'
              type: string
            placement:
              description: Placement configuration for the resource
              properties:
                count:
                  description: Number of instances to place
                  type: integer
              required:
              - count
              type: object
            public_addrs:
              description: Public Address
              items:
                properties:
                  address:
                    description: This address will be assigned to one of the host's
                      network devices
                    type: string
                  destination:
                    description: Defines where the system will assign the managed
                      IPs.
                    type: string
                required:
                - address
                - destination
                type: object
              type: array
            resource_type:
              description: ceph.smb.cluster
              type: string
            user_group_settings:
              description: User group settings for user auth mode
              items:
                properties:
                  ref:
                    description: Reference identifier for the user group resource
                    type: string
                  source_type:
                    description: resource
                    type: string
                required:
                - source_type
                - ref
                type: object
              type: array
          required:
          - resource_type
          - cluster_id
          - auth_mode
          - intent
          - domain_settings
          - user_group_settings
          - custom_dns
          - public_addrs
          - placement
          type: object
        state:
          description: The current state of the resource,                        e.g.,
            'created', 'updated', 'deleted'
          type: string
        success:
          description: Indicates if the operation was successful
          type: boolean
      required:
      - resource
      - state
      - success
      type: object
    type: array
  success:
    description: Indicates if the overall operation was successful
    type: boolean
required:
- results
- success
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


### `GET /api/smb/cluster/{cluster_id}`

- 摘要：Get an smb cluster
- Tags：`SMB`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| cluster_id | path | 是 | string |  | Unique identifier for the cluster |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: auth_mode, cluster_id, custom_dns, domain_settings, intent, placement, public_addrs, resource_type, user_group_settings; required: resource_type, cluster_id, auth_mode, intent, domain_settings, user_group_settings, custom_dns, public_addrs, placement

```yaml
properties:
  auth_mode:
    description: Either 'active-directory' or 'user'
    type: string
  cluster_id:
    description: Unique identifier for the cluster
    type: string
  custom_dns:
    description: List of custom DNS server addresses
    items:
      type: string
    type: array
  domain_settings:
    description: Domain-specific settings for active-directory auth mode
    properties:
      join_sources:
        description: List of join auth sources for domain settings
        items:
          properties:
            ref:
              description: Reference identifier for the join auth resource
              type: string
            source_type:
              description: resource
              type: string
          required:
          - source_type
          - ref
          type: object
        type: array
      realm:
        description: Domain realm, e.g., 'DOMAIN1.SINK.TEST'
        type: string
    required:
    - realm
    - join_sources
    type: object
  intent:
    description: Desired state of the resource, e.g., 'present' or 'removed'
    type: string
  placement:
    description: Placement configuration for the resource
    properties:
      count:
        description: Number of instances to place
        type: integer
    required:
    - count
    type: object
  public_addrs:
    description: Public Address
    items:
      properties:
        address:
          description: This address will be assigned to one of the host's network
            devices
          type: string
        destination:
          description: Defines where the system will assign the managed IPs.
          type: string
      required:
      - address
      - destination
      type: object
    type: array
  resource_type:
    description: ceph.smb.cluster
    type: string
  user_group_settings:
    description: User group settings for user auth mode
    items:
      properties:
        ref:
          description: Reference identifier for the user group resource
          type: string
        source_type:
          description: resource
          type: string
      required:
      - source_type
      - ref
      type: object
    type: array
required:
- resource_type
- cluster_id
- auth_mode
- intent
- domain_settings
- user_group_settings
- custom_dns
- public_addrs
- placement
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


### `DELETE /api/smb/cluster/{cluster_id}`

- 摘要：Remove an smb cluster
- Tags：`SMB`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| cluster_id | path | 是 | string |  | Unique identifier for the cluster |


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
properties: {}
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


### `GET /api/smb/joinauth`

- 摘要：List smb join authorization resources
- Tags：`SMB`
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

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: array<object>

```yaml
items:
  properties:
    auth:
      description: Authentication credentials
      properties:
        password:
          description: Password for authentication
          type: string
        username:
          description: Username for authentication
          type: string
      required:
      - username
      - password
      type: object
    auth_id:
      description: Unique identifier for the join auth resource
      type: string
    intent:
      description: Desired state of the resource, e.g., 'present' or 'removed'
      type: string
    linked_to_cluster:
      description: Optional string containing a cluster ID.     If set, the resource
        is linked to the cluster and will be automatically removed     when the cluster
        is removed
      type: string
    resource_type:
      description: ceph.smb.join.auth
      type: string
  type: object
required:
- resource_type
- auth_id
- intent
- auth
- linked_to_cluster
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


### `POST /api/smb/joinauth`

- 摘要：Create smb join auth
- Tags：`SMB`
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
- Schema: object; fields: join_auth; required: join_auth

```yaml
properties:
  join_auth:
    type: string
required:
- join_auth
type: object
```


#### 返回消息

#### `201`

Resource created.

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: results, success; required: results, success

```yaml
properties:
  results:
    description: List of operation results
    items:
      properties:
        resource:
          description: Resource
          properties:
            auth:
              description: Authentication credentials
              properties:
                password:
                  description: Password for authentication
                  type: string
                username:
                  description: Username for authentication
                  type: string
              required:
              - username
              - password
              type: object
            auth_id:
              description: Unique identifier for the join auth resource
              type: string
            intent:
              description: Desired state of the resource, e.g., 'present' or 'removed'
              type: string
            linked_to_cluster:
              description: Optional string containing a cluster ID.     If set, the
                resource is linked to the cluster and will be automatically removed     when
                the cluster is removed
              type: string
            resource_type:
              description: ceph.smb.join.auth
              type: string
          required:
          - resource_type
          - auth_id
          - intent
          - auth
          - linked_to_cluster
          type: object
        state:
          description: The current state of the resource,                        e.g.,
            'created', 'updated', 'deleted'
          type: string
        success:
          description: Indicates if the operation was successful
          type: boolean
      required:
      - resource
      - state
      - success
      type: object
    type: array
  success:
    description: Indicates if the overall operation was successful
    type: boolean
required:
- results
- success
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


### `GET /api/smb/joinauth/{auth_id}`

- 摘要：Get smb join authorization resource
- Tags：`SMB`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| auth_id | path | 是 | string |  |  |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: auth, auth_id, intent, linked_to_cluster, resource_type; required: resource_type, auth_id, intent, auth, linked_to_cluster

```yaml
properties:
  auth:
    description: Authentication credentials
    properties:
      password:
        description: Password for authentication
        type: string
      username:
        description: Username for authentication
        type: string
    required:
    - username
    - password
    type: object
  auth_id:
    description: Unique identifier for the join auth resource
    type: string
  intent:
    description: Desired state of the resource, e.g., 'present' or 'removed'
    type: string
  linked_to_cluster:
    description: Optional string containing a cluster ID.     If set, the resource
      is linked to the cluster and will be automatically removed     when the cluster
      is removed
    type: string
  resource_type:
    description: ceph.smb.join.auth
    type: string
required:
- resource_type
- auth_id
- intent
- auth
- linked_to_cluster
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


### `DELETE /api/smb/joinauth/{auth_id}`

- 摘要：Delete smb join auth
- Tags：`SMB`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| auth_id | path | 是 | string |  | auth_id |


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
properties: {}
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


### `GET /api/smb/share`

- 摘要：List smb shares
- Tags：`SMB`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| cluster_id | query | 否 | string | "" | Unique identifier for the cluster |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: browseable, cephfs, cluster_id, intent, name, readonly, resource_type, share_id; required: resource_type, cluster_id, share_id, intent, name, readonly, browseable, cephfs

```yaml
properties:
  browseable:
    description: Indicates if the share is browseable
    type: boolean
  cephfs:
    description: Configuration for the CephFS share
    properties:
      path:
        description: Path within the CephFS file system
        type: string
      provider:
        description: Provider of the CephFS share, e.g., 'samba-vfs'
        type: string
      subvolume:
        description: Subvolume within the CephFS file system
        type: string
      subvolumegroup:
        description: Subvolume Group in CephFS file system
        type: string
      volume:
        description: Name of the CephFS file system
        type: string
    required:
    - volume
    - path
    - provider
    - subvolumegroup
    - subvolume
    type: object
  cluster_id:
    description: Unique identifier for the cluster
    type: string
  intent:
    description: Desired state of the resource, e.g., 'present' or 'removed'
    type: string
  name:
    description: Name of the share
    type: string
  readonly:
    description: Indicates if the share is read-only
    type: boolean
  resource_type:
    description: ceph.smb.share
    type: string
  share_id:
    description: Unique identifier for the share
    type: string
required:
- resource_type
- cluster_id
- share_id
- intent
- name
- readonly
- browseable
- cephfs
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


### `POST /api/smb/share`

- 摘要：Create smb share
- Tags：`SMB`
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
- Schema: object; fields: share_resource; required: share_resource

```yaml
properties:
  share_resource:
    description: share_resource
    type: string
required:
- share_resource
type: object
```


#### 返回消息

#### `201`

Resource created.

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: results, success; required: results, success

```yaml
properties:
  results:
    description: List of operation results
    items:
      properties:
        resource:
          description: Resource
          properties:
            browseable:
              description: Indicates if the share is browseable
              type: boolean
            cephfs:
              description: Configuration for the CephFS share
              properties:
                path:
                  description: Path within the CephFS file system
                  type: string
                provider:
                  description: Provider of the CephFS share, e.g., 'samba-vfs'
                  type: string
                subvolume:
                  description: Subvolume within the CephFS file system
                  type: string
                subvolumegroup:
                  description: Subvolume Group in CephFS file system
                  type: string
                volume:
                  description: Name of the CephFS file system
                  type: string
              required:
              - volume
              - path
              - provider
              - subvolumegroup
              - subvolume
              type: object
            cluster_id:
              description: Unique identifier for the cluster
              type: string
            intent:
              description: Desired state of the resource, e.g., 'present' or 'removed'
              type: string
            name:
              description: Name of the share
              type: string
            readonly:
              description: Indicates if the share is read-only
              type: boolean
            resource_type:
              description: ceph.smb.share
              type: string
            share_id:
              description: Unique identifier for the share
              type: string
          required:
          - resource_type
          - cluster_id
          - share_id
          - intent
          - name
          - readonly
          - browseable
          - cephfs
          type: object
        state:
          description: The current state of the resource,                        e.g.,
            'created', 'updated', 'deleted'
          type: string
        success:
          description: Indicates if the operation was successful
          type: boolean
      required:
      - resource
      - state
      - success
      type: object
    type: array
  success:
    description: Indicates if the overall operation was successful
    type: boolean
required:
- results
- success
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


### `GET /api/smb/share/{cluster_id}/{share_id}`

- 摘要：Get an smb share
- Tags：`SMB`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| cluster_id | path | 是 | string |  | Unique identifier for the cluster |
| share_id | path | 是 | string |  | Unique identifier for the share |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: browseable, cephfs, cluster_id, intent, name, readonly, resource_type, share_id; required: resource_type, cluster_id, share_id, intent, name, readonly, browseable, cephfs

```yaml
properties:
  browseable:
    description: Indicates if the share is browseable
    type: boolean
  cephfs:
    description: Configuration for the CephFS share
    properties:
      path:
        description: Path within the CephFS file system
        type: string
      provider:
        description: Provider of the CephFS share, e.g., 'samba-vfs'
        type: string
      subvolume:
        description: Subvolume within the CephFS file system
        type: string
      subvolumegroup:
        description: Subvolume Group in CephFS file system
        type: string
      volume:
        description: Name of the CephFS file system
        type: string
    required:
    - volume
    - path
    - provider
    - subvolumegroup
    - subvolume
    type: object
  cluster_id:
    description: Unique identifier for the cluster
    type: string
  intent:
    description: Desired state of the resource, e.g., 'present' or 'removed'
    type: string
  name:
    description: Name of the share
    type: string
  readonly:
    description: Indicates if the share is read-only
    type: boolean
  resource_type:
    description: ceph.smb.share
    type: string
  share_id:
    description: Unique identifier for the share
    type: string
required:
- resource_type
- cluster_id
- share_id
- intent
- name
- readonly
- browseable
- cephfs
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


### `DELETE /api/smb/share/{cluster_id}/{share_id}`

- 摘要：Remove an smb share
- Tags：`SMB`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| cluster_id | path | 是 | string |  | Unique identifier for the cluster |
| share_id | path | 是 | string |  | Unique identifier for the share |


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
properties: {}
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


### `GET /api/smb/usersgroups`

- 摘要：List smb user resources
- Tags：`SMB`
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

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: array<object>

```yaml
items:
  properties:
    intent:
      description: Desired state of the resource, e.g., 'present' or 'removed'
      type: string
    linked_to_cluster:
      description: Optional string containing a cluster ID.     If set, the resource
        is linked to the cluster and will be automatically removed     when the cluster
        is removed
      type: string
    resource_type:
      description: ceph.smb.usersgroups
      type: string
    users_groups_id:
      description: A short string identifying the usersgroups resource
      type: string
    values:
      description: Required object containing users and groups information
      properties:
        groups:
          description: List of group objects, each containing a name
          items:
            properties:
              name:
                description: The name of the group
                type: string
            required:
            - name
            type: object
          type: array
        users:
          description: List of user objects, each containing a name and password
          items:
            properties:
              name:
                description: The user name
                type: string
              password:
                description: The password for the user
                type: string
            required:
            - name
            - password
            type: object
          type: array
      required:
      - users
      - groups
      type: object
  type: object
required:
- resource_type
- users_groups_id
- intent
- values
- linked_to_cluster
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


### `POST /api/smb/usersgroups`

- 摘要：Create smb usersgroups
- Tags：`SMB`
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
- Schema: object; fields: usersgroups; required: usersgroups

```yaml
properties:
  usersgroups:
    type: string
required:
- usersgroups
type: object
```


#### 返回消息

#### `201`

Resource created.

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: results, success; required: results, success

```yaml
properties:
  results:
    description: List of operation results
    items:
      properties:
        resource:
          description: Resource
          properties:
            results:
              description: List of operation results
              items:
                properties:
                  resource:
                    description: Resource
                    properties:
                      auth:
                        description: Authentication credentials
                        properties:
                          password:
                            description: Password for authentication
                            type: string
                          username:
                            description: Username for authentication
                            type: string
                        required:
                        - username
                        - password
                        type: object
                      auth_id:
                        description: Unique identifier for the join auth resource
                        type: string
                      intent:
                        description: Desired state of the resource, e.g., 'present'
                          or 'removed'
                        type: string
                      linked_to_cluster:
                        description: Optional string containing a cluster ID.     If
                          set, the resource is linked to the cluster and will be automatically
                          removed     when the cluster is removed
                        type: string
                      resource_type:
                        description: ceph.smb.join.auth
                        type: string
                    required:
                    - resource_type
                    - auth_id
                    - intent
                    - auth
                    - linked_to_cluster
                    type: object
                  state:
                    description: The current state of the resource,                        e.g.,
                      'created', 'updated', 'deleted'
                    type: string
                  success:
                    description: Indicates if the operation was successful
                    type: boolean
                required:
                - resource
                - state
                - success
                type: object
              type: array
            success:
              description: Indicates if the overall operation was successful
              type: boolean
          required:
          - results
          - success
          type: object
        state:
          description: The current state of the resource,                        e.g.,
            'created', 'updated', 'deleted'
          type: string
        success:
          description: Indicates if the operation was successful
          type: boolean
      required:
      - resource
      - state
      - success
      type: object
    type: array
  success:
    description: Indicates if the overall operation was successful
    type: boolean
required:
- results
- success
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


### `GET /api/smb/usersgroups/{users_groups_id}`

- 摘要：Get smb usersgroups authorization resource
- Tags：`SMB`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| users_groups_id | path | 是 | string |  |  |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: intent, linked_to_cluster, resource_type, users_groups_id, values; required: resource_type, users_groups_id, intent, values, linked_to_cluster

```yaml
properties:
  intent:
    description: Desired state of the resource, e.g., 'present' or 'removed'
    type: string
  linked_to_cluster:
    description: Optional string containing a cluster ID.     If set, the resource
      is linked to the cluster and will be automatically removed     when the cluster
      is removed
    type: string
  resource_type:
    description: ceph.smb.usersgroups
    type: string
  users_groups_id:
    description: A short string identifying the usersgroups resource
    type: string
  values:
    description: Required object containing users and groups information
    properties:
      groups:
        description: List of group objects, each containing a name
        items:
          properties:
            name:
              description: The name of the group
              type: string
          required:
          - name
          type: object
        type: array
      users:
        description: List of user objects, each containing a name and password
        items:
          properties:
            name:
              description: The user name
              type: string
            password:
              description: The password for the user
              type: string
          required:
          - name
          - password
          type: object
        type: array
    required:
    - users
    - groups
    type: object
required:
- resource_type
- users_groups_id
- intent
- values
- linked_to_cluster
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


### `DELETE /api/smb/usersgroups/{users_groups_id}`

- 摘要：Delete smb join auth
- Tags：`SMB`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| users_groups_id | path | 是 | string |  | users_groups_id |


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
properties: {}
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

