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

export interface CephClusterPayload {
  name: string
  description?: string
  fsid?: string
  enabled: boolean
  dashboard: {
    enabled: boolean
    base_url?: string
    username?: string
    password?: string
    clear_secret?: boolean
    insecure_tls?: boolean
  }
  command: {
    enabled: boolean
    bin?: string
    cluster?: string
    conf?: string
    name?: string
    keyring?: string
    keyring_content?: string
    clear_secret?: boolean
    timeout_seconds?: number
  }
}

const apiBaseUrl = '/api/v1'

export async function listClusters(): Promise<CephCluster[]> {
  return clusterRequest<CephCluster[]>('/clusters')
}

export async function createCluster(payload: CephClusterPayload): Promise<CephCluster> {
  return clusterRequest<CephCluster>('/clusters', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export async function updateCluster(id: number, payload: CephClusterPayload): Promise<CephCluster> {
  return clusterRequest<CephCluster>(`/clusters/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload)
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
