import {
  getAuthToken,
  notifyApiError,
  readApiResponse,
  toApiErrorDetail,
  type ApiRequestInit
} from './client'

export interface SystemSetting {
  key: string
  value: string
  updated_at: string
}

export interface DataFetchConfig {
  module: string
  enabled: boolean
  interval_seconds: number
  timeout_seconds: number
  jitter_seconds: number
  fetch_source: string
  priority: number
  max_retries: number
  retry_backoff_seconds: number
}

export interface DataFetchRun {
  id: number
  cluster_id: number
  module: string
  status: string
  source: string
  started_at: string
  finished_at?: string
  duration_ms: number
  records_upserted: number
  records_deleted: number
  error: string
  created_at: string
}

export function listSystemSettings(prefix?: string): Promise<SystemSetting[]> {
  const query = prefix ? `?prefix=${encodeURIComponent(prefix)}` : ''
  return systemRequest<SystemSetting[]>(`/system/config/setting${query}`)
}

export function updateSystemSetting(key: string, value: string): Promise<SystemSetting> {
  return systemRequest<SystemSetting>(`/system/config/setting/${encodeURIComponent(key)}`, {
    method: 'PUT',
    body: JSON.stringify({ value })
  })
}

export function runDataFetchModule(module: string): Promise<{ message: string }> {
  return systemRequest<{ message: string }>(`/system/config/data-fetch/${encodeURIComponent(module)}/run`, {
    method: 'POST'
  })
}

export function listDataFetchRuns(limit = 50): Promise<DataFetchRun[]> {
  return systemRequest<DataFetchRun[]>(`/system/config/data-fetch/run?limit=${limit}`)
}

export function resetSystemConfigDefaults(): Promise<{ message: string }> {
  return systemRequest<{ message: string }>('/system/config/default/reset', {
    method: 'POST'
  })
}

const systemBaseUrl = '/api/v1'

async function systemRequest<T>(path: string, init?: ApiRequestInit): Promise<T> {
  const { suppressErrorNotification, ...fetchInit } = init ?? {}
  try {
    const response = await fetch(`${systemBaseUrl}${path}`, {
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
