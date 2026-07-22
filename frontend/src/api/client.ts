export type ApiRecord = Record<string, unknown>

const apiBaseUrl = '/api/v1/ceph'
const apiErrorEventName = 'cephtower:api-error'

let authToken = localStorage.getItem('cephtower.auth.token') ?? ''

export interface ApiRequestInit extends RequestInit {
  suppressErrorNotification?: boolean
}

export interface ApiErrorDetail {
  message: string
  status?: number
  path?: string
}

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

export async function request<T>(path: string, init?: ApiRequestInit): Promise<T> {
  const { suppressErrorNotification, ...fetchInit } = init ?? {}
  try {
    const response = await fetch(`${apiBaseUrl}${path}`, {
      ...fetchInit,
      headers: {
        ...(authToken ? { Authorization: `Bearer ${authToken}` } : {}),
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

export async function readApiResponse<T>(response: Response): Promise<T> {
  const text = await response.text()
  const { payload, invalidJSON } = parseJSON(text)

  if (!response.ok) {
    throw new Error(extractErrorMessage(payload, text, response.status))
  }

  if (invalidJSON) {
    throw new Error('后端接口返回的数据格式不正确')
  }

  if (isRecord(payload) && 'code' in payload && 'message' in payload && 'data' in payload) {
    const code = numberValue(payload.code)
    const message = typeof payload.message === 'string' ? payload.message : ''
    if (code !== undefined && code !== 0) {
      throw new Error(message || `Request failed: ${code}`)
    }
    if (payload.data !== null && payload.data !== undefined) {
      return payload.data as T
    }
    return { message } as T
  }

  return (payload ?? {}) as T
}

export function notifyApiError(detail: ApiErrorDetail) {
  window.dispatchEvent(new CustomEvent<ApiErrorDetail>(apiErrorEventName, { detail }))
}

export function subscribeApiErrors(listener: (detail: ApiErrorDetail) => void) {
  const handler = (event: Event) => {
    listener((event as CustomEvent<ApiErrorDetail>).detail)
  }

  window.addEventListener(apiErrorEventName, handler)
  return () => window.removeEventListener(apiErrorEventName, handler)
}

export function toApiErrorDetail(err: unknown, path?: string): ApiErrorDetail {
  return {
    message: formatApiErrorMessage(err),
    path
  }
}

function parseJSON(text: string): { payload: unknown; invalidJSON: boolean } {
  if (!text) {
    return { payload: undefined, invalidJSON: false }
  }

  try {
    return { payload: JSON.parse(text), invalidJSON: false }
  } catch {
    return { payload: undefined, invalidJSON: true }
  }
}

function extractErrorMessage(payload: unknown, text: string, status: number): string {
  if (isRecord(payload) && 'code' in payload && 'message' in payload && 'data' in payload) {
    const message = payload.message
    if (typeof message === 'string' && message.trim()) {
      return message
    }
  }

  if (isRecord(payload)) {
    const message = payload.error ?? payload.message
    if (typeof message === 'string' && message.trim()) {
      return message
    }
  }

  return text || `Request failed: ${status}`
}

function formatApiErrorMessage(err: unknown): string {
  if (!(err instanceof Error)) {
    return '后端接口调用失败'
  }

  if (err instanceof TypeError && /failed to fetch|networkerror|load failed/i.test(err.message)) {
    return '无法连接后端服务，请检查网络或后端服务状态'
  }

  return err.message || '后端接口调用失败'
}
