import { request } from './client'

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
  return request<SystemSetting[]>(`/system/config/settings${query}`)
}

export function updateSystemSetting(key: string, value: string): Promise<SystemSetting> {
  return request<SystemSetting>(`/system/config/settings/${encodeURIComponent(key)}`, {
    method: 'PUT',
    body: JSON.stringify({ value })
  })
}

export function runDataFetchModule(module: string): Promise<{ message: string }> {
  return request<{ message: string }>(`/system/config/data-fetch/${encodeURIComponent(module)}/run`, {
    method: 'POST'
  })
}

export function listDataFetchRuns(limit = 50): Promise<DataFetchRun[]> {
  return request<DataFetchRun[]>(`/system/config/data-fetch/runs?limit=${limit}`)
}

export function resetSystemConfigDefaults(): Promise<{ message: string }> {
  return request<{ message: string }>('/system/config/defaults/reset', {
    method: 'POST'
  })
}
