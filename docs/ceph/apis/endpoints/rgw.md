# Ceph 20.2.2 Dashboard API - RGW

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/rgw` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/rgw/accounts`](#get-api-rgw-accounts) - GET /api/rgw/accounts
- [`POST /api/rgw/accounts`](#post-api-rgw-accounts) - Update RGW account info
- [`GET /api/rgw/accounts/{account_id}`](#get-api-rgw-accounts-account-id) - Get RGW Account by id
- [`PUT /api/rgw/accounts/{account_id}`](#put-api-rgw-accounts-account-id) - Update RGW account info
- [`DELETE /api/rgw/accounts/{account_id}`](#delete-api-rgw-accounts-account-id) - Delete RGW Account
- [`PUT /api/rgw/accounts/{account_id}/quota`](#put-api-rgw-accounts-account-id-quota) - Set RGW Account/Bucket quota
- [`PUT /api/rgw/accounts/{account_id}/quota/status`](#put-api-rgw-accounts-account-id-quota-status) - Enable/Disable RGW Account/Bucket quota
- [`GET /api/rgw/bucket`](#get-api-rgw-bucket) - GET /api/rgw/bucket
- [`POST /api/rgw/bucket`](#post-api-rgw-bucket) - POST /api/rgw/bucket
- [`DELETE /api/rgw/bucket/deleteEncryption`](#delete-api-rgw-bucket-deleteencryption) - DELETE /api/rgw/bucket/deleteEncryption
- [`GET /api/rgw/bucket/getEncryption`](#get-api-rgw-bucket-getencryption) - GET /api/rgw/bucket/getEncryption
- [`GET /api/rgw/bucket/getEncryptionConfig`](#get-api-rgw-bucket-getencryptionconfig) - GET /api/rgw/bucket/getEncryptionConfig
- [`GET /api/rgw/bucket/lifecycle`](#get-api-rgw-bucket-lifecycle) - GET /api/rgw/bucket/lifecycle
- [`PUT /api/rgw/bucket/lifecycle`](#put-api-rgw-bucket-lifecycle) - PUT /api/rgw/bucket/lifecycle
- [`GET /api/rgw/bucket/notification`](#get-api-rgw-bucket-notification) - Get the bucket notification
- [`PUT /api/rgw/bucket/notification`](#put-api-rgw-bucket-notification) - Create or update the bucket notification
- [`DELETE /api/rgw/bucket/notification`](#delete-api-rgw-bucket-notification) - Delete the bucket notification
- [`GET /api/rgw/bucket/ratelimit`](#get-api-rgw-bucket-ratelimit) - Get the bucket global rate limit
- [`PUT /api/rgw/bucket/setEncryptionConfig`](#put-api-rgw-bucket-setencryptionconfig) - PUT /api/rgw/bucket/setEncryptionConfig
- [`GET /api/rgw/bucket/{bucket}`](#get-api-rgw-bucket-bucket) - GET /api/rgw/bucket/{bucket}
- [`PUT /api/rgw/bucket/{bucket}`](#put-api-rgw-bucket-bucket) - PUT /api/rgw/bucket/{bucket}
- [`DELETE /api/rgw/bucket/{bucket}`](#delete-api-rgw-bucket-bucket) - DELETE /api/rgw/bucket/{bucket}
- [`GET /api/rgw/bucket/{uid}/ratelimit`](#get-api-rgw-bucket-uid-ratelimit) - Get the bucket rate limit
- [`PUT /api/rgw/bucket/{uid}/ratelimit`](#put-api-rgw-bucket-uid-ratelimit) - Update the bucket rate limit
- [`GET /api/rgw/daemon`](#get-api-rgw-daemon) - Display RGW Daemons
- [`PUT /api/rgw/daemon/set_multisite_config`](#put-api-rgw-daemon-set-multisite-config) - PUT /api/rgw/daemon/set_multisite_config
- [`GET /api/rgw/daemon/{svc_id}`](#get-api-rgw-daemon-svc-id) - GET /api/rgw/daemon/{svc_id}
- [`PUT /api/rgw/multisite/sync-flow`](#put-api-rgw-multisite-sync-flow) - Create or update the sync flow
- [`DELETE /api/rgw/multisite/sync-flow/{flow_id}/{flow_type}/{group_id}`](#delete-api-rgw-multisite-sync-flow-flow-id-flow-type-group-id) - Remove the sync flow
- [`PUT /api/rgw/multisite/sync-pipe`](#put-api-rgw-multisite-sync-pipe) - Create or update the sync pipe
- [`DELETE /api/rgw/multisite/sync-pipe/{group_id}/{pipe_id}`](#delete-api-rgw-multisite-sync-pipe-group-id-pipe-id) - Remove the sync pipe
- [`GET /api/rgw/multisite/sync-policy`](#get-api-rgw-multisite-sync-policy) - Get the sync policy
- [`POST /api/rgw/multisite/sync-policy-group`](#post-api-rgw-multisite-sync-policy-group) - Create the sync policy group
- [`PUT /api/rgw/multisite/sync-policy-group`](#put-api-rgw-multisite-sync-policy-group) - Update the sync policy group
- [`GET /api/rgw/multisite/sync-policy-group/{group_id}`](#get-api-rgw-multisite-sync-policy-group-group-id) - Get the sync policy group
- [`DELETE /api/rgw/multisite/sync-policy-group/{group_id}`](#delete-api-rgw-multisite-sync-policy-group-group-id) - Remove the sync policy group
- [`GET /api/rgw/multisite/sync_status`](#get-api-rgw-multisite-sync-status) - Get the sync status
- [`GET /api/rgw/realm`](#get-api-rgw-realm) - GET /api/rgw/realm
- [`POST /api/rgw/realm`](#post-api-rgw-realm) - POST /api/rgw/realm
- [`GET /api/rgw/realm/get_all_realms_info`](#get-api-rgw-realm-get-all-realms-info) - GET /api/rgw/realm/get_all_realms_info
- [`GET /api/rgw/realm/get_realm_tokens`](#get-api-rgw-realm-get-realm-tokens) - GET /api/rgw/realm/get_realm_tokens
- [`POST /api/rgw/realm/import_realm_token`](#post-api-rgw-realm-import-realm-token) - POST /api/rgw/realm/import_realm_token
- [`GET /api/rgw/realm/{realm_name}`](#get-api-rgw-realm-realm-name) - GET /api/rgw/realm/{realm_name}
- [`PUT /api/rgw/realm/{realm_name}`](#put-api-rgw-realm-realm-name) - PUT /api/rgw/realm/{realm_name}
- [`DELETE /api/rgw/realm/{realm_name}`](#delete-api-rgw-realm-realm-name) - DELETE /api/rgw/realm/{realm_name}
- [`GET /api/rgw/roles`](#get-api-rgw-roles) - List RGW roles
- [`POST /api/rgw/roles`](#post-api-rgw-roles) - Create RGW role
- [`PUT /api/rgw/roles`](#put-api-rgw-roles) - Edit RGW role
- [`DELETE /api/rgw/roles/{role_name}`](#delete-api-rgw-roles-role-name) - Delete RGW role
- [`GET /api/rgw/site`](#get-api-rgw-site) - GET /api/rgw/site
- [`GET /api/rgw/topic`](#get-api-rgw-topic) - Get RGW Topic List
- [`POST /api/rgw/topic`](#post-api-rgw-topic) - Create a new RGW Topic
- [`GET /api/rgw/topic/{key}`](#get-api-rgw-topic-key) - Get RGW Topic
- [`DELETE /api/rgw/topic/{key}`](#delete-api-rgw-topic-key) - Delete RGW Topic
- [`GET /api/rgw/user`](#get-api-rgw-user) - Display RGW Users
- [`POST /api/rgw/user`](#post-api-rgw-user) - POST /api/rgw/user
- [`GET /api/rgw/user/get_emails`](#get-api-rgw-user-get-emails) - GET /api/rgw/user/get_emails
- [`GET /api/rgw/user/ratelimit`](#get-api-rgw-user-ratelimit) - Get the user global rate limit
- [`GET /api/rgw/user/{uid}`](#get-api-rgw-user-uid) - GET /api/rgw/user/{uid}
- [`PUT /api/rgw/user/{uid}`](#put-api-rgw-user-uid) - PUT /api/rgw/user/{uid}
- [`DELETE /api/rgw/user/{uid}`](#delete-api-rgw-user-uid) - DELETE /api/rgw/user/{uid}
- [`POST /api/rgw/user/{uid}/capability`](#post-api-rgw-user-uid-capability) - POST /api/rgw/user/{uid}/capability
- [`DELETE /api/rgw/user/{uid}/capability`](#delete-api-rgw-user-uid-capability) - DELETE /api/rgw/user/{uid}/capability
- [`POST /api/rgw/user/{uid}/key`](#post-api-rgw-user-uid-key) - POST /api/rgw/user/{uid}/key
- [`DELETE /api/rgw/user/{uid}/key`](#delete-api-rgw-user-uid-key) - DELETE /api/rgw/user/{uid}/key
- [`GET /api/rgw/user/{uid}/quota`](#get-api-rgw-user-uid-quota) - GET /api/rgw/user/{uid}/quota
- [`PUT /api/rgw/user/{uid}/quota`](#put-api-rgw-user-uid-quota) - PUT /api/rgw/user/{uid}/quota
- [`GET /api/rgw/user/{uid}/ratelimit`](#get-api-rgw-user-uid-ratelimit) - Get the user rate limit
- [`PUT /api/rgw/user/{uid}/ratelimit`](#put-api-rgw-user-uid-ratelimit) - Update the user rate limit
- [`POST /api/rgw/user/{uid}/subuser`](#post-api-rgw-user-uid-subuser) - POST /api/rgw/user/{uid}/subuser
- [`DELETE /api/rgw/user/{uid}/subuser/{subuser}`](#delete-api-rgw-user-uid-subuser-subuser) - DELETE /api/rgw/user/{uid}/subuser/{subuser}
- [`GET /api/rgw/zone`](#get-api-rgw-zone) - GET /api/rgw/zone
- [`POST /api/rgw/zone`](#post-api-rgw-zone) - POST /api/rgw/zone
- [`PUT /api/rgw/zone/create_system_user`](#put-api-rgw-zone-create-system-user) - PUT /api/rgw/zone/create_system_user
- [`GET /api/rgw/zone/get_all_zones_info`](#get-api-rgw-zone-get-all-zones-info) - GET /api/rgw/zone/get_all_zones_info
- [`GET /api/rgw/zone/get_pool_names`](#get-api-rgw-zone-get-pool-names) - GET /api/rgw/zone/get_pool_names
- [`GET /api/rgw/zone/get_user_list`](#get-api-rgw-zone-get-user-list) - GET /api/rgw/zone/get_user_list
- [`POST /api/rgw/zone/storage-class`](#post-api-rgw-zone-storage-class) - POST /api/rgw/zone/storage-class
- [`PUT /api/rgw/zone/storage-class`](#put-api-rgw-zone-storage-class) - PUT /api/rgw/zone/storage-class
- [`GET /api/rgw/zone/{zone_name}`](#get-api-rgw-zone-zone-name) - GET /api/rgw/zone/{zone_name}
- [`PUT /api/rgw/zone/{zone_name}`](#put-api-rgw-zone-zone-name) - PUT /api/rgw/zone/{zone_name}
- [`DELETE /api/rgw/zone/{zone_name}`](#delete-api-rgw-zone-zone-name) - DELETE /api/rgw/zone/{zone_name}
- [`GET /api/rgw/zonegroup`](#get-api-rgw-zonegroup) - GET /api/rgw/zonegroup
- [`POST /api/rgw/zonegroup`](#post-api-rgw-zonegroup) - POST /api/rgw/zonegroup
- [`GET /api/rgw/zonegroup/get_all_zonegroups_info`](#get-api-rgw-zonegroup-get-all-zonegroups-info) - GET /api/rgw/zonegroup/get_all_zonegroups_info
- [`GET /api/rgw/zonegroup/get_placement_target_by_placement_id/{placement_id}`](#get-api-rgw-zonegroup-get-placement-target-by-placement-id-placement-id) - GET /api/rgw/zonegroup/get_placement_target_by_placement_id/{placement_id}
- [`POST /api/rgw/zonegroup/storage-class`](#post-api-rgw-zonegroup-storage-class) - POST /api/rgw/zonegroup/storage-class
- [`PUT /api/rgw/zonegroup/storage-class`](#put-api-rgw-zonegroup-storage-class) - PUT /api/rgw/zonegroup/storage-class
- [`DELETE /api/rgw/zonegroup/storage-class/{placement_id}/{storage_class}`](#delete-api-rgw-zonegroup-storage-class-placement-id-storage-class) - DELETE /api/rgw/zonegroup/storage-class/{placement_id}/{storage_class}
- [`GET /api/rgw/zonegroup/{zonegroup_name}`](#get-api-rgw-zonegroup-zonegroup-name) - GET /api/rgw/zonegroup/{zonegroup_name}
- [`PUT /api/rgw/zonegroup/{zonegroup_name}`](#put-api-rgw-zonegroup-zonegroup-name) - PUT /api/rgw/zonegroup/{zonegroup_name}
- [`DELETE /api/rgw/zonegroup/{zonegroup_name}`](#delete-api-rgw-zonegroup-zonegroup-name) - DELETE /api/rgw/zonegroup/{zonegroup_name}

## 接口详情

### `GET /api/rgw/accounts`

- 摘要：GET /api/rgw/accounts
- Tags：`RgwUserAccounts`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| daemon_name | query | 否 | string |  |  |
| detailed | query | 否 | boolean | false |  |


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


### `POST /api/rgw/accounts`

- 摘要：Update RGW account info
- Tags：`RgwUserAccounts`
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
- Schema: object; fields: account_name, daemon_name, email, max_access_keys, max_buckets, max_group, max_roles, max_users, tenant; required: account_name

```yaml
properties:
  account_name:
    description: Account name
    type: string
  daemon_name:
    description: Name of the daemon
    type: string
  email:
    description: Email
    type: string
  max_access_keys:
    description: Max access keys
    type: integer
  max_buckets:
    description: Max buckets
    type: integer
  max_group:
    description: Max groups
    type: integer
  max_roles:
    description: Max roles
    type: integer
  max_users:
    description: Max users
    type: integer
  tenant:
    description: Tenant
    type: string
required:
- account_name
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


### `GET /api/rgw/accounts/{account_id}`

- 摘要：Get RGW Account by id
- Tags：`RgwUserAccounts`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| account_id | path | 是 | string |  | Account id |
| daemon_name | query | 否 | string |  | Name of the daemon |


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


### `PUT /api/rgw/accounts/{account_id}`

- 摘要：Update RGW account info
- Tags：`RgwUserAccounts`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| account_id | path | 是 | string |  | Account id |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: account_name, daemon_name, email, max_access_keys, max_buckets, max_group, max_roles, max_users, tenant; required: account_name

```yaml
properties:
  account_name:
    description: Account name
    type: string
  daemon_name:
    description: Name of the daemon
    type: string
  email:
    description: Email
    type: string
  max_access_keys:
    description: Max access keys
    type: integer
  max_buckets:
    description: Max buckets
    type: integer
  max_group:
    description: Max groups
    type: integer
  max_roles:
    description: Max roles
    type: integer
  max_users:
    description: Max users
    type: integer
  tenant:
    description: Tenant
    type: string
required:
- account_name
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


### `DELETE /api/rgw/accounts/{account_id}`

- 摘要：Delete RGW Account
- Tags：`RgwUserAccounts`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| account_id | path | 是 | string |  | Account id |
| daemon_name | query | 否 | string |  | Name of the daemon |


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


### `PUT /api/rgw/accounts/{account_id}/quota`

- 摘要：Set RGW Account/Bucket quota
- Tags：`RgwUserAccounts`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| account_id | path | 是 | string |  | Account id |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: enabled, max_objects, max_size, quota_type; required: quota_type, max_size, max_objects, enabled

```yaml
properties:
  enabled:
    type: string
  max_objects:
    description: Max objects
    type: string
  max_size:
    description: Max size
    type: string
  quota_type:
    description: Quota type
    type: string
required:
- quota_type
- max_size
- max_objects
- enabled
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


### `PUT /api/rgw/accounts/{account_id}/quota/status`

- 摘要：Enable/Disable RGW Account/Bucket quota
- Tags：`RgwUserAccounts`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| account_id | path | 是 | string |  | Account id |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: quota_status, quota_type; required: quota_type, quota_status

```yaml
properties:
  quota_status:
    description: Quota status
    type: string
  quota_type:
    description: Quota type
    type: string
required:
- quota_type
- quota_status
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


### `GET /api/rgw/bucket`

- 摘要：GET /api/rgw/bucket
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| stats | query | 否 | boolean | false |  |
| daemon_name | query | 否 | string |  |  |
| uid | query | 否 | string |  |  |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.1+json`
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


### `POST /api/rgw/bucket`

- 摘要：POST /api/rgw/bucket
- Tags：`RgwBucket`
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
- Schema: object; fields: bucket, bucket_policy, canned_acl, daemon_name, encryption_state, encryption_type, key_id, lock_enabled, lock_mode, lock_retention_period_days, lock_retention_period_years, placement_target, replication, tags, uid, zonegroup; required: bucket, uid

```yaml
properties:
  bucket:
    type: string
  bucket_policy:
    type: string
  canned_acl:
    type: string
  daemon_name:
    type: string
  encryption_state:
    default: 'false'
    type: string
  encryption_type:
    type: string
  key_id:
    type: string
  lock_enabled:
    default: 'false'
    type: string
  lock_mode:
    type: string
  lock_retention_period_days:
    type: string
  lock_retention_period_years:
    type: string
  placement_target:
    type: string
  replication:
    default: 'false'
    type: string
  tags:
    type: string
  uid:
    type: string
  zonegroup:
    type: string
required:
- bucket
- uid
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


### `DELETE /api/rgw/bucket/deleteEncryption`

- 摘要：DELETE /api/rgw/bucket/deleteEncryption
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v17.2.9
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| bucket_name | query | 是 | string |  |  |
| daemon_name | query | 否 | string |  |  |
| owner | query | 否 | string |  |  |


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


### `GET /api/rgw/bucket/getEncryption`

- 摘要：GET /api/rgw/bucket/getEncryption
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v17.2.9
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| bucket_name | query | 是 | string |  |  |
| daemon_name | query | 否 | string |  |  |
| owner | query | 否 | string |  |  |


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


### `GET /api/rgw/bucket/getEncryptionConfig`

- 摘要：GET /api/rgw/bucket/getEncryptionConfig
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v17.2.9
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| daemon_name | query | 否 | string |  |  |
| owner | query | 否 | string |  |  |


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


### `GET /api/rgw/bucket/lifecycle`

- 摘要：GET /api/rgw/bucket/lifecycle
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| bucket_name | query | 否 | string | "" |  |
| daemon_name | query | 否 | string |  |  |
| owner | query | 否 | string |  |  |
| tenant | query | 否 | string |  |  |


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


### `PUT /api/rgw/bucket/lifecycle`

- 摘要：PUT /api/rgw/bucket/lifecycle
- Tags：`RgwBucket`
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
- Schema: object; fields: bucket_name, daemon_name, lifecycle, owner, tenant; required: 无

```yaml
properties:
  bucket_name:
    default: ''
    type: string
  daemon_name:
    type: string
  lifecycle:
    default: ''
    type: string
  owner:
    type: string
  tenant:
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


### `GET /api/rgw/bucket/notification`

- 摘要：Get the bucket notification
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| bucket_name | query | 是 | string |  |  |
| daemon_name | query | 否 | string |  |  |
| owner | query | 否 | string |  |  |


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


### `PUT /api/rgw/bucket/notification`

- 摘要：Create or update the bucket notification
- Tags：`RgwBucket`
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
- Schema: object; fields: bucket_name, daemon_name, notification, owner; required: bucket_name

```yaml
properties:
  bucket_name:
    type: string
  daemon_name:
    type: string
  notification:
    default: ''
    type: string
  owner:
    type: string
required:
- bucket_name
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


### `DELETE /api/rgw/bucket/notification`

- 摘要：Delete the bucket notification
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| bucket_name | query | 是 | string |  |  |
| notification_id | query | 是 | string |  |  |
| daemon_name | query | 否 | string |  |  |
| owner | query | 否 | string |  |  |


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


### `GET /api/rgw/bucket/ratelimit`

- 摘要：Get the bucket global rate limit
- Tags：`RgwBucket`
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


### `PUT /api/rgw/bucket/setEncryptionConfig`

- 摘要：PUT /api/rgw/bucket/setEncryptionConfig
- Tags：`RgwBucket`
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
- Schema: object; fields: config, daemon_name, encryption_type, kms_provider; required: 无

```yaml
properties:
  config:
    type: string
  daemon_name:
    type: string
  encryption_type:
    type: string
  kms_provider:
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


### `GET /api/rgw/bucket/{bucket}`

- 摘要：GET /api/rgw/bucket/{bucket}
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| bucket | path | 是 | string |  |  |
| daemon_name | query | 否 | string |  |  |


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


### `PUT /api/rgw/bucket/{bucket}`

- 摘要：PUT /api/rgw/bucket/{bucket}
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| bucket | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: bucket_id, bucket_policy, canned_acl, daemon_name, encryption_state, encryption_type, key_id, lifecycle, lock_mode, lock_retention_period_days, lock_retention_period_years, mfa_delete, mfa_token_pin, mfa_token_serial, replication, tags, uid, versioning_state; required: bucket_id

```yaml
properties:
  bucket_id:
    type: string
  bucket_policy:
    type: string
  canned_acl:
    type: string
  daemon_name:
    type: string
  encryption_state:
    default: 'false'
    type: string
  encryption_type:
    type: string
  key_id:
    type: string
  lifecycle:
    type: string
  lock_mode:
    type: string
  lock_retention_period_days:
    type: string
  lock_retention_period_years:
    type: string
  mfa_delete:
    type: string
  mfa_token_pin:
    type: string
  mfa_token_serial:
    type: string
  replication:
    type: string
  tags:
    type: string
  uid:
    type: string
  versioning_state:
    type: string
required:
- bucket_id
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


### `DELETE /api/rgw/bucket/{bucket}`

- 摘要：DELETE /api/rgw/bucket/{bucket}
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| bucket | path | 是 | string |  |  |
| daemon_name | query | 否 | string |  |  |


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


### `GET /api/rgw/bucket/{uid}/ratelimit`

- 摘要：Get the bucket rate limit
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |


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


### `PUT /api/rgw/bucket/{uid}/ratelimit`

- 摘要：Update the bucket rate limit
- Tags：`RgwBucket`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: enabled, max_read_bytes, max_read_ops, max_write_bytes, max_write_ops; required: enabled, max_read_ops, max_write_ops, max_read_bytes, max_write_bytes

```yaml
properties:
  enabled:
    type: string
  max_read_bytes:
    type: string
  max_read_ops:
    type: string
  max_write_bytes:
    type: string
  max_write_ops:
    type: string
required:
- enabled
- max_read_ops
- max_write_ops
- max_read_bytes
- max_write_bytes
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


### `GET /api/rgw/daemon`

- 摘要：Display RGW Daemons
- Tags：`RgwDaemon`
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
    id:
      description: Daemon ID
      type: string
    port:
      description: Port
      type: integer
    server_hostname:
      description: ''
      type: string
    version:
      description: Ceph Version
      type: string
    zone_name:
      description: Zone
      type: string
    zonegroup_name:
      description: Zone Group
      type: string
  type: object
required:
- id
- version
- server_hostname
- zonegroup_name
- zone_name
- port
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


### `PUT /api/rgw/daemon/set_multisite_config`

- 摘要：PUT /api/rgw/daemon/set_multisite_config
- Tags：`RgwDaemon`
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
- Schema: object; fields: daemon_name, realm_name, zone_name, zonegroup_name; required: 无

```yaml
properties:
  daemon_name:
    type: string
  realm_name:
    type: string
  zone_name:
    type: string
  zonegroup_name:
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


### `GET /api/rgw/daemon/{svc_id}`

- 摘要：GET /api/rgw/daemon/{svc_id}
- Tags：`RgwDaemon`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| svc_id | path | 是 | string |  |  |


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


### `PUT /api/rgw/multisite/sync-flow`

- 摘要：Create or update the sync flow
- Tags：`RgwMultisite`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: bucket_name, destination_zone, flow_id, flow_type, group_id, source_zone, zones; required: flow_id, flow_type, group_id

```yaml
properties:
  bucket_name:
    default: ''
    type: string
  destination_zone:
    type: string
  flow_id:
    type: string
  flow_type:
    type: string
  group_id:
    type: string
  source_zone:
    type: string
  zones:
    type: string
required:
- flow_id
- flow_type
- group_id
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


### `DELETE /api/rgw/multisite/sync-flow/{flow_id}/{flow_type}/{group_id}`

- 摘要：Remove the sync flow
- Tags：`RgwMultisite`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| flow_id | path | 是 | string |  |  |
| flow_type | path | 是 | string |  |  |
| group_id | path | 是 | string |  |  |
| source_zone | query | 否 | string | "" |  |
| destination_zone | query | 否 | string | "" |  |
| zones | query | 否 | string |  |  |
| bucket_name | query | 否 | string | "" |  |


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


### `PUT /api/rgw/multisite/sync-pipe`

- 摘要：Create or update the sync pipe
- Tags：`RgwMultisite`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: bucket_name, destination_bucket, destination_zones, group_id, mode, pipe_id, source_bucket, source_zones, user; required: group_id, pipe_id, source_zones, destination_zones

```yaml
properties:
  bucket_name:
    default: ''
    type: string
  destination_bucket:
    default: ''
    type: string
  destination_zones:
    type: string
  group_id:
    type: string
  mode:
    default: ''
    type: string
  pipe_id:
    type: string
  source_bucket:
    default: ''
    type: string
  source_zones:
    type: string
  user:
    default: ''
    type: string
required:
- group_id
- pipe_id
- source_zones
- destination_zones
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


### `DELETE /api/rgw/multisite/sync-pipe/{group_id}/{pipe_id}`

- 摘要：Remove the sync pipe
- Tags：`RgwMultisite`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| group_id | path | 是 | string |  |  |
| pipe_id | path | 是 | string |  |  |
| source_zones | query | 否 | string |  |  |
| destination_zones | query | 否 | string |  |  |
| bucket_name | query | 否 | string | "" |  |


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


### `GET /api/rgw/multisite/sync-policy`

- 摘要：Get the sync policy
- Tags：`RgwMultisite`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| bucket_name | query | 否 | string | "" |  |
| zonegroup_name | query | 否 | string | "" |  |
| all_policy | query | 否 | string |  |  |


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


### `POST /api/rgw/multisite/sync-policy-group`

- 摘要：Create the sync policy group
- Tags：`RgwMultisite`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: bucket_name, group_id, status; required: group_id, status

```yaml
properties:
  bucket_name:
    default: ''
    type: string
  group_id:
    type: string
  status:
    type: string
required:
- group_id
- status
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


### `PUT /api/rgw/multisite/sync-policy-group`

- 摘要：Update the sync policy group
- Tags：`RgwMultisite`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: bucket_name, group_id, status; required: group_id, status

```yaml
properties:
  bucket_name:
    default: ''
    type: string
  group_id:
    type: string
  status:
    type: string
required:
- group_id
- status
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


### `GET /api/rgw/multisite/sync-policy-group/{group_id}`

- 摘要：Get the sync policy group
- Tags：`RgwMultisite`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| group_id | path | 是 | string |  |  |
| bucket_name | query | 否 | string | "" |  |


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


### `DELETE /api/rgw/multisite/sync-policy-group/{group_id}`

- 摘要：Remove the sync policy group
- Tags：`RgwMultisite`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| group_id | path | 是 | string |  |  |
| bucket_name | query | 否 | string | "" |  |


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


### `GET /api/rgw/multisite/sync_status`

- 摘要：Get the sync status
- Tags：`RgwMultisite`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| daemon_name | query | 否 | string |  |  |


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


### `GET /api/rgw/realm`

- 摘要：GET /api/rgw/realm
- Tags：`RgwRealm`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| replicable | query | 否 | string |  |  |


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


### `POST /api/rgw/realm`

- 摘要：POST /api/rgw/realm
- Tags：`RgwRealm`
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
- Schema: object; fields: default, realm_name; required: realm_name, default

```yaml
properties:
  default:
    type: string
  realm_name:
    type: string
required:
- realm_name
- default
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


### `GET /api/rgw/realm/get_all_realms_info`

- 摘要：GET /api/rgw/realm/get_all_realms_info
- Tags：`RgwRealm`
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


### `GET /api/rgw/realm/get_realm_tokens`

- 摘要：GET /api/rgw/realm/get_realm_tokens
- Tags：`RgwRealm`
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


### `POST /api/rgw/realm/import_realm_token`

- 摘要：POST /api/rgw/realm/import_realm_token
- Tags：`RgwRealm`
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
- Schema: object; fields: placement_spec, port, realm_token, tier_type, zone_name; required: realm_token, zone_name, port

```yaml
properties:
  placement_spec:
    type: string
  port:
    type: string
  realm_token:
    type: string
  tier_type:
    type: string
  zone_name:
    type: string
required:
- realm_token
- zone_name
- port
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


### `GET /api/rgw/realm/{realm_name}`

- 摘要：GET /api/rgw/realm/{realm_name}
- Tags：`RgwRealm`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| realm_name | path | 是 | string |  |  |


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


### `PUT /api/rgw/realm/{realm_name}`

- 摘要：PUT /api/rgw/realm/{realm_name}
- Tags：`RgwRealm`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| realm_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: default, new_realm_name; required: new_realm_name

```yaml
properties:
  default:
    default: ''
    type: string
  new_realm_name:
    type: string
required:
- new_realm_name
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


### `DELETE /api/rgw/realm/{realm_name}`

- 摘要：DELETE /api/rgw/realm/{realm_name}
- Tags：`RgwRealm`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| realm_name | path | 是 | string |  |  |


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


### `GET /api/rgw/roles`

- 摘要：List RGW roles
- Tags：`RGW`
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


### `POST /api/rgw/roles`

- 摘要：Create RGW role
- Tags：`RGW`
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
- Schema: object; fields: role_assume_policy_doc, role_name, role_path; required: 无

```yaml
properties:
  role_assume_policy_doc:
    default: ''
    type: string
  role_name:
    default: ''
    type: string
  role_path:
    default: ''
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


### `PUT /api/rgw/roles`

- 摘要：Edit RGW role
- Tags：`RGW`
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
- Schema: object; fields: max_session_duration, role_name; required: role_name, max_session_duration

```yaml
properties:
  max_session_duration:
    type: string
  role_name:
    type: string
required:
- role_name
- max_session_duration
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


### `DELETE /api/rgw/roles/{role_name}`

- 摘要：Delete RGW role
- Tags：`RGW`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| role_name | path | 是 | string |  |  |


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


### `GET /api/rgw/site`

- 摘要：GET /api/rgw/site
- Tags：`RgwSite`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| query | query | 否 | string |  |  |
| daemon_name | query | 否 | string |  |  |


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


### `GET /api/rgw/topic`

- 摘要：Get RGW Topic List
- Tags：`RGW Topic Management`
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


### `POST /api/rgw/topic`

- 摘要：Create a new RGW Topic
- Tags：`RGW Topic Management`
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
- Schema: object; fields: ack_level, amqp_exchange, ca_location, cloud_events, daemon_name, kafka_brokers, max_retries, mechanism, name, opaque_data, owner, persistent, policy, push_endpoint, retry_sleep_duration, time_to_live, use_ssl, verify_ssl; required: name

```yaml
properties:
  ack_level:
    description: Amqp ack level
    type: string
  amqp_exchange:
    description: Amqp exchange
    type: string
  ca_location:
    description: Ca location
    type: string
  cloud_events:
    default: false
    description: Cloud events
    type: string
  daemon_name:
    description: Name of the daemon
    type: string
  kafka_brokers:
    description: Kafka brokers
    type: string
  max_retries:
    description: Max retries
    type: string
  mechanism:
    description: Mechanism
    type: string
  name:
    description: Name of the topic
    type: string
  opaque_data:
    description: OpaqueData
    type: string
  owner:
    description: Name of the owner
    type: string
  persistent:
    default: false
    description: Persistent
    type: boolean
  policy:
    description: Policy
    type: string
  push_endpoint:
    description: Push Endpoint
    type: string
  retry_sleep_duration:
    description: Retry sleep duration
    type: string
  time_to_live:
    description: Time to live
    type: string
  use_ssl:
    default: false
    description: Use ssl
    type: boolean
  verify_ssl:
    default: false
    description: Verify ssl
    type: boolean
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


### `GET /api/rgw/topic/{key}`

- 摘要：Get RGW Topic
- Tags：`RGW Topic Management`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| key | path | 是 | string |  | The metadata object key to retrieve the topic e.g owner:topic_name |


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


### `DELETE /api/rgw/topic/{key}`

- 摘要：Delete RGW Topic
- Tags：`RGW Topic Management`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| key | path | 是 | string |  | The metadata object key to retrieve the topic e.g topic:topic_name |


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


### `GET /api/rgw/user`

- 摘要：Display RGW Users
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| daemon_name | query | 否 | string |  |  |
| detailed | query | 否 | boolean | false | If true, returns complete user details for each user. If false, returns only the list of usernames. |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: list_of_users; required: list_of_users

```yaml
properties:
  list_of_users:
    description: list of rgw users
    items:
      type: string
    type: array
required:
- list_of_users
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


### `POST /api/rgw/user`

- 摘要：POST /api/rgw/user
- Tags：`RgwUser`
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
- Schema: object; fields: access_key, account_id, account_policies, account_root_user, daemon_name, display_name, email, generate_key, max_buckets, secret_key, suspended, system, uid; required: uid, display_name

```yaml
properties:
  access_key:
    type: string
  account_id:
    type: integer
  account_policies:
    type: integer
  account_root_user:
    default: false
    type: integer
  daemon_name:
    type: string
  display_name:
    type: string
  email:
    type: string
  generate_key:
    type: string
  max_buckets:
    type: string
  secret_key:
    type: string
  suspended:
    type: string
  system:
    type: string
  uid:
    type: string
required:
- uid
- display_name
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


### `GET /api/rgw/user/get_emails`

- 摘要：GET /api/rgw/user/get_emails
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| daemon_name | query | 否 | string |  |  |


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


### `GET /api/rgw/user/ratelimit`

- 摘要：Get the user global rate limit
- Tags：`RgwUser`
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


### `GET /api/rgw/user/{uid}`

- 摘要：GET /api/rgw/user/{uid}
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |
| daemon_name | query | 否 | string |  |  |
| stats | query | 否 | boolean | true |  |


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


### `PUT /api/rgw/user/{uid}`

- 摘要：PUT /api/rgw/user/{uid}
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: account_id, account_policies, account_root_user, daemon_name, display_name, email, max_buckets, suspended, system; required: 无

```yaml
properties:
  account_id:
    type: integer
  account_policies:
    type: integer
  account_root_user:
    default: false
    type: integer
  daemon_name:
    type: string
  display_name:
    type: string
  email:
    type: string
  max_buckets:
    type: string
  suspended:
    type: string
  system:
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


### `DELETE /api/rgw/user/{uid}`

- 摘要：DELETE /api/rgw/user/{uid}
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |
| daemon_name | query | 否 | string |  |  |


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


### `POST /api/rgw/user/{uid}/capability`

- 摘要：POST /api/rgw/user/{uid}/capability
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: daemon_name, perm, type; required: type, perm

```yaml
properties:
  daemon_name:
    type: string
  perm:
    type: string
  type:
    type: string
required:
- type
- perm
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


### `DELETE /api/rgw/user/{uid}/capability`

- 摘要：DELETE /api/rgw/user/{uid}/capability
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |
| type | query | 是 | string |  |  |
| perm | query | 是 | string |  |  |
| daemon_name | query | 否 | string |  |  |


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


### `POST /api/rgw/user/{uid}/key`

- 摘要：POST /api/rgw/user/{uid}/key
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: access_key, daemon_name, generate_key, key_type, secret_key, subuser; required: 无

```yaml
properties:
  access_key:
    type: string
  daemon_name:
    type: string
  generate_key:
    default: 'true'
    type: string
  key_type:
    default: s3
    type: string
  secret_key:
    type: string
  subuser:
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


### `DELETE /api/rgw/user/{uid}/key`

- 摘要：DELETE /api/rgw/user/{uid}/key
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |
| key_type | query | 否 | string | "s3" |  |
| subuser | query | 否 | string |  |  |
| access_key | query | 否 | string |  |  |
| daemon_name | query | 否 | string |  |  |


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


### `GET /api/rgw/user/{uid}/quota`

- 摘要：GET /api/rgw/user/{uid}/quota
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |
| daemon_name | query | 否 | string |  |  |


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


### `PUT /api/rgw/user/{uid}/quota`

- 摘要：PUT /api/rgw/user/{uid}/quota
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: daemon_name, enabled, max_objects, max_size_kb, quota_type; required: quota_type, enabled, max_size_kb, max_objects

```yaml
properties:
  daemon_name:
    type: string
  enabled:
    type: string
  max_objects:
    type: string
  max_size_kb:
    type: integer
  quota_type:
    type: string
required:
- quota_type
- enabled
- max_size_kb
- max_objects
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


### `GET /api/rgw/user/{uid}/ratelimit`

- 摘要：Get the user rate limit
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |


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


### `PUT /api/rgw/user/{uid}/ratelimit`

- 摘要：Update the user rate limit
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: enabled, max_read_bytes, max_read_ops, max_write_bytes, max_write_ops; required: 无

```yaml
properties:
  enabled:
    default: false
    type: boolean
  max_read_bytes:
    default: 0
    type: integer
  max_read_ops:
    default: 0
    type: integer
  max_write_bytes:
    default: 0
    type: integer
  max_write_ops:
    default: 0
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


### `POST /api/rgw/user/{uid}/subuser`

- 摘要：POST /api/rgw/user/{uid}/subuser
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: access, access_key, daemon_name, generate_secret, key_type, secret_key, subuser; required: subuser, access

```yaml
properties:
  access:
    type: string
  access_key:
    type: string
  daemon_name:
    type: string
  generate_secret:
    default: 'true'
    type: string
  key_type:
    default: s3
    type: string
  secret_key:
    type: string
  subuser:
    type: string
required:
- subuser
- access
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


### `DELETE /api/rgw/user/{uid}/subuser/{subuser}`

- 摘要：DELETE /api/rgw/user/{uid}/subuser/{subuser}
- Tags：`RgwUser`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| uid | path | 是 | string |  |  |
| subuser | path | 是 | string |  |  |
| purge_keys | query | 否 | string | "true" |  |
| daemon_name | query | 否 | string |  |  |


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


### `GET /api/rgw/zone`

- 摘要：GET /api/rgw/zone
- Tags：`RgwZone`
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


### `POST /api/rgw/zone`

- 摘要：POST /api/rgw/zone
- Tags：`RgwZone`
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
- Schema: object; fields: access_key, default, master, secret_key, sync_from, sync_from_all, tier_type, zone_endpoints, zone_name, zonegroup_name; required: zone_name

```yaml
properties:
  access_key:
    type: string
  default:
    default: false
    type: boolean
  master:
    default: false
    type: boolean
  secret_key:
    type: string
  sync_from:
    default: ''
    type: string
  sync_from_all:
    default: ''
    type: string
  tier_type:
    default: ''
    type: string
  zone_endpoints:
    type: string
  zone_name:
    type: string
  zonegroup_name:
    type: string
required:
- zone_name
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


### `PUT /api/rgw/zone/create_system_user`

- 摘要：PUT /api/rgw/zone/create_system_user
- Tags：`RgwZone`
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
- Schema: object; fields: userName, zoneName; required: userName, zoneName

```yaml
properties:
  userName:
    type: string
  zoneName:
    type: string
required:
- userName
- zoneName
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


### `GET /api/rgw/zone/get_all_zones_info`

- 摘要：GET /api/rgw/zone/get_all_zones_info
- Tags：`RgwZone`
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


### `GET /api/rgw/zone/get_pool_names`

- 摘要：GET /api/rgw/zone/get_pool_names
- Tags：`RgwZone`
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


### `GET /api/rgw/zone/get_user_list`

- 摘要：GET /api/rgw/zone/get_user_list
- Tags：`RgwZone`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| zoneName | query | 否 | string |  |  |
| realmName | query | 否 | string |  |  |


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


### `POST /api/rgw/zone/storage-class`

- 摘要：POST /api/rgw/zone/storage-class
- Tags：`RgwZone`
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
- Schema: object; fields: compression, data_pool, placement_target, storage_class, zone_name; required: zone_name, placement_target, storage_class, data_pool

```yaml
properties:
  compression:
    default: ''
    type: string
  data_pool:
    type: string
  placement_target:
    type: string
  storage_class:
    type: string
  zone_name:
    type: string
required:
- zone_name
- placement_target
- storage_class
- data_pool
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


### `PUT /api/rgw/zone/storage-class`

- 摘要：PUT /api/rgw/zone/storage-class
- Tags：`RgwZone`
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
- Schema: object; fields: compression, data_pool, placement_target, storage_class, zone_name; required: zone_name, placement_target, storage_class, data_pool

```yaml
properties:
  compression:
    default: ''
    type: string
  data_pool:
    type: string
  placement_target:
    type: string
  storage_class:
    type: string
  zone_name:
    type: string
required:
- zone_name
- placement_target
- storage_class
- data_pool
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


### `GET /api/rgw/zone/{zone_name}`

- 摘要：GET /api/rgw/zone/{zone_name}
- Tags：`RgwZone`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| zone_name | path | 是 | string |  |  |


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


### `PUT /api/rgw/zone/{zone_name}`

- 摘要：PUT /api/rgw/zone/{zone_name}
- Tags：`RgwZone`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| zone_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: access_key, compression, data_extra_pool, data_pool, data_pool_class, default, index_pool, master, new_zone_name, placement_target, secret_key, storage_class, sync_from, sync_from_all, tier_type, zone_endpoints, zonegroup_name; required: new_zone_name, zonegroup_name

```yaml
properties:
  access_key:
    default: ''
    type: string
  compression:
    default: ''
    type: string
  data_extra_pool:
    default: ''
    type: string
  data_pool:
    default: ''
    type: string
  data_pool_class:
    default: ''
    type: string
  default:
    default: ''
    type: string
  index_pool:
    default: ''
    type: string
  master:
    default: ''
    type: string
  new_zone_name:
    type: string
  placement_target:
    default: ''
    type: string
  secret_key:
    default: ''
    type: string
  storage_class:
    default: ''
    type: string
  sync_from:
    default: ''
    type: string
  sync_from_all:
    default: ''
    type: string
  tier_type:
    default: ''
    type: string
  zone_endpoints:
    default: ''
    type: string
  zonegroup_name:
    type: string
required:
- new_zone_name
- zonegroup_name
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


### `DELETE /api/rgw/zone/{zone_name}`

- 摘要：DELETE /api/rgw/zone/{zone_name}
- Tags：`RgwZone`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| zone_name | path | 是 | string |  |  |
| delete_pools | query | 是 | string |  |  |
| pools | query | 否 | string |  |  |
| zonegroup_name | query | 否 | string |  |  |
| realm_name | query | 否 | string |  |  |


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


### `GET /api/rgw/zonegroup`

- 摘要：GET /api/rgw/zonegroup
- Tags：`RgwZonegroup`
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


### `POST /api/rgw/zonegroup`

- 摘要：POST /api/rgw/zonegroup
- Tags：`RgwZonegroup`
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
- Schema: object; fields: default, master, realm_name, zonegroup_endpoints, zonegroup_name; required: realm_name, zonegroup_name

```yaml
properties:
  default:
    type: string
  master:
    type: string
  realm_name:
    type: string
  zonegroup_endpoints:
    type: string
  zonegroup_name:
    type: string
required:
- realm_name
- zonegroup_name
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


### `GET /api/rgw/zonegroup/get_all_zonegroups_info`

- 摘要：GET /api/rgw/zonegroup/get_all_zonegroups_info
- Tags：`RgwZonegroup`
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


### `GET /api/rgw/zonegroup/get_placement_target_by_placement_id/{placement_id}`

- 摘要：GET /api/rgw/zonegroup/get_placement_target_by_placement_id/{placement_id}
- Tags：`RgwZonegroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| placement_id | path | 是 | string |  |  |


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


### `POST /api/rgw/zonegroup/storage-class`

- 摘要：POST /api/rgw/zonegroup/storage-class
- Tags：`RgwZonegroup`
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
- Schema: object; fields: placement_targets, zone_group; required: zone_group

```yaml
properties:
  placement_targets:
    default: []
    type: string
  zone_group:
    type: string
required:
- zone_group
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


### `PUT /api/rgw/zonegroup/storage-class`

- 摘要：PUT /api/rgw/zonegroup/storage-class
- Tags：`RgwZonegroup`
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
- Schema: object; fields: placement_targets, zone_group; required: zone_group

```yaml
properties:
  placement_targets:
    default: []
    type: string
  zone_group:
    type: string
required:
- zone_group
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


### `DELETE /api/rgw/zonegroup/storage-class/{placement_id}/{storage_class}`

- 摘要：DELETE /api/rgw/zonegroup/storage-class/{placement_id}/{storage_class}
- Tags：`RgwZonegroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| placement_id | path | 是 | string |  |  |
| storage_class | path | 是 | string |  |  |
| zone_name | query | 否 | string | "" |  |


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


### `GET /api/rgw/zonegroup/{zonegroup_name}`

- 摘要：GET /api/rgw/zonegroup/{zonegroup_name}
- Tags：`RgwZonegroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| zonegroup_name | path | 是 | string |  |  |


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


### `PUT /api/rgw/zonegroup/{zonegroup_name}`

- 摘要：PUT /api/rgw/zonegroup/{zonegroup_name}
- Tags：`RgwZonegroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| zonegroup_name | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: add_zones, default, master, new_zonegroup_name, placement_targets, realm_name, remove_zones, zonegroup_endpoints; required: realm_name, new_zonegroup_name

```yaml
properties:
  add_zones:
    default: []
    type: string
  default:
    default: ''
    type: string
  master:
    default: ''
    type: string
  new_zonegroup_name:
    type: string
  placement_targets:
    default: []
    type: string
  realm_name:
    type: string
  remove_zones:
    default: []
    type: string
  zonegroup_endpoints:
    default: ''
    type: string
required:
- realm_name
- new_zonegroup_name
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


### `DELETE /api/rgw/zonegroup/{zonegroup_name}`

- 摘要：DELETE /api/rgw/zonegroup/{zonegroup_name}
- Tags：`RgwZonegroup`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v18.2.8
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| zonegroup_name | path | 是 | string |  |  |
| delete_pools | query | 是 | string |  |  |
| pools | query | 否 | string |  |  |
| realm_name | query | 否 | string |  |  |


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

