export type ApiRecord = Record<string, unknown>

const apiBaseUrl = '/api/v1/ceph'

let authToken = localStorage.getItem('cephtower.auth.token') ?? ''

export function setAuthToken(token: string) {
  authToken = token
  if (token) {
    localStorage.setItem('cephtower.auth.token', token)
  } else {
    localStorage.removeItem('cephtower.auth.token')
  }
}

export function getAuthToken() {
  return authToken
}

export async function request<T>(path: string): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    headers: authToken ? { Authorization: `Bearer ${authToken}` } : undefined
  })
  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || `Request failed: ${response.status}`)
  }

  return response.json() as Promise<T>
}

export function asArray(payload: unknown): ApiRecord[] {
  if (Array.isArray(payload)) {
    return payload.filter(isRecord)
  }

  if (!isRecord(payload)) {
    return []
  }

  for (const key of ['items', 'data', 'result', 'results', 'value']) {
    const value = payload[key]
    if (Array.isArray(value)) {
      return value.filter(isRecord)
    }
  }

  return [payload]
}

export function isRecord(value: unknown): value is ApiRecord {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

export function textValue(value: unknown, fallback = '—'): string {
  if (value === null || value === undefined || value === '') {
    return fallback
  }

  if (Array.isArray(value)) {
    return value.length ? value.map((item) => textValue(item, '')).filter(Boolean).join(', ') : fallback
  }

  if (typeof value === 'object') {
    return JSON.stringify(value)
  }

  return String(value)
}

export function numberValue(value: unknown): number | undefined {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }

  if (typeof value === 'string') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : undefined
  }

  return undefined
}
