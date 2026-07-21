import {
  getAuthToken,
  notifyApiError,
  readApiResponse,
  setAuthToken,
  toApiErrorDetail,
  type ApiRequestInit
} from './client'

export type UserRole = 'admin' | 'user'

export interface UserAccount {
  id: number
  username: string
  display_name: string
  email: string
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
    path: string
  }
  mysql: {
    host: string
    port: number
    username: string
    password: string
    password_set: boolean
    database: string
    params: string
  }
}

export interface SetupStatus {
  initialized: boolean
  database?: SetupDatabaseConfig
}

interface LoginResponse {
  token: string
  expires_at: string
  user: UserAccount
}

const authBaseUrl = '/api/v1'

export async function login(username: string, password: string): Promise<LoginResponse> {
  return requestPublic<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password })
  }).then((payload) => {
    setAuthToken(payload.token)
    return payload
  })
}

export async function currentUser(): Promise<UserAccount> {
  return requestAuth<UserAccount>('/auth/me')
}

export async function setupStatus(): Promise<SetupStatus> {
  return requestPublic<SetupStatus>('/setup/status', { suppressErrorNotification: true })
}

export async function initializeSetup(payload: {
  database: {
    engine: 'sqlite' | 'mysql'
    sqlite: {
      path: string
    }
    mysql: {
      host: string
      port: number
      username: string
      password: string
      database: string
      params: string
    }
  }
  admin: {
    username: string
    email: string
    password: string
  }
}): Promise<{ message: string }> {
  return requestPublic<{ message: string }>('/setup/initialize', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export async function listUsers(): Promise<UserAccount[]> {
  return requestAuth<UserAccount[]>('/users')
}

export async function createUser(payload: {
  username: string
  display_name: string
  email?: string
  role: UserRole
  permissions: string[]
  password: string
  enabled: boolean
}): Promise<UserAccount> {
  return requestAuth<UserAccount>('/users', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export async function updateUser(id: number, payload: Partial<{
  display_name: string
  email: string
  role: UserRole
  permissions: string[]
  password: string
  enabled: boolean
}>): Promise<UserAccount> {
  return requestAuth<UserAccount>(`/users/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(payload)
  })
}

export async function requestPasswordReset(account: string): Promise<{ message: string }> {
  return requestPublic<{ message: string }>('/auth/password-reset/request', {
    method: 'POST',
    body: JSON.stringify({ account })
  })
}

export async function confirmPasswordReset(payload: {
  account: string
  code: string
  new_password: string
}): Promise<{ message: string }> {
  return requestPublic<{ message: string }>('/auth/password-reset/confirm', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export function logout() {
  setAuthToken('')
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
