import { request, type ApiRecord } from './client'

export interface TaskSummary {
  name?: string
  begin_time?: string
  end_time?: string
  duration?: number
  progress?: number
  success?: boolean
  ret_value?: string
  exception?: string
  metadata?: ApiRecord
}

export interface ClusterSummary {
  health_status: string
  version?: string
  mgr_id?: string
  mgr_host?: string
  have_mon_connection?: string
  executing_tasks?: string[]
  finished_tasks?: TaskSummary[]
  rbd_mirroring?: Record<string, number>
}

export function getClusterSummary(): Promise<ClusterSummary> {
  return request<ClusterSummary>('/cluster/summary')
}

export function getClusterHealth(): Promise<ApiRecord> {
  return request<ApiRecord>('/cluster/health/full')
}

