import { asArray, jsonInit, request, type ApiRecord } from './client'
import { listCredentials, listEndpoints, type CredentialView, type EndpointView } from './endpoint'
import { getOptionalResource, listResource } from './resource'
import type { ListEnvelope } from './types'

export interface CephCluster {
  id: number
  cluster_id: number
  name: string
  monitor_addresses: string
  client_username: string
  created_at: string
  updated_at: string
  description: string
  fsid: string
  ceph_version: string
  status: string
  enabled: boolean
  generation: number
  last_seen_at: string
  last_error_code: string
  last_error_message: string
  observed_at: string
  dashboard: {
    enabled: boolean
    base_url: string
    username: string
    password: string
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
}

export interface CephClusterFormPayload {
  name: string
  monitor_addresses?: string
  monitor_host?: string
  client_username?: string
  client_key?: string
  keyring?: string
}

export interface ClusterActionResponse {
  message: string
  cluster?: CephCluster
  result?: ActionResult
}

export interface ActionResult {
  resource_url?: string
  details?: ApiRecord
}

interface ClusterMutationPayload {
  cluster: ApiRecord
  result?: ActionResult
}

export interface CephClusterDiscovery {
  hosts: CephDiscoveredRecord[]
  osds: CephDiscoveredRecord[]
  osd_flags: Array<{ name: string; discovered_at: string }>
  daemons: CephDiscoveredRecord[]
  services: CephDiscoveredRecord[]
  mons: CephDiscoveredRecord[]
  mgrs: CephDiscoveredRecord[]
  mdss: CephDiscoveredRecord[]
  mgr_modules: CephDiscoveredRecord[]
  configuration: CephDiscoveredRecord[]
}

export interface CephClusterDetail {
  cluster: CephCluster
  discovery: CephClusterDiscovery
  overview: ApiRecord
  endpoints: EndpointView[]
  credentials: CredentialView[]
}

export interface CephDiscoveredRecord {
  key: string
  type?: string
  hostname?: string
  status?: string
  payload: unknown
  discovered_at: string
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

export interface ClusterCapability {
  name: string
  supported: boolean
  reason?: string | null
  version?: string | null
  details?: unknown
  observed_at: string
}

export async function listClusters(): Promise<CephCluster[]> {
  const payload = await request<ListEnvelope<ApiRecord> | ApiRecord[]>('/clusters', { method: 'GET' })
  const rows = Array.isArray(payload) ? payload : asArray(payload)
  return rows.map(normalizeCluster)
}

export async function getClusterDetail(name: string): Promise<CephClusterDetail> {
  const cluster = await getClusterByName(name)
  const [
    hosts,
    osds,
    daemons,
    services,
    mons,
    mgrs,
    mdss,
    mgrModules,
    configuration,
    overview,
    endpoints,
    credentials
  ] = await Promise.all([
    detailRows('/hosts', cluster.id, 'hosts'),
    detailRows('/osds', cluster.id, 'osds'),
    detailRows('/daemons', cluster.id, 'daemons'),
    detailRows('/services', cluster.id, 'services'),
    detailRows('/monitors', cluster.id, 'mons'),
    detailRows('/managers', cluster.id, 'mgrs'),
    detailRows('/daemons', cluster.id, 'mdss', { daemon_type: 'mds' }),
    detailRows('/manager/modules', cluster.id, 'mgr_modules'),
    detailRows('/configuration/values', cluster.id, 'configuration'),
    getClusterOverview(cluster.id),
    listEndpoints(cluster.id),
    listCredentials(cluster.id)
  ])
  return {
    cluster,
    discovery: {
      hosts,
      osds,
      osd_flags: [],
      daemons,
      services,
      mons,
      mgrs,
      mdss,
      mgr_modules: mgrModules,
      configuration
    },
    overview,
    endpoints,
    credentials
  }
}

export async function getCluster(id: number): Promise<CephCluster> {
  const payload = await request<ApiRecord>('/cluster', jsonInit('GET', { cluster_id: id }))
  return normalizeCluster(payload)
}

export async function getClusterByName(name: string): Promise<CephCluster> {
  const clusterName = name.trim()
  if (!clusterName) {
    throw new Error('集群名称不能为空')
  }
  const clusters = await listClusters()
  const cluster = clusters.find((item) => item.name === clusterName)
  if (!cluster) {
    throw new Error(`未找到集群：${clusterName}`)
  }
  return cluster
}

export async function createCluster(values: CephClusterFormPayload): Promise<ClusterActionResponse> {
  const payload = await request<ClusterMutationPayload>('/cluster', jsonInit('POST', createPayload(values)))
  return { message: '集群已保存', cluster: normalizeCluster(payload.cluster), result: payload.result }
}

export async function updateCluster(id: number, values: CephClusterFormPayload): Promise<ClusterActionResponse> {
  const result = await request<ActionResult>('/cluster', jsonInit('PATCH', {
    cluster_id: id,
    ...updatePayload(values)
  }))
  return { message: '集群已更新', result }
}

export async function deleteCluster(id: number | string): Promise<ClusterActionResponse> {
  const result = await request<ActionResult>('/cluster', jsonInit('DELETE', {
    cluster_id: Number(id),
    delete_cached_data: true
  }))
  return { message: '集群连接已删除', result }
}

export async function probeCluster(id: number): Promise<ActionResult> {
  return request<ActionResult>('/cluster/probe', jsonInit('POST', { cluster_id: id }))
}

export async function listClusterCapabilities(id: number): Promise<ClusterCapability[]> {
  const payload = await request<ListEnvelope<ClusterCapability>>('/cluster/capabilities', jsonInit('GET', { cluster_id: id }))
  return payload.items ?? []
}

export async function getClusterSummary(clusterId?: number): Promise<ClusterSummary> {
  if (!clusterId) {
    return { health_status: 'UNKNOWN' }
  }
  const overview = await request<ApiRecord>('/overview', jsonInit('GET', { cluster_id: clusterId }))
  const data = (overview.data && typeof overview.data === 'object' ? overview.data : overview) as ApiRecord
  return {
    health_status: String(data.health_status ?? data.status ?? 'UNKNOWN'),
    version: typeof data.version === 'string' ? data.version : undefined
  }
}

export async function getClusterHealth(clusterId?: number): Promise<ApiRecord> {
  if (!clusterId) {
    return {}
  }
  return request<ApiRecord>('/health', jsonInit('GET', { cluster_id: clusterId }))
}

function normalizeCluster(row: ApiRecord): CephCluster {
  const id = Number(row.cluster_id ?? row.id ?? 0)
  const monitorAddresses = String(row.monitor_addresses ?? row.monitor_host ?? '')
  const clientUsername = String(row.client_username ?? 'client.admin')
  return {
    id,
    cluster_id: id,
    name: String(row.name ?? ''),
    monitor_addresses: monitorAddresses,
    client_username: clientUsername,
    created_at: String(row.created_at ?? ''),
    updated_at: String(row.updated_at ?? ''),
    description: String(row.description ?? ''),
    fsid: String(row.fsid ?? ''),
    ceph_version: String(row.ceph_version ?? ''),
    status: String(row.status ?? 'unknown'),
    enabled: row.enabled !== false,
    generation: Number(row.generation ?? 0),
    last_seen_at: String(row.last_seen_at ?? ''),
    last_error_code: String(row.last_error_code ?? ''),
    last_error_message: String(row.last_error_message ?? ''),
    observed_at: String(row.observed_at ?? ''),
    dashboard: {
      enabled: false,
      base_url: '',
      username: '',
      password: '',
      password_set: false,
      insecure_tls: false
    },
    command: {
      enabled: true,
      bin: 'ceph',
      cluster: 'ceph',
      conf: '',
      monitor_host: monitorAddresses,
      name: clientUsername,
      keyring: '',
      keyring_content_set: true,
      timeout_seconds: 30
    }
  }
}

async function detailRows(path: string, clusterId: number, category: string, body?: ApiRecord): Promise<CephDiscoveredRecord[]> {
  const payload = await listResource(path, clusterId, { limit: 200, body })
  return payload.items.map((row, index) => toDiscoveredRecord(row, category, index))
}

async function getClusterOverview(clusterId: number): Promise<ApiRecord> {
  const payload = await getOptionalResource<ApiRecord>('/overview', clusterId)
  const data = payload?.item.data
  return data && typeof data === 'object' && !Array.isArray(data) ? data as ApiRecord : {}
}

function toDiscoveredRecord(row: ApiRecord, category: string, index: number): CephDiscoveredRecord {
  const key = String(row.natural_key ?? row.name ?? row.hostname ?? row.id ?? `${category}-${index + 1}`)
  return {
    key,
    type: String(row.type ?? row.daemon_type ?? row.service_type ?? row.kind ?? ''),
    hostname: String(row.hostname ?? row.host ?? ''),
    status: String(row.status ?? ''),
    payload: row,
    discovered_at: String(row.observed_at ?? '')
  }
}

function createPayload(values: CephClusterFormPayload) {
  return {
    name: values.name,
    monitor_addresses: values.monitor_addresses || values.monitor_host || '',
    client_username: values.client_username || 'client.admin',
    client_key: values.client_key || values.keyring || ''
  }
}

function updatePayload(values: CephClusterFormPayload) {
  const payload: ApiRecord = {}
  if (values.name) {
    payload.name = values.name
  }
  if (values.monitor_addresses || values.monitor_host) {
    payload.monitor_addresses = values.monitor_addresses || values.monitor_host
  }
  if (values.client_username) {
    payload.client_username = values.client_username
  }
  if (values.client_key || values.keyring) {
    payload.client_key = values.client_key || values.keyring
  }
  return payload
}
