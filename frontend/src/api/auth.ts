import { getAuthToken, setAuthToken } from './client'

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
  const response = await fetch(`${authBaseUrl}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  })
  const payload = await readJSON<LoginResponse>(response)
  setAuthToken(payload.token)
  return payload
}

export async function currentUser(): Promise<UserAccount> {
  return requestAuth<UserAccount>('/auth/me')
}

export async function setupStatus(): Promise<SetupStatus> {
  return requestPublic<SetupStatus>('/setup/status')
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

async function requestAuth<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${authBaseUrl}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(getAuthToken() ? { Authorization: `Bearer ${getAuthToken()}` } : {}),
      ...init?.headers
    }
  })
  return readJSON<T>(response)
}

async function requestPublic<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${authBaseUrl}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers
    }
  })
  return readJSON<T>(response)
}

async function readJSON<T>(response: Response): Promise<T> {
  const text = await response.text()
  const payload = text ? JSON.parse(text) : {}
  if (!response.ok) {
    const message = payload?.error ?? text ?? `Request failed: ${response.status}`
    throw new Error(message)
  }
  return payload as T
}
