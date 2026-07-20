# Ceph 20.2.2 Dashboard API - 健康状态

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/health` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/health/full`](#get-api-health-full) - Get Cluster's detailed health report
- [`GET /api/health/get_cluster_capacity`](#get-api-health-get-cluster-capacity) - GET /api/health/get_cluster_capacity
- [`GET /api/health/get_cluster_fsid`](#get-api-health-get-cluster-fsid) - GET /api/health/get_cluster_fsid
- [`GET /api/health/get_telemetry_status`](#get-api-health-get-telemetry-status) - GET /api/health/get_telemetry_status
- [`GET /api/health/minimal`](#get-api-health-minimal) - Get Cluster's health report with lesser details
- [`GET /api/health/snapshot`](#get-api-health-snapshot) - Get a quick overview of cluster health at a moment, analogous to the ceph status command in CLI.

## 接口详情

### `GET /api/health/full`

- 摘要：Get Cluster's detailed health report
- Tags：`Health`
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


### `GET /api/health/get_cluster_capacity`

- 摘要：GET /api/health/get_cluster_capacity
- Tags：`Health`
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


### `GET /api/health/get_cluster_fsid`

- 摘要：GET /api/health/get_cluster_fsid
- Tags：`Health`
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


### `GET /api/health/get_telemetry_status`

- 摘要：GET /api/health/get_telemetry_status
- Tags：`Health`
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


### `GET /api/health/minimal`

- 摘要：Get Cluster's health report with lesser details
- Tags：`Health`
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
- Schema: object; fields: client_perf, df, fs_map, health, hosts, iscsi_daemons, mgr_map, mon_status, osd_map, pg_info, pools, rgw, scrub_status; required: client_perf, df, fs_map, health, hosts, iscsi_daemons, mgr_map, mon_status, osd_map, pg_info, pools, rgw, scrub_status

```yaml
properties:
  client_perf:
    description: ''
    properties:
      read_bytes_sec:
        description: ''
        type: integer
      read_op_per_sec:
        description: ''
        type: integer
      recovering_bytes_per_sec:
        description: ''
        type: integer
      write_bytes_sec:
        description: ''
        type: integer
      write_op_per_sec:
        description: ''
        type: integer
    required:
    - read_bytes_sec
    - read_op_per_sec
    - recovering_bytes_per_sec
    - write_bytes_sec
    - write_op_per_sec
    type: object
  df:
    description: ''
    properties:
      stats:
        description: ''
        properties:
          total_avail_bytes:
            description: ''
            type: integer
          total_bytes:
            description: ''
            type: integer
          total_used_raw_bytes:
            description: ''
            type: integer
        required:
        - total_avail_bytes
        - total_bytes
        - total_used_raw_bytes
        type: object
    required:
    - stats
    type: object
  fs_map:
    description: ''
    properties:
      filesystems:
        description: ''
        items:
          properties:
            mdsmap:
              description: ''
              properties:
                balancer:
                  description: ''
                  type: string
                btime:
                  description: ''
                  type: string
                compat:
                  description: ''
                  properties:
                    compat:
                      description: ''
                      type: string
                    incompat:
                      description: ''
                      type: string
                    ro_compat:
                      description: ''
                      type: string
                  required:
                  - compat
                  - ro_compat
                  - incompat
                  type: object
                created:
                  description: ''
                  type: string
                damaged:
                  description: ''
                  items:
                    type: integer
                  type: array
                data_pools:
                  description: ''
                  items:
                    type: integer
                  type: array
                enabled:
                  description: ''
                  type: boolean
                epoch:
                  description: ''
                  type: integer
                ever_allowed_features:
                  description: ''
                  type: integer
                explicitly_allowed_features:
                  description: ''
                  type: integer
                failed:
                  description: ''
                  items:
                    type: integer
                  type: array
                flags:
                  description: ''
                  type: integer
                fs_name:
                  description: ''
                  type: string
                in:
                  description: ''
                  items:
                    type: integer
                  type: array
                info:
                  description: ''
                  type: string
                last_failure:
                  description: ''
                  type: integer
                last_failure_osd_epoch:
                  description: ''
                  type: integer
                max_file_size:
                  description: ''
                  type: integer
                max_mds:
                  description: ''
                  type: integer
                metadata_pool:
                  description: ''
                  type: integer
                modified:
                  description: ''
                  type: string
                required_client_features:
                  description: ''
                  type: string
                root:
                  description: ''
                  type: integer
                session_autoclose:
                  description: ''
                  type: integer
                session_timeout:
                  description: ''
                  type: integer
                standby_count_wanted:
                  description: ''
                  type: integer
                stopped:
                  description: ''
                  items:
                    type: integer
                  type: array
                tableserver:
                  description: ''
                  type: integer
                up:
                  description: ''
                  type: string
              required:
              - session_autoclose
              - balancer
              - up
              - last_failure_osd_epoch
              - in
              - last_failure
              - max_file_size
              - explicitly_allowed_features
              - damaged
              - tableserver
              - failed
              - metadata_pool
              - epoch
              - btime
              - stopped
              - max_mds
              - compat
              - required_client_features
              - data_pools
              - info
              - fs_name
              - created
              - standby_count_wanted
              - enabled
              - modified
              - session_timeout
              - flags
              - ever_allowed_features
              - root
              type: object
            standbys:
              description: ''
              type: string
          required:
          - mdsmap
          - standbys
          type: object
        type: array
    required:
    - filesystems
    type: object
  health:
    description: ''
    properties:
      checks:
        description: ''
        type: string
      mutes:
        description: ''
        type: string
      status:
        description: ''
        type: string
    required:
    - checks
    - mutes
    - status
    type: object
  hosts:
    description: ''
    type: integer
  iscsi_daemons:
    description: ''
    properties:
      down:
        description: ''
        type: integer
      up:
        description: ''
        type: integer
    required:
    - up
    - down
    type: object
  mgr_map:
    description: ''
    properties:
      active_name:
        description: ''
        type: string
      standbys:
        description: ''
        type: string
    required:
    - active_name
    - standbys
    type: object
  mon_status:
    description: ''
    properties:
      monmap:
        description: ''
        properties:
          mons:
            description: ''
            type: string
        required:
        - mons
        type: object
      quorum:
        description: ''
        items:
          type: integer
        type: array
    required:
    - monmap
    - quorum
    type: object
  osd_map:
    description: ''
    properties:
      osds:
        description: ''
        items:
          properties:
            in:
              description: ''
              type: integer
            up:
              description: ''
              type: integer
          required:
          - in
          - up
          type: object
        type: array
    required:
    - osds
    type: object
  pg_info:
    description: ''
    properties:
      object_stats:
        description: ''
        properties:
          num_object_copies:
            description: ''
            type: integer
          num_objects:
            description: ''
            type: integer
          num_objects_degraded:
            description: ''
            type: integer
          num_objects_misplaced:
            description: ''
            type: integer
          num_objects_unfound:
            description: ''
            type: integer
        required:
        - num_objects
        - num_object_copies
        - num_objects_degraded
        - num_objects_misplaced
        - num_objects_unfound
        type: object
      pgs_per_osd:
        description: ''
        type: integer
      statuses:
        description: ''
        type: string
    required:
    - object_stats
    - pgs_per_osd
    - statuses
    type: object
  pools:
    description: ''
    type: string
  rgw:
    description: ''
    type: integer
  scrub_status:
    description: ''
    type: string
required:
- client_perf
- df
- fs_map
- health
- hosts
- iscsi_daemons
- mgr_map
- mon_status
- osd_map
- pg_info
- pools
- rgw
- scrub_status
type: object
```

#### `400`

Operation exception. Please check the response body for details.

无响应体 schema。

#### `401`

Unauthenticated access. Please login first.

无响应体 schema。

#### `403`

Unauthorized access. Please check your permissions.

无响应体 schema。

#### `500`

Unexpected error. Please check the response body for the stack trace.

无响应体 schema。


### `GET /api/health/snapshot`

- 摘要：Get a quick overview of cluster health at a moment, analogous to the ceph status command in CLI.
- Tags：`Health`
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
- Schema: object; fields: fsid, fsmap, health, mgrmap, monmap, num_hosts, num_hosts_available, num_iscsi_gateways, num_rgw_gateways, osdmap, pgmap; required: fsid, health, monmap, osdmap, pgmap, mgrmap, fsmap, num_rgw_gateways, num_iscsi_gateways, num_hosts, num_hosts_available

```yaml
properties:
  fsid:
    description: Cluster filesystem ID
    type: string
  fsmap:
    description: Filesystem map details
    properties:
      num_active:
        description: Number of active mds
        type: integer
      num_standbys:
        description: Standby MDS count
        type: integer
    required:
    - num_active
    - num_standbys
    type: object
  health:
    description: Cluster health overview
    properties:
      checks:
        description: Health checks keyed by name
        properties:
          "<check_name>":
            description: Individual health check object
            properties:
              muted:
                description: Whether the check is muted
                type: boolean
              severity:
                description: Health severity level
                type: string
              summary:
                description: Summary details
                properties:
                  count:
                    description: Occurrence count
                    type: integer
                  message:
                    description: Human-readable summary
                    type: string
                required:
                - message
                - count
                type: object
            required:
            - severity
            - summary
            - muted
            type: object
        required:
        - "<check_name>"
        type: object
      mutes:
        description: List of muted check names
        items:
          type: string
        type: array
      status:
        description: Overall health status
        type: string
    required:
    - status
    - checks
    - mutes
    type: object
  mgrmap:
    description: Manager map details
    properties:
      num_active:
        description: Number of active managers
        type: integer
      num_standbys:
        description: Standby manager count
        type: integer
    required:
    - num_active
    - num_standbys
    type: object
  monmap:
    description: Monitor map details
    properties:
      num_mons:
        description: Number of monitors
        type: integer
      quorum:
        description: List of monitors in quorum
        items:
          type: integer
        type: array
    required:
    - num_mons
    - quorum
    type: object
  num_hosts:
    description: Count of hosts
    type: integer
  num_hosts_available:
    description: Count of available hosts
    type: integer
  num_iscsi_gateways:
    description: Iscsi gateways status
    properties:
      down:
        description: Count of iSCSI gateways not running
        type: integer
      up:
        description: Count of iSCSI gateways running
        type: integer
    required:
    - up
    - down
    type: object
  num_rgw_gateways:
    description: Count of RGW gateway daemons running
    type: integer
  osdmap:
    description: OSD map details
    properties:
      in:
        description: Number of OSDs in
        type: integer
      num_osds:
        description: Total OSD count
        type: integer
      up:
        description: Number of OSDs up
        type: integer
    required:
    - in
    - up
    - num_osds
    type: object
  pgmap:
    description: Placement group map details
    properties:
      bytes_total:
        description: Total capacity in bytes
        type: integer
      bytes_used:
        description: Used capacity in bytes
        type: integer
      num_pgs:
        description: Total PG count
        type: integer
      num_pools:
        description: Number of pools
        type: integer
      pgs_by_state:
        description: List of PG counts by state
        items:
          properties:
            count:
              description: Count of PGs in this state
              type: integer
            state_name:
              description: Placement group state
              type: string
          required:
          - state_name
          - count
          type: object
        type: array
      recovering_bytes_per_sec:
        description: Total recovery in bytes
        type: integer
    required:
    - pgs_by_state
    - num_pools
    - num_pgs
    - bytes_used
    - bytes_total
    - recovering_bytes_per_sec
    type: object
required:
- fsid
- health
- monmap
- osdmap
- pgmap
- mgrmap
- fsmap
- num_rgw_gateways
- num_iscsi_gateways
- num_hosts
- num_hosts_available
type: object
```

#### `400`

Operation exception. Please check the response body for details.

无响应体 schema。

#### `401`

Unauthenticated access. Please login first.

无响应体 schema。

#### `403`

Unauthorized access. Please check your permissions.

无响应体 schema。

#### `500`

Unexpected error. Please check the response body for the stack trace.

无响应体 schema。

