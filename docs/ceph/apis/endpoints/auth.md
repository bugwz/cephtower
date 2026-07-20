# Ceph 20.2.2 Dashboard API - 认证

> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。
> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/auth` 路径域归类。
> 版本支持扫描范围：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2。

## 接口目录

- [`POST /api/auth`](#post-api-auth) - Dashboard Authentication
- [`POST /api/auth/check`](#post-api-auth-check) - Check token Authentication
- [`POST /api/auth/logout`](#post-api-auth-logout) - POST /api/auth/logout

## 接口详情

### `POST /api/auth`

- 摘要：Dashboard Authentication
- Tags：`Auth`
- 安全：OpenAPI 未声明 JWT security，通常为公开或由控制器单独处理

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: password, ttl, username; required: username, password

```yaml
properties:
  password:
    description: Password
    type: string
  ttl:
    description: Token Time to Live (in hours)
    type: integer
  username:
    description: Username
    type: string
required:
- username
- password
type: object
```


#### 返回消息

#### `201`

Resource created.

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: permissions, pwdExpirationDate, pwdUpdateRequired, sso, token, username; required: token, username, permissions, pwdExpirationDate, sso, pwdUpdateRequired

```yaml
properties:
  permissions:
    description: List of permissions acquired
    properties:
      cephfs:
        description: ''
        items:
          type: string
        type: array
    required:
    - cephfs
    type: object
  pwdExpirationDate:
    description: Password expiration date
    type: string
  pwdUpdateRequired:
    description: Is password update required?
    type: boolean
  sso:
    description: Uses single sign on?
    type: boolean
  token:
    description: Authentication Token
    type: string
  username:
    description: Username
    type: string
required:
- token
- username
- permissions
- pwdExpirationDate
- sso
- pwdUpdateRequired
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


### `POST /api/auth/check`

- 摘要：Check token Authentication
- Tags：`Auth`
- 安全：OpenAPI 未声明 JWT security，通常为公开或由控制器单独处理

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

| 名称 | 位置 | 必填 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| token | query | 是 | string |  | Authentication Token |


#### 请求体

请求体必填：否

- Content-Type: `application/json`
- Schema: object; fields: token; required: token

```yaml
properties:
  token:
    description: Authentication Token
    type: string
required:
- token
type: object
```


#### 返回消息

#### `201`

Resource created.

- Content-Type: `application/vnd.ceph.api.v1.0+json`
- Schema: object; fields: permissions, pwdUpdateRequired, sso, username; required: username, permissions, sso, pwdUpdateRequired

```yaml
properties:
  permissions:
    description: List of permissions acquired
    properties:
      cephfs:
        description: ''
        items:
          type: string
        type: array
    required:
    - cephfs
    type: object
  pwdUpdateRequired:
    description: Is password update required?
    type: boolean
  sso:
    description: Uses single sign on?
    type: boolean
  username:
    description: Username
    type: string
required:
- username
- permissions
- sso
- pwdUpdateRequired
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


### `POST /api/auth/logout`

- 摘要：POST /api/auth/logout
- Tags：`Auth`
- 安全：OpenAPI 未声明 JWT security，通常为公开或由控制器单独处理

#### 版本支持

- 支持版本：v16.2.15, v17.2.9, v18.2.8, v19.2.5, v20.2.2
- 首次出现在扫描范围：v16.2.15
- v20.2.2 当前文档支持：是

#### 请求参数

无。


#### 请求体

无请求体。


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

