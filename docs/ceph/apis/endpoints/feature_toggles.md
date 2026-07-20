# Ceph 20.2.2 Dashboard API - 功能开关

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/feature_toggles` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/feature_toggles`](#get-api-feature-toggles) - Get List Of Features

## 接口详情

### `GET /api/feature_toggles`

- 摘要：Get List Of Features
- Tags：`FeatureTogglesEndpoint`
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
- Schema: object; fields: cephfs, dashboard, iscsi, mirroring, nfs, rbd, rgw; required: rbd, mirroring, iscsi, cephfs, rgw, nfs, dashboard

```yaml
properties:
  cephfs:
    description: ''
    type: boolean
  dashboard:
    description: ''
    type: boolean
  iscsi:
    description: ''
    type: boolean
  mirroring:
    description: ''
    type: boolean
  nfs:
    description: ''
    type: boolean
  rbd:
    description: ''
    type: boolean
  rgw:
    description: ''
    type: boolean
required:
- rbd
- mirroring
- iscsi
- cephfs
- rgw
- nfs
- dashboard
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

