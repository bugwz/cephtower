# Ceph 20.2.2 Dashboard API - CRUSH 规则

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/crush_rule` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`GET /api/crush_rule`](#get-api-crush-rule) - List Crush Rule Configuration
- [`POST /api/crush_rule`](#post-api-crush-rule) - POST /api/crush_rule
- [`GET /api/crush_rule/{name}`](#get-api-crush-rule-name) - GET /api/crush_rule/{name}
- [`DELETE /api/crush_rule/{name}`](#delete-api-crush-rule-name) - DELETE /api/crush_rule/{name}

## 接口详情

### `GET /api/crush_rule`

- 摘要：List Crush Rule Configuration
- Tags：`CrushRule`
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

- Content-Type: `application/vnd.ceph.api.v2.0+json`
- Schema: object; fields: max_size, min_size, rule_id, rule_name, ruleset, steps, type; required: rule_id, rule_name, ruleset, type, min_size, max_size, steps

```yaml
properties:
  max_size:
    description: Maximum size of Rule
    type: integer
  min_size:
    description: Minimum size of Rule
    type: integer
  rule_id:
    description: Rule ID
    type: integer
  rule_name:
    description: Rule Name
    type: string
  ruleset:
    description: RuleSet related to the rule
    type: integer
  steps:
    description: Steps included in the rule
    items:
      type: object
    type: array
  type:
    description: Type of Rule
    type: integer
required:
- rule_id
- rule_name
- ruleset
- type
- min_size
- max_size
- steps
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


### `POST /api/crush_rule`

- 摘要：POST /api/crush_rule
- Tags：`CrushRule`
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
- Schema: object; fields: device_class, failure_domain, name, pool_type, profile, root; required: name, failure_domain

```yaml
properties:
  device_class:
    type: string
  failure_domain:
    type: string
  name:
    type: string
  pool_type:
    default: replication
    type: string
  profile:
    type: string
  root:
    type: string
required:
- name
- failure_domain
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


### `GET /api/crush_rule/{name}`

- 摘要：GET /api/crush_rule/{name}
- Tags：`CrushRule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| name | path | 是 | string |  |  |


#### 请求体

无请求体。


#### 返回消息

#### `200`

OK

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


### `DELETE /api/crush_rule/{name}`

- 摘要：DELETE /api/crush_rule/{name}
- Tags：`CrushRule`
- 安全：- jwt: []

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| name | path | 是 | string |  |  |


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

