import type { ApiRecord } from './client'

export interface ListEnvelope<T> {
  items: T[]
  pagination?: {
    next_cursor?: string | null
  }
  meta?: {
    request_id?: string
    observed_at?: string | null
    stale?: boolean
    stale_reason?: string | null
  }
}

export interface FilterOptionsEnvelope {
  filter_options?: Record<string, string[]>
}

export interface ResourceDTO<T = ApiRecord> {
  kind: string
  natural_key: string
  name?: string | null
  status?: string | null
  resource_version: number
  source: string
  observed_at: string
  created_at: string
  updated_at: string
  stale: boolean
  data: T
}

export interface ActionResult {
  resource_url?: string
  details?: unknown
}

export type OperationStatus = 'queued' | 'running' | 'succeeded' | 'failed'

export interface Operation {
  operation_id: number
  cluster_id: number
  request_id: string
  action: string
  resource_kind: string
  resource_key: string
  risk: string
  status: OperationStatus
  expected_version?: number
  result?: ActionResult
  error_code?: string
  error_message?: string
  retryable: boolean
  attempts: number
  max_attempts: number
  next_attempt_at?: string
  started_at?: string
  finished_at?: string
  created_at: string
  updated_at: string
}

export type OperationRisk = 'low' | 'medium' | 'high'
