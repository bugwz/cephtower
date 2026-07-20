# Ceph 20.2.2 Dashboard API - 主机

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/host` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/host`](#get-api-host) - List Host Specifications
- [`POST /api/host`](#post-api-host) - POST /api/host
- [`GET /api/host/{hostname}`](#get-api-host-hostname) - GET /api/host/{hostname}
- [`PUT /api/host/{hostname}`](#put-api-host-hostname) - PUT /api/host/{hostname}
- [`DELETE /api/host/{hostname}`](#delete-api-host-hostname) - DELETE /api/host/{hostname}
- [`GET /api/host/{hostname}/daemons`](#get-api-host-hostname-daemons) - GET /api/host/{hostname}/daemons
- [`GET /api/host/{hostname}/devices`](#get-api-host-hostname-devices) - GET /api/host/{hostname}/devices
- [`POST /api/host/{hostname}/identify_device`](#post-api-host-hostname-identify-device) - POST /api/host/{hostname}/identify_device
- [`GET /api/host/{hostname}/inventory`](#get-api-host-hostname-inventory) - Get inventory of a host
- [`GET /api/host/{hostname}/smart`](#get-api-host-hostname-smart) - GET /api/host/{hostname}/smart

## 接口详情

### `GET /api/host`

- 摘要：List Host Specifications
- Tags：`Host`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| sources | query | 否 | string |  | Host Sources |
| facts | query | 否 | boolean | false | Host Facts |
| offset | query | 否 | integer | 0 |  |
| limit | query | 否 | integer | 5 |  |
| search | query | 否 | string | "" |  |
| sort | query | 否 | string | "" |  |
| include_service_instances | query | 否 | boolean | true | Include Service Instances |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.3+json`
- Schema: object; fields: addr, ceph_version, hostname, labels, service_instances, service_type, services, sources, status; required: hostname, services, service_instances, ceph_version, addr, labels, service_type, sources, status

```yaml
properties:
  addr:
    description: Host address
    type: string
  ceph_version:
    description: Ceph version
    type: string
  hostname:
    description: Hostname
    type: string
  labels:
    description: Labels related to the host
    items:
      type: string
    type: array
  service_instances:
    description: Service instances related to the host
    items:
      properties:
        count:
          description: Number of instances of the service
          type: integer
        type:
          description: type of service
          type: string
      required:
      - type
      - count
      type: object
    type: array
  service_type:
    description: ''
    type: string
  services:
    description: Services related to the host
    items:
      properties:
        id:
          description: Service Id
          type: string
        type:
          description: type of service
          type: string
      required:
      - type
      - id
      type: object
    type: array
  sources:
    description: Host Sources
    properties:
      ceph:
        description: ''
        type: boolean
      orchestrator:
        description: ''
        type: boolean
    required:
    - ceph
    - orchestrator
    type: object
  status:
    description: ''
    type: string
required:
- hostname
- services
- service_instances
- ceph_version
- addr
- labels
- service_type
- sources
- status
type: object
```

#### `400`

Operation exception. Please check the response body for details.

无响应体 schema。

#### `401`

Unauthenticated access. Please login first.

无响应体 schema。

#### `403`

Unauthorized access. Please check your permissions.

无响应体 schema。

#### `500`

Unexpected error. Please check the response body for the stack trace.

无响应体 schema。


### `POST /api/host`

- 摘要：POST /api/host
- Tags：`Host`
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
- Schema: object; fields: addr, hostname, labels, status; required: hostname

```yaml
properties:
  addr:
    description: Network Address
    type: string
  hostname:
    description: Hostname
    type: string
  labels:
    description: Host Labels
    items:
      type: string
    type: array
  status:
    description: Host Status
    type: string
required:
- hostname
type: object
```


#### 返回消息

#### `201`

Resource created.

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


### `GET /api/host/{hostname}`

- 摘要：GET /api/host/{hostname}
- Tags：`Host`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| hostname | path | 是 | string |  |  |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.2+json`
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


### `PUT /api/host/{hostname}`

- 摘要：PUT /api/host/{hostname}
- Tags：`Host`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| hostname | path | 是 | string |  | Hostname |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: drain, force, labels, maintenance, update_labels; required: 无

```yaml
properties:
  drain:
    default: false
    description: Drain Host
    type: boolean
  force:
    default: false
    description: Force Enter Maintenance
    type: boolean
  labels:
    description: Host Labels
    items:
      type: string
    type: array
  maintenance:
    default: false
    description: Enter/Exit Maintenance
    type: boolean
  update_labels:
    default: false
    description: Update Labels
    type: boolean
type: object
```


#### 返回消息

#### `200`

Resource updated.

- Content-Type: `application/vnd.ceph.api.v0.1+json`
- Schema: object

```yaml
properties: {}
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


### `DELETE /api/host/{hostname}`

- 摘要：DELETE /api/host/{hostname}
- Tags：`Host`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| hostname | path | 是 | string |  |  |


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


### `GET /api/host/{hostname}/daemons`

- 摘要：GET /api/host/{hostname}/daemons
- Tags：`Host`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| hostname | path | 是 | string |  |  |


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


### `GET /api/host/{hostname}/devices`

- 摘要：GET /api/host/{hostname}/devices
- Tags：`Host`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| hostname | path | 是 | string |  |  |


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


### `POST /api/host/{hostname}/identify_device`

- 摘要：POST /api/host/{hostname}/identify_device
- Tags：`Host`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| hostname | path | 是 | string |  |  |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: device, duration; required: device, duration

```yaml
properties:
  device:
    type: string
  duration:
    type: string
required:
- device
- duration
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


### `GET /api/host/{hostname}/inventory`

- 摘要：Get inventory of a host
- Tags：`Host`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| hostname | path | 是 | string |  | Hostname |
| refresh | query | 否 | string |  | Trigger asynchronous refresh |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: addr, devices, labels, name; required: name, addr, devices, labels

```yaml
properties:
  addr:
    description: Host address
    type: string
  devices:
    description: Host devices
    items:
      properties:
        available:
          description: If the device can be provisioned to an OSD
          type: boolean
        device_id:
          description: Device's udev ID
          type: string
        human_readable_type:
          description: Device type. ssd or hdd
          type: string
        lsm_data:
          description: ''
          properties:
            errors:
              description: ''
              items:
                type: string
              type: array
            health:
              description: ''
              type: string
            ledSupport:
              description: ''
              properties:
                FAILstatus:
                  description: ''
                  type: string
                FAILsupport:
                  description: ''
                  type: string
                IDENTstatus:
                  description: ''
                  type: string
                IDENTsupport:
                  description: ''
                  type: string
              required:
              - IDENTsupport
              - IDENTstatus
              - FAILsupport
              - FAILstatus
              type: object
            linkSpeed:
              description: ''
              type: string
            mediaType:
              description: ''
              type: string
            rpm:
              description: ''
              type: string
            serialNum:
              description: ''
              type: string
            transport:
              description: ''
              type: string
          required:
          - serialNum
          - transport
          - mediaType
          - rpm
          - linkSpeed
          - health
          - ledSupport
          - errors
          type: object
        lvs:
          description: ''
          items:
            properties:
              block_uuid:
                description: ''
                type: string
              cluster_fsid:
                description: ''
                type: string
              cluster_name:
                description: ''
                type: string
              name:
                description: ''
                type: string
              osd_fsid:
                description: ''
                type: string
              osd_id:
                description: ''
                type: string
              osdspec_affinity:
                description: ''
                type: string
              type:
                description: ''
                type: string
            required:
            - name
            - osd_id
            - cluster_name
            - type
            - osd_fsid
            - cluster_fsid
            - osdspec_affinity
            - block_uuid
            type: object
          type: array
        osd_ids:
          description: Device OSD IDs
          items:
            type: integer
          type: array
        path:
          description: Device path
          type: string
        rejected_reasons:
          description: ''
          items:
            type: string
          type: array
        sys_api:
          description: ''
          properties:
            human_readable_size:
              description: ''
              type: string
            locked:
              description: ''
              type: integer
            model:
              description: ''
              type: string
            nr_requests:
              description: ''
              type: string
            partitions:
              description: ''
              properties:
                partition_name:
                  description: ''
                  properties:
                    holders:
                      description: ''
                      items:
                        type: string
                      type: array
                    human_readable_size:
                      description: ''
                      type: string
                    sectors:
                      description: ''
                      type: string
                    sectorsize:
                      description: ''
                      type: integer
                    size:
                      description: ''
                      type: integer
                    start:
                      description: ''
                      type: string
                  required:
                  - start
                  - sectors
                  - sectorsize
                  - size
                  - human_readable_size
                  - holders
                  type: object
              required:
              - partition_name
              type: object
            path:
              description: ''
              type: string
            removable:
              description: ''
              type: string
            rev:
              description: ''
              type: string
            ro:
              description: ''
              type: string
            rotational:
              description: ''
              type: string
            sas_address:
              description: ''
              type: string
            sas_device_handle:
              description: ''
              type: string
            scheduler_mode:
              description: ''
              type: string
            sectors:
              description: ''
              type: integer
            sectorsize:
              description: ''
              type: string
            size:
              description: ''
              type: integer
            support_discard:
              description: ''
              type: string
            vendor:
              description: ''
              type: string
          required:
          - removable
          - ro
          - vendor
          - model
          - rev
          - sas_address
          - sas_device_handle
          - support_discard
          - rotational
          - nr_requests
          - scheduler_mode
          - partitions
          - sectors
          - sectorsize
          - size
          - human_readable_size
          - path
          - locked
          type: object
      required:
      - rejected_reasons
      - available
      - path
      - sys_api
      - lvs
      - human_readable_type
      - device_id
      - lsm_data
      - osd_ids
      type: object
    type: array
  labels:
    description: Host labels
    items:
      type: string
    type: array
  name:
    description: Hostname
    type: string
required:
- name
- addr
- devices
- labels
type: object
```

#### `400`

Operation exception. Please check the response body for details.

无响应体 schema。

#### `401`

Unauthenticated access. Please login first.

无响应体 schema。

#### `403`

Unauthorized access. Please check your permissions.

无响应体 schema。

#### `500`

Unexpected error. Please check the response body for the stack trace.

无响应体 schema。


### `GET /api/host/{hostname}/smart`

- 摘要：GET /api/host/{hostname}/smart
- Tags：`Host`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| hostname | path | 是 | string |  |  |


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

