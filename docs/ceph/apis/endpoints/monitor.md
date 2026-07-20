# Ceph 20.2.2 Dashboard API - Monitor

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/monitor` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/monitor`](#get-api-monitor) - Get Monitor Details

## 接口详情

### `GET /api/monitor`

- 摘要：Get Monitor Details
- Tags：`Monitor`
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
- Schema: object; fields: in_quorum, mon_status, out_quorum; required: mon_status, in_quorum, out_quorum

```yaml
properties:
  in_quorum:
    description: ''
    items:
      properties:
        addr:
          description: ''
          type: string
        name:
          description: ''
          type: string
        priority:
          description: ''
          type: integer
        public_addr:
          description: ''
          type: string
        public_addrs:
          description: ''
          properties:
            addrvec:
              description: ''
              items:
                properties:
                  addr:
                    description: ''
                    type: string
                  nonce:
                    description: ''
                    type: integer
                  type:
                    description: ''
                    type: string
                required:
                - type
                - addr
                - nonce
                type: object
              type: array
          required:
          - addrvec
          type: object
        rank:
          description: ''
          type: integer
        stats:
          description: ''
          properties:
            num_sessions:
              description: ''
              items:
                type: integer
              type: array
          required:
          - num_sessions
          type: object
        weight:
          description: ''
          type: integer
      required:
      - rank
      - name
      - public_addrs
      - addr
      - public_addr
      - priority
      - weight
      - stats
      type: object
    type: array
  mon_status:
    description: ''
    properties:
      election_epoch:
        description: ''
        type: integer
      extra_probe_peers:
        description: ''
        items:
          type: string
        type: array
      feature_map:
        description: ''
        properties:
          client:
            description: ''
            items:
              properties:
                features:
                  description: ''
                  type: string
                num:
                  description: ''
                  type: integer
                release:
                  description: ''
                  type: string
              required:
              - features
              - release
              - num
              type: object
            type: array
          mds:
            description: ''
            items:
              properties:
                features:
                  description: ''
                  type: string
                num:
                  description: ''
                  type: integer
                release:
                  description: ''
                  type: string
              required:
              - features
              - release
              - num
              type: object
            type: array
          mgr:
            description: ''
            items:
              properties:
                features:
                  description: ''
                  type: string
                num:
                  description: ''
                  type: integer
                release:
                  description: ''
                  type: string
              required:
              - features
              - release
              - num
              type: object
            type: array
          mon:
            description: ''
            items:
              properties:
                features:
                  description: ''
                  type: string
                num:
                  description: ''
                  type: integer
                release:
                  description: ''
                  type: string
              required:
              - features
              - release
              - num
              type: object
            type: array
        required:
        - mon
        - mds
        - client
        - mgr
        type: object
      features:
        description: ''
        properties:
          quorum_con:
            description: ''
            type: string
          quorum_mon:
            description: ''
            items:
              type: string
            type: array
          required_con:
            description: ''
            type: string
          required_mon:
            description: ''
            items:
              type: integer
            type: array
        required:
        - required_con
        - required_mon
        - quorum_con
        - quorum_mon
        type: object
      monmap:
        description: ''
        properties:
          created:
            description: ''
            type: string
          epoch:
            description: ''
            type: integer
          features:
            description: ''
            properties:
              optional:
                description: ''
                items:
                  type: string
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
          fsid:
            description: ''
            type: string
          min_mon_release:
            description: ''
            type: integer
          min_mon_release_name:
            description: ''
            type: string
          modified:
            description: ''
            type: string
          mons:
            description: ''
            items:
              properties:
                addr:
                  description: ''
                  type: string
                name:
                  description: ''
                  type: string
                priority:
                  description: ''
                  type: integer
                public_addr:
                  description: ''
                  type: string
                public_addrs:
                  description: ''
                  properties:
                    addrvec:
                      description: ''
                      items:
                        properties:
                          addr:
                            description: ''
                            type: string
                          nonce:
                            description: ''
                            type: integer
                          type:
                            description: ''
                            type: string
                        required:
                        - type
                        - addr
                        - nonce
                        type: object
                      type: array
                  required:
                  - addrvec
                  type: object
                rank:
                  description: ''
                  type: integer
                stats:
                  description: ''
                  properties:
                    num_sessions:
                      description: ''
                      items:
                        type: integer
                      type: array
                  required:
                  - num_sessions
                  type: object
                weight:
                  description: ''
                  type: integer
              required:
              - rank
              - name
              - public_addrs
              - addr
              - public_addr
              - priority
              - weight
              - stats
              type: object
            type: array
        required:
        - epoch
        - fsid
        - modified
        - created
        - min_mon_release
        - min_mon_release_name
        - features
        - mons
        type: object
      name:
        description: ''
        type: string
      outside_quorum:
        description: ''
        items:
          type: string
        type: array
      quorum:
        description: ''
        items:
          type: integer
        type: array
      quorum_age:
        description: ''
        type: integer
      rank:
        description: ''
        type: integer
      state:
        description: ''
        type: string
      sync_provider:
        description: ''
        items:
          type: string
        type: array
    required:
    - name
    - rank
    - state
    - election_epoch
    - quorum
    - quorum_age
    - features
    - outside_quorum
    - extra_probe_peers
    - sync_provider
    - monmap
    - feature_map
    type: object
  out_quorum:
    description: ''
    items:
      type: integer
    type: array
required:
- mon_status
- in_quorum
- out_quorum
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

