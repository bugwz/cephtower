# Ceph 20.2.2 Dashboard API - NVMe-oF

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/nvmeof` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/nvmeof/gateway`](#get-api-nvmeof-gateway) - Get information about the NVMeoF gateway
- [`GET /api/nvmeof/gateway/group`](#get-api-nvmeof-gateway-group) - GET /api/nvmeof/gateway/group
- [`GET /api/nvmeof/gateway/listener_info/{nqn}`](#get-api-nvmeof-gateway-listener-info-nqn) - Get NVMeoF gateway's listeners info
- [`GET /api/nvmeof/gateway/log_level`](#get-api-nvmeof-gateway-log-level) - Get NVMeoF gateway log level information
- [`PUT /api/nvmeof/gateway/log_level`](#put-api-nvmeof-gateway-log-level) - Set NVMeoF gateway log levels
- [`GET /api/nvmeof/gateway/stats`](#get-api-nvmeof-gateway-stats) - Get NVMeoF statistics for the gateway
- [`GET /api/nvmeof/gateway/version`](#get-api-nvmeof-gateway-version) - Get the version of the NVMeoF gateway
- [`GET /api/nvmeof/spdk/log_level`](#get-api-nvmeof-spdk-log-level) - Get NVMeoF gateway spdk log levels
- [`PUT /api/nvmeof/spdk/log_level`](#put-api-nvmeof-spdk-log-level) - Set NVMeoF gateway spdk log levels
- [`PUT /api/nvmeof/spdk/log_level/disable`](#put-api-nvmeof-spdk-log-level-disable) - Disable NVMeoF gateway spdk log
- [`GET /api/nvmeof/subsystem`](#get-api-nvmeof-subsystem) - List all NVMeoF subsystems
- [`POST /api/nvmeof/subsystem`](#post-api-nvmeof-subsystem) - Create a new NVMeoF subsystem
- [`GET /api/nvmeof/subsystem/{nqn}`](#get-api-nvmeof-subsystem-nqn) - Get information from a specific NVMeoF subsystem
- [`DELETE /api/nvmeof/subsystem/{nqn}`](#delete-api-nvmeof-subsystem-nqn) - Delete an existing NVMeoF subsystem
- [`GET /api/nvmeof/subsystem/{nqn}/connection`](#get-api-nvmeof-subsystem-nqn-connection) - List all NVMeoF Subsystem Connections
- [`GET /api/nvmeof/subsystem/{nqn}/host`](#get-api-nvmeof-subsystem-nqn-host) - List all allowed hosts for an NVMeoF subsystem
- [`POST /api/nvmeof/subsystem/{nqn}/host`](#post-api-nvmeof-subsystem-nqn-host) - Allow hosts to access an NVMeoF subsystem
- [`DELETE /api/nvmeof/subsystem/{nqn}/host/{host_nqn}`](#delete-api-nvmeof-subsystem-nqn-host-host-nqn) - Disallow hosts from accessing an NVMeoF subsystem
- [`PUT /api/nvmeof/subsystem/{nqn}/host/{host_nqn}/change_controller_key`](#put-api-nvmeof-subsystem-nqn-host-host-nqn-change-controller-key) - Change host DH-HMAC-CHAP controller key
- [`PUT /api/nvmeof/subsystem/{nqn}/host/{host_nqn}/change_key`](#put-api-nvmeof-subsystem-nqn-host-host-nqn-change-key) - Change host DH-HMAC-CHAP key
- [`PUT /api/nvmeof/subsystem/{nqn}/host/{host_nqn}/del_controller_key`](#put-api-nvmeof-subsystem-nqn-host-host-nqn-del-controller-key) - Delete host DH-HMAC-CHAP controller key
- [`PUT /api/nvmeof/subsystem/{nqn}/host/{host_nqn}/del_key`](#put-api-nvmeof-subsystem-nqn-host-host-nqn-del-key) - Delete host DH-HMAC-CHAP key
- [`GET /api/nvmeof/subsystem/{nqn}/listener`](#get-api-nvmeof-subsystem-nqn-listener) - List all NVMeoF listeners
- [`POST /api/nvmeof/subsystem/{nqn}/listener`](#post-api-nvmeof-subsystem-nqn-listener) - Create a new NVMeoF listener
- [`DELETE /api/nvmeof/subsystem/{nqn}/listener/{host_name}/{traddr}/{trsvcid}`](#delete-api-nvmeof-subsystem-nqn-listener-host-name-traddr-trsvcid) - Delete an existing NVMeoF listener
- [`GET /api/nvmeof/subsystem/{nqn}/namespace`](#get-api-nvmeof-subsystem-nqn-namespace) - List all NVMeoF namespaces in a subsystem
- [`POST /api/nvmeof/subsystem/{nqn}/namespace`](#post-api-nvmeof-subsystem-nqn-namespace) - Create a new NVMeoF namespace.
- [`GET /api/nvmeof/subsystem/{nqn}/namespace/{nsid}`](#get-api-nvmeof-subsystem-nqn-namespace-nsid) - Get info from specified NVMeoF namespace
- [`PATCH /api/nvmeof/subsystem/{nqn}/namespace/{nsid}`](#patch-api-nvmeof-subsystem-nqn-namespace-nsid) - Update an existing NVMeoF namespace
- [`DELETE /api/nvmeof/subsystem/{nqn}/namespace/{nsid}`](#delete-api-nvmeof-subsystem-nqn-namespace-nsid) - Delete an existing NVMeoF namespace
- [`PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/add_host`](#put-api-nvmeof-subsystem-nqn-namespace-nsid-add-host) - Adds a host to the specified NVMeoF namespace
- [`PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/change_load_balancing_group`](#put-api-nvmeof-subsystem-nqn-namespace-nsid-change-load-balancing-group) - set the load balancing group for specified NVMeoF namespace
- [`PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/change_visibility`](#put-api-nvmeof-subsystem-nqn-namespace-nsid-change-visibility) - changes the visibility of the specified NVMeoF namespace to all or selected hosts
- [`PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/del_host`](#put-api-nvmeof-subsystem-nqn-namespace-nsid-del-host) - Removes a host from the specified NVMeoF namespace
- [`GET /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/io_stats`](#get-api-nvmeof-subsystem-nqn-namespace-nsid-io-stats) - Get IO stats from specified NVMeoF namespace
- [`PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/refresh_size`](#put-api-nvmeof-subsystem-nqn-namespace-nsid-refresh-size) - refresh the specified NVMeoF namespace to current RBD image size
- [`PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/resize`](#put-api-nvmeof-subsystem-nqn-namespace-nsid-resize) - resize the specified NVMeoF namespace
- [`PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/set_auto_resize`](#put-api-nvmeof-subsystem-nqn-namespace-nsid-set-auto-resize) - Enable or disable namespace auto resize when RBD image is resized
- [`PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/set_qos`](#put-api-nvmeof-subsystem-nqn-namespace-nsid-set-qos) - set QOS for specified NVMeoF namespace
- [`PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/set_rbd_trash_image`](#put-api-nvmeof-subsystem-nqn-namespace-nsid-set-rbd-trash-image) - changes the trash image on delete of the specified NVMeoF                 namespace to all or selected hosts

## 接口详情

### `GET /api/nvmeof/gateway`

- 摘要：Get information about the NVMeoF gateway
- Tags：`NVMe-oF Gateway`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| gw_group | query | 否 | string |  |  |
| traddr | query | 否 | string |  |  |


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


### `GET /api/nvmeof/gateway/group`

- 摘要：GET /api/nvmeof/gateway/group
- Tags：`NVMe-oF Gateway`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
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


### `GET /api/nvmeof/gateway/listener_info/{nqn}`

- 摘要：Get NVMeoF gateway's listeners info
- Tags：`NVMe-oF Gateway`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  |  |
| gw_group | query | 否 | string |  |  |
| traddr | query | 否 | string |  |  |


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


### `GET /api/nvmeof/gateway/log_level`

- 摘要：Get NVMeoF gateway log level information
- Tags：`NVMe-oF Gateway`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| gw_group | query | 否 | string |  |  |
| traddr | query | 否 | string |  |  |


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


### `PUT /api/nvmeof/gateway/log_level`

- 摘要：Set NVMeoF gateway log levels
- Tags：`NVMe-oF Gateway`
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
- Schema: object; fields: gw_group, log_level, traddr; required: log_level

```yaml
properties:
  gw_group:
    type: string
  log_level:
    type: string
  traddr:
    type: string
required:
- log_level
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


### `GET /api/nvmeof/gateway/stats`

- 摘要：Get NVMeoF statistics for the gateway
- Tags：`NVMe-oF Gateway`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| gw_group | query | 否 | string |  |  |
| traddr | query | 否 | string |  |  |


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


### `GET /api/nvmeof/gateway/version`

- 摘要：Get the version of the NVMeoF gateway
- Tags：`NVMe-oF Gateway`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| gw_group | query | 否 | string |  |  |
| traddr | query | 否 | string |  |  |


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


### `GET /api/nvmeof/spdk/log_level`

- 摘要：Get NVMeoF gateway spdk log levels
- Tags：`NVMe-oF SPDK`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| all_log_flags | query | 否 | string |  |  |
| gw_group | query | 否 | string |  |  |
| traddr | query | 否 | string |  |  |


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


### `PUT /api/nvmeof/spdk/log_level`

- 摘要：Set NVMeoF gateway spdk log levels
- Tags：`NVMe-oF SPDK`
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
- Schema: object; fields: extra_log_flags, gw_group, log_level, print_level, traddr; required: 无

```yaml
properties:
  extra_log_flags:
    type: string
  gw_group:
    type: string
  log_level:
    type: string
  print_level:
    type: string
  traddr:
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


### `PUT /api/nvmeof/spdk/log_level/disable`

- 摘要：Disable NVMeoF gateway spdk log
- Tags：`NVMe-oF SPDK`
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
- Schema: object; fields: extra_log_flags, gw_group, traddr; required: 无

```yaml
properties:
  extra_log_flags:
    type: string
  gw_group:
    type: string
  traddr:
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


### `GET /api/nvmeof/subsystem`

- 摘要：List all NVMeoF subsystems
- Tags：`NVMe-oF Subsystem`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| gw_group | query | 否 | string |  |  |
| traddr | query | 否 | string |  |  |


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


### `POST /api/nvmeof/subsystem`

- 摘要：Create a new NVMeoF subsystem
- Tags：`NVMe-oF Subsystem`
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
- Schema: object; fields: dhchap_key, enable_ha, gw_group, max_namespaces, network_mask, no_group_append, nqn, port, secure_listeners, serial_number, traddr; required: nqn

```yaml
properties:
  dhchap_key:
    description: Subsystem DH-HMAC-CHAP key
    type: string
  enable_ha:
    default: true
    description: Enable high availability
    type: boolean
  gw_group:
    description: NVMeoF gateway group
    type: string
  max_namespaces:
    description: Maximum number of namespaces
    type: integer
  network_mask:
    description: Network mask to automatically create listeners
    items:
      type: string
    type: array
  no_group_append:
    default: false
    description: Do not append gateway group name to the NQN
    type: integer
  nqn:
    description: NVMeoF subsystem NQN
    type: string
  port:
    description: Port to use for the created listeners
    type: integer
  secure_listeners:
    default: false
    description: Make all the auto-listeners for this subsystem secure
    type: boolean
  serial_number:
    description: Subsystem serial number
    type: string
  traddr:
    description: NVMeoF gateway address
    type: string
required:
- nqn
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


### `GET /api/nvmeof/subsystem/{nqn}`

- 摘要：Get information from a specific NVMeoF subsystem
- Tags：`NVMe-oF Subsystem`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| gw_group | query | 否 | string |  | NVMeoF gateway group |
| traddr | query | 否 | string |  |  |


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


### `DELETE /api/nvmeof/subsystem/{nqn}`

- 摘要：Delete an existing NVMeoF subsystem
- Tags：`NVMe-oF Subsystem`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| force | query | 否 | boolean | "false" | Force delete |
| gw_group | query | 否 | string |  | NVMeoF gateway group |
| traddr | query | 否 | string |  |  |


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


### `GET /api/nvmeof/subsystem/{nqn}/connection`

- 摘要：List all NVMeoF Subsystem Connections
- Tags：`NVMe-oF Subsystem Connection`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 否 | string |  | NVMeoF subsystem NQN |
| gw_group | query | 否 | string |  | NVMeoF gateway group |
| traddr | query | 否 | string |  |  |


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


### `GET /api/nvmeof/subsystem/{nqn}/host`

- 摘要：List all allowed hosts for an NVMeoF subsystem
- Tags：`NVMe-oF Subsystem Host Allowlist`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| clear_alerts | query | 否 | boolean |  | Clear any host alert signal after getting its value |
| gw_group | query | 否 | string |  | NVMeoF gateway group |
| traddr | query | 否 | string |  |  |


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


### `POST /api/nvmeof/subsystem/{nqn}/host`

- 摘要：Allow hosts to access an NVMeoF subsystem
- Tags：`NVMe-oF Subsystem Host Allowlist`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: dhchap_controller_key, dhchap_key, gw_group, host_nqn, psk, traddr; required: host_nqn

```yaml
properties:
  dhchap_controller_key:
    type: string
  dhchap_key:
    type: string
  gw_group:
    description: NVMeoF gateway group
    type: string
  host_nqn:
    description: NVMeoF host NQN. Use "*" to allow any host.
    type: string
  psk:
    type: string
  traddr:
    type: string
required:
- host_nqn
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


### `DELETE /api/nvmeof/subsystem/{nqn}/host/{host_nqn}`

- 摘要：Disallow hosts from accessing an NVMeoF subsystem
- Tags：`NVMe-oF Subsystem Host Allowlist`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| host_nqn | path | 是 | string |  | NVMeoF host NQN. Use "*" to disallow any host. |
| gw_group | query | 否 | string |  | NVMeoF gateway group |
| traddr | query | 否 | string |  |  |


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


### `PUT /api/nvmeof/subsystem/{nqn}/host/{host_nqn}/change_controller_key`

- 摘要：Change host DH-HMAC-CHAP controller key
- Tags：`NVMe-oF Subsystem Host Allowlist`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| host_nqn | path | 是 | string |  | NVMeoF host NQN |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: dhchap_controller_key, gw_group, traddr; required: dhchap_controller_key

```yaml
properties:
  dhchap_controller_key:
    description: Host DH-HMAC-CHAP controller key
    type: string
  gw_group:
    description: NVMeoF gateway group
    type: string
  traddr:
    type: string
required:
- dhchap_controller_key
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


### `PUT /api/nvmeof/subsystem/{nqn}/host/{host_nqn}/change_key`

- 摘要：Change host DH-HMAC-CHAP key
- Tags：`NVMe-oF Subsystem Host Allowlist`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| host_nqn | path | 是 | string |  | NVMeoF host NQN |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: dhchap_key, gw_group, traddr; required: dhchap_key

```yaml
properties:
  dhchap_key:
    description: Host DH-HMAC-CHAP key
    type: string
  gw_group:
    description: NVMeoF gateway group
    type: string
  traddr:
    type: string
required:
- dhchap_key
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


### `PUT /api/nvmeof/subsystem/{nqn}/host/{host_nqn}/del_controller_key`

- 摘要：Delete host DH-HMAC-CHAP controller key
- Tags：`NVMe-oF Subsystem Host Allowlist`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| host_nqn | path | 是 | string |  | NVMeoF host NQN. |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: gw_group, traddr; required: 无

```yaml
properties:
  gw_group:
    description: NVMeoF gateway group
    type: string
  traddr:
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


### `PUT /api/nvmeof/subsystem/{nqn}/host/{host_nqn}/del_key`

- 摘要：Delete host DH-HMAC-CHAP key
- Tags：`NVMe-oF Subsystem Host Allowlist`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| host_nqn | path | 是 | string |  | NVMeoF host NQN. |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: gw_group, traddr; required: 无

```yaml
properties:
  gw_group:
    description: NVMeoF gateway group
    type: string
  traddr:
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


### `GET /api/nvmeof/subsystem/{nqn}/listener`

- 摘要：List all NVMeoF listeners
- Tags：`NVMe-oF Subsystem Listener`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| gw_group | query | 否 | string |  | NVMeoF gateway group |
| traddr | query | 否 | string |  |  |


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


### `POST /api/nvmeof/subsystem/{nqn}/listener`

- 摘要：Create a new NVMeoF listener
- Tags：`NVMe-oF Subsystem Listener`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: adrfam, gw_group, host_name, secure, traddr, trsvcid, verify_host_name; required: host_name, traddr

```yaml
properties:
  adrfam:
    default: 0
    description: NVMeoF address family (0 - IPv4, 1 - IPv6)
    type: integer
  gw_group:
    description: NVMeoF gateway group
    type: string
  host_name:
    description: NVMeoF hostname
    type: string
  secure:
    default: false
    description: Use a secure channel
    type: boolean
  traddr:
    description: NVMeoF transport address
    type: string
  trsvcid:
    description: NVMeoF transport service port
    type: integer
  verify_host_name:
    default: false
    description: Fail if the host name doesn't match the gateway's host name
    type: boolean
required:
- host_name
- traddr
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


### `DELETE /api/nvmeof/subsystem/{nqn}/listener/{host_name}/{traddr}/{trsvcid}`

- 摘要：Delete an existing NVMeoF listener
- Tags：`NVMe-oF Subsystem Listener`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| host_name | path | 是 | string |  | NVMeoF hostname |
| traddr | path | 是 | string |  | NVMeoF transport address |
| trsvcid | path | 是 | integer |  | NVMeoF transport service port |
| adrfam | query | 否 | integer | 0 | NVMeoF address family (0 - IPv4, 1 - IPv6) |
| force | query | 否 | boolean | false |  |
| gw_group | query | 否 | string |  | NVMeoF gateway group |


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


### `GET /api/nvmeof/subsystem/{nqn}/namespace`

- 摘要：List all NVMeoF namespaces in a subsystem
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 否 | string |  | NVMeoF subsystem NQN |
| nsid | query | 否 | string |  | NVMeoF Namespace ID to filter by |
| gw_group | query | 否 | string |  | NVMeoF gateway group |
| traddr | query | 否 | string |  |  |


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


### `POST /api/nvmeof/subsystem/{nqn}/namespace`

- 摘要：Create a new NVMeoF namespace.
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: block_size, create_image, disable_auto_resize, encryption_algorithm, encryption_format, force, gw_group, key_id, load_balancing_group, no_auto_visible, nsid, rados_namespace, rbd_image_name, rbd_image_size, rbd_pool, read_only, size, traddr, trash_image; required: rbd_image_name

```yaml
properties:
  block_size:
    default: 512
    description: NVMeoF namespace block size
    type: integer
  create_image:
    default: false
    description: Create RBD image
    type: boolean
  disable_auto_resize:
    default: false
    description: Disable auto resize
    type: string
  encryption_algorithm:
    description: Algorithm to use for encryption
    type: string
  encryption_format:
    description: Encryption format(s) to use, LUKS1 or LUKS2, separated by commas
    items:
      type: string
    type: array
  force:
    default: false
    description: Force create namespace even it image is used by other namespace
    type: boolean
  gw_group:
    description: NVMeoF gateway group
    type: string
  key_id:
    description: Key ID(s) to use for encryption pass phrases, separated by commas
    items:
      type: string
    type: array
  load_balancing_group:
    description: Load balancing group
    type: integer
  no_auto_visible:
    default: false
    description: Namespace will be visible only for the allowed hosts
    type: boolean
  nsid:
    type: string
  rados_namespace:
    description: RADOS namespace name
    type: string
  rbd_image_name:
    description: RBD image name
    type: string
  rbd_image_size:
    description: RBD image size
    type: integer
  rbd_pool:
    default: rbd
    description: RBD pool name
    type: string
  read_only:
    default: false
    description: Read only namespace
    type: string
  size:
    description: Deprecated. Use `rbd_image_size` instead
    type: integer
  traddr:
    description: Target gateway address
    type: string
  trash_image:
    default: false
    description: Trash the RBD image when namespace is removed
    type: boolean
required:
- rbd_image_name
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


### `GET /api/nvmeof/subsystem/{nqn}/namespace/{nsid}`

- 摘要：Get info from specified NVMeoF namespace
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |
| gw_group | query | 否 | string |  | NVMeoF gateway group |
| traddr | query | 否 | string |  |  |


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


### `PATCH /api/nvmeof/subsystem/{nqn}/namespace/{nsid}`

- 摘要：Update an existing NVMeoF namespace
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: gw_group, load_balancing_group, r_mbytes_per_second, rbd_image_size, rw_ios_per_second, rw_mbytes_per_second, traddr, trash_image, w_mbytes_per_second; required: 无

```yaml
properties:
  gw_group:
    description: NVMeoF gateway group
    type: string
  load_balancing_group:
    description: Load balancing group
    type: integer
  r_mbytes_per_second:
    description: Read MB/s
    type: integer
  rbd_image_size:
    description: RBD image size
    type: integer
  rw_ios_per_second:
    description: Read/Write IOPS
    type: integer
  rw_mbytes_per_second:
    description: Read/Write MB/s
    type: integer
  traddr:
    type: string
  trash_image:
    description: Trash RBD image after removing namespace
    type: boolean
  w_mbytes_per_second:
    description: Write MB/s
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


### `DELETE /api/nvmeof/subsystem/{nqn}/namespace/{nsid}`

- 摘要：Delete an existing NVMeoF namespace
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |
| gw_group | query | 否 | string |  | NVMeoF gateway group |
| traddr | query | 否 | string |  |  |
| force | query | 否 | string | "false" | Force remove the RBD image |


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


### `PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/add_host`

- 摘要：Adds a host to the specified NVMeoF namespace
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: force, gw_group, host_nqn, traddr; required: host_nqn

```yaml
properties:
  force:
    description: Allow adding the host to the namespace even if the host has no access
      to the subsystem
    type: boolean
  gw_group:
    description: NVMeoF gateway group
    type: string
  host_nqn:
    description: NVMeoF host NQN. Use "*" to allow any host.
    type: string
  traddr:
    description: NVMeoF gateway address
    type: string
required:
- host_nqn
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


### `PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/change_load_balancing_group`

- 摘要：set the load balancing group for specified NVMeoF namespace
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: gw_group, load_balancing_group, traddr; required: 无

```yaml
properties:
  gw_group:
    description: NVMeoF gateway group
    type: string
  load_balancing_group:
    description: Load balancing group
    type: integer
  traddr:
    description: NVMeoF gateway address
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


### `PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/change_visibility`

- 摘要：changes the visibility of the specified NVMeoF namespace to all or selected hosts
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: auto_visible, force, gw_group, traddr; required: auto_visible

```yaml
properties:
  auto_visible:
    description: True if visible to all hosts
    type: boolean
  force:
    default: false
    type: boolean
  gw_group:
    description: NVMeoF gateway group
    type: string
  traddr:
    description: NVMeoF gateway address
    type: string
required:
- auto_visible
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


### `PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/del_host`

- 摘要：Removes a host from the specified NVMeoF namespace
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: gw_group, host_nqn, traddr; required: host_nqn

```yaml
properties:
  gw_group:
    description: NVMeoF gateway group
    type: string
  host_nqn:
    description: NVMeoF host NQN. Use "*" to allow any host.
    type: string
  traddr:
    description: NVMeoF gateway address
    type: string
required:
- host_nqn
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


### `GET /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/io_stats`

- 摘要：Get IO stats from specified NVMeoF namespace
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v19.2.5, v20.2.2
- 首次出现在扫描范围：v19.2.5
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |
| gw_group | query | 否 | string |  | NVMeoF gateway group |
| traddr | query | 否 | string |  |  |


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


### `PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/refresh_size`

- 摘要：refresh the specified NVMeoF namespace to current RBD image size
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: gw_group, traddr; required: 无

```yaml
properties:
  gw_group:
    description: NVMeoF gateway group
    type: string
  traddr:
    description: NVMeoF gateway address
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


### `PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/resize`

- 摘要：resize the specified NVMeoF namespace
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: gw_group, rbd_image_size, traddr; required: rbd_image_size

```yaml
properties:
  gw_group:
    description: NVMeoF gateway group
    type: string
  rbd_image_size:
    description: RBD image size
    type: integer
  traddr:
    description: NVMeoF gateway address
    type: string
required:
- rbd_image_size
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


### `PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/set_auto_resize`

- 摘要：Enable or disable namespace auto resize when RBD image is resized
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: auto_resize_enabled, gw_group, traddr; required: auto_resize_enabled

```yaml
properties:
  auto_resize_enabled:
    description: Enable or disable auto resize of namespace when RBD image is resized
    type: boolean
  gw_group:
    description: NVMeoF gateway group
    type: string
  traddr:
    description: NVMeoF gateway address
    type: string
required:
- auto_resize_enabled
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


### `PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/set_qos`

- 摘要：set QOS for specified NVMeoF namespace
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: force, gw_group, r_mbytes_per_second, rw_ios_per_second, rw_mbytes_per_second, traddr, w_mbytes_per_second; required: 无

```yaml
properties:
  force:
    default: false
    description: Set QOS limits even if they were changed by RBD
    type: boolean
  gw_group:
    description: NVMeoF gateway group
    type: string
  r_mbytes_per_second:
    description: Read MB/s
    type: integer
  rw_ios_per_second:
    description: Read/Write IOPS
    type: integer
  rw_mbytes_per_second:
    description: Read/Write MB/s
    type: integer
  traddr:
    description: NVMeoF gateway address
    type: string
  w_mbytes_per_second:
    description: Write MB/s
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


### `PUT /api/nvmeof/subsystem/{nqn}/namespace/{nsid}/set_rbd_trash_image`

- 摘要：changes the trash image on delete of the specified NVMeoF                 namespace to all or selected hosts
- Tags：`NVMe-oF Subsystem Namespace`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v20.2.2
- 首次出现在扫描范围：v20.2.2
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| nqn | path | 是 | string |  | NVMeoF subsystem NQN |
| nsid | path | 是 | string |  | NVMeoF Namespace ID |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: gw_group, rbd_trash_image_on_delete, traddr; required: rbd_trash_image_on_delete

```yaml
properties:
  gw_group:
    description: NVMeoF gateway group
    type: string
  rbd_trash_image_on_delete:
    description: True if active
    type: boolean
  traddr:
    description: NVMeoF gateway address
    type: string
required:
- rbd_trash_image_on_delete
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

