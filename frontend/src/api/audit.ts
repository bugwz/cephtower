import { jsonInit, request, type ApiRecord } from './client'
import type { ListEnvelope } from './types'

export interface AuditEventView {
  audit_event_id: number
  occurred_at: string
  event_type: string
  request_id: string
  actor_username: string
  cluster_id?: number | null
  cluster_name?: string | null
  action: string
  resource_kind?: string | null
  resource_key?: string | null
  risk?: string | null
  outcome: string
  http_status?: number | null
  error_code?: string | null
  parameters?: unknown
  details?: unknown
  event_hash: string
}

export interface AuditFilter {
  clusterId: number
  username?: string
  action?: string
  resourceKind?: string
  resourceKey?: string
  userId?: number
  limit?: number
}

export async function listAuditEvents(filter: AuditFilter) {
  const query = new URLSearchParams()
  if (filter.username) {
    query.set('username', filter.username)
  }
  if (filter.action) {
    query.set('action', filter.action)
  }
  if (filter.resourceKind) {
    query.set('resource_kind', filter.resourceKind)
  }
  if (filter.resourceKey) {
    query.set('resource_key', filter.resourceKey)
  }
  if (filter.userId) {
    query.set('user_id', String(filter.userId))
  }
  if (filter.limit) {
    query.set('limit', String(filter.limit))
  }
  const suffix = query.toString() ? `?${query}` : ''
  const payload = await request<ListEnvelope<AuditEventView>>(`/audit/events${suffix}`, jsonInit('GET', { cluster_id: filter.clusterId } satisfies ApiRecord))
  return payload.items ?? []
}
