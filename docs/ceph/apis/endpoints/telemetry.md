# Ceph 20.2.2 Dashboard API - Telemetry

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/telemetry` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`PUT /api/telemetry`](#put-api-telemetry) - PUT /api/telemetry
- [`GET /api/telemetry/report`](#get-api-telemetry-report) - Get Detailed Telemetry report

## 接口详情

### `PUT /api/telemetry`

- 摘要：PUT /api/telemetry
- Tags：`Telemetry`
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
- Schema: object; fields: enable, license_name; required: 无

```yaml
properties:
  enable:
    default: true
    type: boolean
  license_name:
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


### `GET /api/telemetry/report`

- 摘要：Get Detailed Telemetry report
- Tags：`Telemetry`
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
- Schema: object; fields: device_report, report; required: report, device_report

```yaml
properties:
  device_report:
    description: ''
    type: string
  report:
    description: ''
    properties:
      balancer:
        description: ''
        properties:
          active:
            description: ''
            type: boolean
          mode:
            description: ''
            type: string
        required:
        - active
        - mode
        type: object
      channels:
        description: ''
        items:
          type: string
        type: array
      channels_available:
        description: ''
        items:
          type: string
        type: array
      config:
        description: ''
        properties:
          active_changed:
            description: ''
            items:
              type: string
            type: array
          cluster_changed:
            description: ''
            items:
              type: string
            type: array
        required:
        - cluster_changed
        - active_changed
        type: object
      crashes:
        description: ''
        items:
          type: integer
        type: array
      created:
        description: ''
        type: string
      crush:
        description: ''
        properties:
          bucket_algs:
            description: ''
            properties:
              straw2:
                description: ''
                type: integer
            required:
            - straw2
            type: object
          bucket_sizes:
            description: ''
            properties:
              '1':
                description: ''
                type: integer
              '3':
                description: ''
                type: integer
            required:
            - '1'
            - '3'
            type: object
          bucket_types:
            description: ''
            properties:
              '1':
                description: ''
                type: integer
              '11':
                description: ''
                type: integer
            required:
            - '1'
            - '11'
            type: object
          compat_weight_set:
            description: ''
            type: boolean
          device_classes:
            description: ''
            items:
              type: integer
            type: array
          num_buckets:
            description: ''
            type: integer
          num_devices:
            description: ''
            type: integer
          num_rules:
            description: ''
            type: integer
          num_types:
            description: ''
            type: integer
          num_weight_sets:
            description: ''
            type: integer
          tunables:
            description: ''
            properties:
              allowed_bucket_algs:
                description: ''
                type: integer
              choose_local_fallback_tries:
                description: ''
                type: integer
              choose_local_tries:
                description: ''
                type: integer
              choose_total_tries:
                description: ''
                type: integer
              chooseleaf_descend_once:
                description: ''
                type: integer
              chooseleaf_stable:
                description: ''
                type: integer
              chooseleaf_vary_r:
                description: ''
                type: integer
              has_v2_rules:
                description: ''
                type: integer
              has_v3_rules:
                description: ''
                type: integer
              has_v4_buckets:
                description: ''
                type: integer
              has_v5_rules:
                description: ''
                type: integer
              legacy_tunables:
                description: ''
                type: integer
              minimum_required_version:
                description: ''
                type: string
              optimal_tunables:
                description: ''
                type: integer
              profile:
                description: ''
                type: string
              require_feature_tunables:
                description: ''
                type: integer
              require_feature_tunables2:
                description: ''
                type: integer
              require_feature_tunables3:
                description: ''
                type: integer
              require_feature_tunables5:
                description: ''
                type: integer
              straw_calc_version:
                description: ''
                type: integer
            required:
            - choose_local_tries
            - choose_local_fallback_tries
            - choose_total_tries
            - chooseleaf_descend_once
            - chooseleaf_vary_r
            - chooseleaf_stable
            - straw_calc_version
            - allowed_bucket_algs
            - profile
            - optimal_tunables
            - legacy_tunables
            - minimum_required_version
            - require_feature_tunables
            - require_feature_tunables2
            - has_v2_rules
            - require_feature_tunables3
            - has_v3_rules
            - has_v4_buckets
            - require_feature_tunables5
            - has_v5_rules
            type: object
        required:
        - num_devices
        - num_types
        - num_buckets
        - num_rules
        - device_classes
        - tunables
        - compat_weight_set
        - num_weight_sets
        - bucket_algs
        - bucket_sizes
        - bucket_types
        type: object
      fs:
        description: ''
        properties:
          count:
            description: ''
            type: integer
          feature_flags:
            description: ''
            properties:
              enable_multiple:
                description: ''
                type: boolean
              ever_enabled_multiple:
                description: ''
                type: boolean
            required:
            - enable_multiple
            - ever_enabled_multiple
            type: object
          filesystems:
            description: ''
            items:
              type: integer
            type: array
          num_standby_mds:
            description: ''
            type: integer
          total_num_mds:
            description: ''
            type: integer
        required:
        - count
        - feature_flags
        - num_standby_mds
        - filesystems
        - total_num_mds
        type: object
      hosts:
        description: ''
        properties:
          num:
            description: ''
            type: integer
          num_with_mds:
            description: ''
            type: integer
          num_with_mgr:
            description: ''
            type: integer
          num_with_mon:
            description: ''
            type: integer
          num_with_osd:
            description: ''
            type: integer
        required:
        - num
        - num_with_mon
        - num_with_mds
        - num_with_osd
        - num_with_mgr
        type: object
      leaderboard:
        description: ''
        type: boolean
      license:
        description: ''
        type: string
      metadata:
        description: ''
        properties:
          mon:
            description: ''
            properties:
              arch:
                description: ''
                properties:
                  x86_64:
                    description: ''
                    type: integer
                required:
                - x86_64
                type: object
              ceph_version:
                description: ''
                properties:
                  ceph version 16.0.0-3151-gf202994fcf:
                    description: ''
                    type: integer
                required:
                - ceph version 16.0.0-3151-gf202994fcf
                type: object
              cpu:
                description: ''
                properties:
                  Intel(R) Core(TM) i7-8665U CPU @ 1.90GHz:
                    description: ''
                    type: integer
                required:
                - Intel(R) Core(TM) i7-8665U CPU @ 1.90GHz
                type: object
              distro:
                description: ''
                properties:
                  centos:
                    description: ''
                    type: integer
                required:
                - centos
                type: object
              distro_description:
                description: ''
                properties:
                  CentOS Linux 8 (Core):
                    description: ''
                    type: integer
                required:
                - CentOS Linux 8 (Core)
                type: object
              kernel_description:
                description: ''
                properties:
                  "#1 SMP Wed Jul 1 19:53:01 UTC 2020":
                    description: ''
                    type: integer
                required:
                - "#1 SMP Wed Jul 1 19:53:01 UTC 2020"
                type: object
              kernel_version:
                description: ''
                properties:
                  5.7.7-200.fc32.x86_64:
                    description: ''
                    type: integer
                required:
                - 5.7.7-200.fc32.x86_64
                type: object
              os:
                description: ''
                properties:
                  Linux:
                    description: ''
                    type: integer
                required:
                - Linux
                type: object
            required:
            - arch
            - ceph_version
            - os
            - cpu
            - kernel_description
            - kernel_version
            - distro_description
            - distro
            type: object
          osd:
            description: ''
            properties:
              arch:
                description: ''
                properties:
                  x86_64:
                    description: ''
                    type: integer
                required:
                - x86_64
                type: object
              ceph_version:
                description: ''
                properties:
                  ceph version 16.0.0-3151-gf202994fcf:
                    description: ''
                    type: integer
                required:
                - ceph version 16.0.0-3151-gf202994fcf
                type: object
              cpu:
                description: ''
                properties:
                  Intel(R) Core(TM) i7-8665U CPU @ 1.90GHz:
                    description: ''
                    type: integer
                required:
                - Intel(R) Core(TM) i7-8665U CPU @ 1.90GHz
                type: object
              distro:
                description: ''
                properties:
                  centos:
                    description: ''
                    type: integer
                required:
                - centos
                type: object
              distro_description:
                description: ''
                properties:
                  CentOS Linux 8 (Core):
                    description: ''
                    type: integer
                required:
                - CentOS Linux 8 (Core)
                type: object
              kernel_description:
                description: ''
                properties:
                  "#1 SMP Wed Jul 1 19:53:01 UTC 2020":
                    description: ''
                    type: integer
                required:
                - "#1 SMP Wed Jul 1 19:53:01 UTC 2020"
                type: object
              kernel_version:
                description: ''
                properties:
                  5.7.7-200.fc32.x86_64:
                    description: ''
                    type: integer
                required:
                - 5.7.7-200.fc32.x86_64
                type: object
              os:
                description: ''
                properties:
                  Linux:
                    description: ''
                    type: integer
                required:
                - Linux
                type: object
              osd_objectstore:
                description: ''
                properties:
                  bluestore:
                    description: ''
                    type: integer
                required:
                - bluestore
                type: object
              rotational:
                description: ''
                properties:
                  '1':
                    description: ''
                    type: integer
                required:
                - '1'
                type: object
            required:
            - osd_objectstore
            - rotational
            - arch
            - ceph_version
            - os
            - cpu
            - kernel_description
            - kernel_version
            - distro_description
            - distro
            type: object
        required:
        - osd
        - mon
        type: object
      mon:
        description: ''
        properties:
          count:
            description: ''
            type: integer
          features:
            description: ''
            properties:
              optional:
                description: ''
                items:
                  type: integer
                type: array
              persistent:
                description: ''
                items:
                  type: string
                type: array
            required:
            - persistent
            - optional
            type: object
          ipv4_addr_mons:
            description: ''
            type: integer
          ipv6_addr_mons:
            description: ''
            type: integer
          min_mon_release:
            description: ''
            type: integer
          v1_addr_mons:
            description: ''
            type: integer
          v2_addr_mons:
            description: ''
            type: integer
        required:
        - count
        - features
        - min_mon_release
        - v1_addr_mons
        - v2_addr_mons
        - ipv4_addr_mons
        - ipv6_addr_mons
        type: object
      osd:
        description: ''
        properties:
          cluster_network:
            description: ''
            type: boolean
          count:
            description: ''
            type: integer
          require_min_compat_client:
            description: ''
            type: string
          require_osd_release:
            description: ''
            type: string
        required:
        - count
        - require_osd_release
        - require_min_compat_client
        - cluster_network
        type: object
      pools:
        description: ''
        items:
          properties:
            cache_mode:
              description: ''
              type: string
            erasure_code_profile:
              description: ''
              type: string
            min_size:
              description: ''
              type: integer
            pg_autoscale_mode:
              description: ''
              type: string
            pg_num:
              description: ''
              type: integer
            pgp_num:
              description: ''
              type: integer
            pool:
              description: ''
              type: integer
            size:
              description: ''
              type: integer
            target_max_bytes:
              description: ''
              type: integer
            target_max_objects:
              description: ''
              type: integer
            type:
              description: ''
              type: string
          required:
          - pool
          - type
          - pg_num
          - pgp_num
          - size
          - min_size
          - pg_autoscale_mode
          - target_max_bytes
          - target_max_objects
          - erasure_code_profile
          - cache_mode
          type: object
        type: array
      rbd:
        description: ''
        properties:
          mirroring_by_pool:
            description: ''
            items:
              type: boolean
            type: array
          num_images_by_pool:
            description: ''
            items:
              type: integer
            type: array
          num_pools:
            description: ''
            type: integer
        required:
        - num_pools
        - num_images_by_pool
        - mirroring_by_pool
        type: object
      report_id:
        description: ''
        type: string
      report_timestamp:
        description: ''
        type: string
      report_version:
        description: ''
        type: integer
      rgw:
        description: ''
        properties:
          count:
            description: ''
            type: integer
          frontends:
            description: ''
            items:
              type: string
            type: array
          zonegroups:
            description: ''
            type: integer
          zones:
            description: ''
            type: integer
        required:
        - count
        - zones
        - zonegroups
        - frontends
        type: object
      services:
        description: ''
        properties:
          rgw:
            description: ''
            type: integer
        required:
        - rgw
        type: object
      usage:
        description: ''
        properties:
          pg_num:
            description: ''
            type: integer
          pools:
            description: ''
            type: integer
          total_avail_bytes:
            description: ''
            type: integer
          total_bytes:
            description: ''
            type: integer
          total_used_bytes:
            description: ''
            type: integer
        required:
        - pools
        - pg_num
        - total_used_bytes
        - total_bytes
        - total_avail_bytes
        type: object
    required:
    - leaderboard
    - report_version
    - report_timestamp
    - report_id
    - channels
    - channels_available
    - license
    - created
    - mon
    - config
    - rbd
    - pools
    - osd
    - crush
    - fs
    - metadata
    - hosts
    - usage
    - services
    - rgw
    - balancer
    - crashes
    type: object
required:
- report
- device_report
type: object
```

#### `400`

Operation exception. Please check the response body for details.

无响应体 schema。

#### `401`

Unauthenticated access. Please login first.

无响应体 schema。

#### `403`

Unauthorized access. Please check your permissions.

无响应体 schema。

#### `500`

Unexpected error. Please check the response body for the stack trace.

无响应体 schema。

