import {
  getAuthToken,
  notifyApiError,
  readApiResponse,
  request,
  toApiErrorDetail,
  type ApiRecord,
  type ApiRequestInit
} from './client'

export interface CephCluster {
  id: number
  name: string
  description: string
  fsid: string
  enabled: boolean
  dashboard: {
    enabled: boolean
    base_url: string
    username: string
    password_set: boolean
    insecure_tls: boolean
  }
  command: {
    enabled: boolean
    bin: string
    cluster: string
    conf: string
    monitor_host: string
    name: string
    keyring: string
    keyring_content_set: boolean
    timeout_seconds: number
  }
  created_at: string
  updated_at: string
}

export interface CephClusterFormPayload {
  name: string
  monitor_host?: string
  dashboard_username?: string
  dashboard_password?: string
  keyring?: string
}

export interface ClusterActionResponse {
  message: string
}

export interface CephClusterDiscovery {
  hosts: CephDiscoveredRecord[]
  osds: CephDiscoveredRecord[]
  osd_flags: Array<{
    name: string
    discovered_at: string
  }>
  daemons: CephDiscoveredRecord[]
  services: CephDiscoveredRecord[]
  mons: CephDiscoveredRecord[]
  mgrs: CephDiscoveredRecord[]
  mdss: CephDiscoveredRecord[]
  mgr_modules: CephDiscoveredRecord[]
  configuration: CephDiscoveredRecord[]
}

export interface CephDiscoveredRecord {
  key: string
  type?: string
  hostname?: string
  status?: string
  payload: unknown
  discovered_at: string
}

export interface CephClusterDetail {
  cluster: CephCluster
  discovery: CephClusterDiscovery
}

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

export async function listClusters(): Promise<CephCluster[]> {
  return clusterRequest<CephCluster[]>('/cluster')
}

export async function getClusterDetail(id: number | string): Promise<CephClusterDetail> {
  return clusterRequest<CephClusterDetail>(`/cluster/${id}`)
}

export async function getClusterKeyring(id: number | string): Promise<string> {
  const response = await clusterRequest<{ keyring: string }>(`/cluster/${id}/credentials/keyring`)
  return response.keyring
}

export async function getClusterDashboardPassword(id: number | string): Promise<string> {
  const response = await clusterRequest<{ dashboard_password: string }>(`/cluster/${id}/credentials/dashboard-password`)
  return response.dashboard_password
}

export async function createCluster(values: CephClusterFormPayload): Promise<ClusterActionResponse> {
  return clusterRequest<ClusterActionResponse>('/cluster', {
    method: 'POST',
    body: JSON.stringify(values)
  })
}

export async function updateCluster(id: number, values: CephClusterFormPayload): Promise<ClusterActionResponse> {
  return clusterRequest<ClusterActionResponse>(`/cluster/${id}`, {
    method: 'PUT',
    body: JSON.stringify(values)
  })
}

export async function deleteCluster(id: number | string): Promise<ClusterActionResponse> {
  return clusterRequest<ClusterActionResponse>(`/cluster/${id}`, {
    method: 'DELETE'
  })
}

async function clusterRequest<T>(path: string, init?: ApiRequestInit): Promise<T> {
  const { suppressErrorNotification, ...fetchInit } = init ?? {}
  try {
    const response = await fetch(`${apiBaseUrl}${path}`, {
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

const apiBaseUrl = '/api/v1'
