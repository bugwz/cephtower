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

export interface ResourceDTO<T = ApiRecord> {
  kind: string
  natural_key: string
  name?: string | null
  status?: string | null
  resource_version: number
  source: string
  observed_at: string
  stale: boolean
  data: T
}

export interface ActionResult {
  resource_url?: string
  details?: unknown
}

export type OperationRisk = 'low' | 'medium' | 'high'
