import {
  getAuthToken,
  notifyApiError,
  readApiResponse,
  toApiErrorDetail,
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
  dashboard_username?: string
  dashboard_password?: string
  keyring?: string
}

export interface ClusterActionResponse {
  message: string
}

export interface CephResourceSnapshot {
  category: string
  resource_key: string
  payload: unknown
  last_synced_at: string
  last_error: string
}

export interface CephClusterDetail {
  cluster: CephCluster
  snapshots: CephResourceSnapshot[]
}

const apiBaseUrl = '/api/v1'

export async function listClusters(): Promise<CephCluster[]> {
  return clusterRequest<CephCluster[]>('/clusters')
}

export async function getClusterDetail(id: number | string): Promise<CephClusterDetail> {
  return clusterRequest<CephClusterDetail>(`/clusters/${id}`)
}

export async function createCluster(values: CephClusterFormPayload): Promise<ClusterActionResponse> {
  return clusterRequest<ClusterActionResponse>('/clusters', {
    method: 'POST',
    body: JSON.stringify(values)
  })
}

export async function updateCluster(id: number, values: CephClusterFormPayload): Promise<ClusterActionResponse> {
  return clusterRequest<ClusterActionResponse>(`/clusters/${id}`, {
    method: 'PUT',
    body: JSON.stringify(values)
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
