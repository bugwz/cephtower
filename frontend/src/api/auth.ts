import {
  getAuthToken,
  jsonInit,
  notifyApiError,
  readApiResponse,
  request,
  setAuthToken,
  toApiErrorDetail,
  type ApiRequestInit
} from './client'

export type UserRole = 'cluster-admin' | 'security-admin' | 'storage-admin' | 'operator' | 'viewer'

export interface UserAccount {
  id: number
  username: string
  display_name: string
  email?: string | null
  role: UserRole
  permissions: string[]
  enabled: boolean
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface SetupDatabaseConfig {
  engine: 'sqlite' | 'mysql'
  sqlite: {
    name: string
  }
  mysql: {
    host: string
    port: number
    username: string
    password: string
    password_set: boolean
    database: string
    params: string
    tls: 'false' | 'true' | 'skip-verify' | 'preferred'
  }
}

export interface SetupStatus {
  initialized: boolean
  database?: SetupDatabaseConfig
}

export interface SetupDatabaseInput {
  engine: 'sqlite' | 'mysql'
  sqlite: {
    name: string
  }
  mysql: {
    host: string
    port: number
    username: string
    password: string
    database: string
    params: string
    tls: 'false' | 'true' | 'skip-verify' | 'preferred'
  }
}

interface LoginResponse {
  token: string
  expires_at: string
  user: UserAccount
}

const authBaseUrl = '/api/v1'
const storedUserKey = 'cephtower.auth.user'

export async function login(username: string, password: string): Promise<LoginResponse> {
  return requestPublic<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
    suppressErrorNotification: true
  }).then((payload) => {
    payload.user = normalizeUser(payload.user)
    setAuthToken(payload.token)
    storeCurrentUser(payload.user)
    return payload
  })
}

export async function currentUser(): Promise<UserAccount> {
  const stored = readStoredUser()
  if (stored) {
    return stored
  }
  const users = await listUsers()
  if (users.length > 0) {
    storeCurrentUser(users[0])
    return users[0]
  }
  throw new Error('当前登录用户不存在')
}

export async function setupStatus(): Promise<SetupStatus> {
  const payload = await requestPublic<{ required: boolean }>('/bootstrap', { suppressErrorNotification: true })
  return { initialized: !payload.required, database: defaultSetupDatabase() }
}

export async function testSetupDatabase(payload: SetupDatabaseInput): Promise<{ message: string }> {
  await requestPublic<{ status?: string }>('/bootstrap/dbtest', {
    ...jsonInit('POST', payload),
    suppressErrorNotification: true
  })
  return { message: '后端数据库连接正常' }
}

export async function initializeSetup(payload: {
  database: SetupDatabaseInput
  user: {
    username: string
    email: string
    password: string
  }
}): Promise<{ message: string }> {
  await requestPublic<UserAccount>('/bootstrap/run', jsonInit('POST', {
    database: payload.database,
    user: {
      username: payload.user.username,
      display_name: payload.user.username,
      email: payload.user.email,
      password: payload.user.password,
      role: 'cluster-admin'
    }
  }, { suppressErrorNotification: true }))
  return { message: '初始化完成' }
}

export async function listUsers(): Promise<UserAccount[]> {
  const payload = await request<unknown>('/user', { method: 'GET' })
  if (Array.isArray(payload)) {
    return payload.map(normalizeUser)
  }
  if (isListPayload(payload)) {
    return payload.items.map(normalizeUser)
  }
  return []
}

export async function createUser(payload: {
  username: string
  display_name: string
  email?: string
  role: UserRole
  permissions?: string[]
  password: string
  enabled: boolean
}): Promise<UserAccount> {
  const created = await requestAuth<UserAccount>('/user', jsonInit('POST', {
    username: payload.username,
    display_name: payload.display_name,
    email: payload.email,
    password: payload.password,
    role: payload.role
  }))
  return normalizeUser(created)
}

export async function requestPasswordReset(account: string): Promise<{ message: string }> {
  return requestPublic<{ message: string }>('/auth/password-reset/request', {
    method: 'POST',
    body: JSON.stringify({ account }),
    suppressErrorNotification: true
  })
}

export async function confirmPasswordReset(payload: {
  account: string
  code: string
  new_password: string
}): Promise<{ message: string }> {
  return requestPublic<{ message: string }>('/auth/password-reset/confirm', {
    method: 'POST',
    body: JSON.stringify(payload),
    suppressErrorNotification: true
  })
}

export function logout() {
  setAuthToken('')
  try {
    localStorage.removeItem(storedUserKey)
  } catch {
    // Ignore storage failures during logout.
  }
}

export function hasStoredToken() {
  return Boolean(getAuthToken())
}

async function requestAuth<T>(path: string, init?: ApiRequestInit): Promise<T> {
  const { suppressErrorNotification, ...fetchInit } = init ?? {}
  try {
    const response = await fetch(`${authBaseUrl}${path}`, {
      ...fetchInit,
      headers: {
        'Content-Type': 'application/json',
        ...(getAuthToken() ? { Authorization: `Bearer ${getAuthToken()}` } : {}),
        ...fetchInit.headers
      }
    })
    return await readApiResponse<T>(response)
  } catch (err) {
    if (!suppressErrorNotification) {
      notifyApiError(toApiErrorDetail(err, path))
    }
    throw err
  }
}

async function requestPublic<T>(path: string, init?: ApiRequestInit): Promise<T> {
  const { suppressErrorNotification, ...fetchInit } = init ?? {}
  try {
    const response = await fetch(`${authBaseUrl}${path}`, {
      ...fetchInit,
      headers: {
        'Content-Type': 'application/json',
        ...fetchInit.headers
      }
    })
    return await readApiResponse<T>(response)
  } catch (err) {
    if (!suppressErrorNotification) {
      notifyApiError(toApiErrorDetail(err, path))
    }
    throw err
  }
}

function normalizeUser(user: Partial<UserAccount>): UserAccount {
  const role = normalizeRole(user.role)
  const status = (user as Partial<UserAccount> & { status?: string }).status
  return {
    id: Number(user.id ?? 0),
    username: user.username ?? '',
    display_name: user.display_name || user.username || '',
    email: user.email ?? null,
    role,
    permissions: user.permissions ?? permissionsForRole(role),
    enabled: user.enabled ?? status !== 'disabled',
    last_login_at: user.last_login_at,
    created_at: user.created_at ?? '',
    updated_at: user.updated_at ?? ''
  }
}

function normalizeRole(role?: string): UserRole {
  switch (role) {
    case 'cluster-admin':
    case 'security-admin':
    case 'storage-admin':
    case 'operator':
    case 'viewer':
      return role
    case 'admin':
      return 'cluster-admin'
    default:
      return 'viewer'
  }
}

function permissionsForRole(role: UserRole) {
  if (role === 'cluster-admin' || role === 'security-admin') {
    return ['cluster:read', 'storage:read', 'system:read']
  }
  if (role === 'storage-admin' || role === 'operator') {
    return ['cluster:read', 'storage:read']
  }
  return ['cluster:read']
}

function storeCurrentUser(user: UserAccount) {
  try {
    localStorage.setItem(storedUserKey, JSON.stringify(user))
  } catch {
    // Ignore storage failures; token auth remains the source of truth.
  }
}

function readStoredUser() {
  try {
    const value = localStorage.getItem(storedUserKey)
    return value ? normalizeUser(JSON.parse(value)) : null
  } catch {
    return null
  }
}

function defaultSetupDatabase(): SetupDatabaseConfig {
  return {
    engine: 'sqlite',
    sqlite: { name: 'cephtower.db' },
    mysql: {
      host: '127.0.0.1',
      port: 3306,
      username: 'root',
      password: '',
      password_set: false,
      database: 'cephtower',
      params: 'charset=utf8mb4&parseTime=True&loc=Local',
      tls: 'false'
    }
  }
}

function isListPayload(value: unknown): value is { items: Partial<UserAccount>[] } {
  return typeof value === 'object' && value !== null && Array.isArray((value as { items?: unknown }).items)
}
