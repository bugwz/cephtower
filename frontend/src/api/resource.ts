import { asArray, isApiError, jsonInit, request, requestWithResponse, textValue, type ApiRecord, type ApiRequestInit } from './client'
import type { ActionResult, ListEnvelope, ResourceDTO } from './types'

export const selectedClusterStorageKey = 'cephtower.selectedClusterId'

export interface ResourceListOptions {
  limit?: number
  cursor?: string
  name?: string
  status?: string
  body?: ApiRecord
}

export interface ResourceListResult<T = ApiRecord> {
  items: Array<T & ApiRecord>
  nextCursor?: string | null
  observedAt?: string | null
  stale: boolean
  staleReason?: string | null
}

export interface ResourceItemResult<T = ApiRecord> {
  item: ResourceDTO<T>
  etag?: string | null
}

export function currentClusterId() {
  try {
    const raw = localStorage.getItem(selectedClusterStorageKey)
    const value = raw ? Number(raw) : 0
    return Number.isFinite(value) && value > 0 ? value : undefined
  } catch {
    return undefined
  }
}

export function setStoredClusterId(clusterId: number | undefined) {
  try {
    if (clusterId) {
      localStorage.setItem(selectedClusterStorageKey, String(clusterId))
    } else {
      localStorage.removeItem(selectedClusterStorageKey)
    }
  } catch {
    // Ignore storage failures; the context still keeps the in-memory value.
  }
}

export async function listResource<T = ApiRecord>(path: string, clusterId = requiredClusterId(), options: ResourceListOptions = {}) {
  const query = new URLSearchParams()
  if (options.limit) {
    query.set('limit', String(options.limit))
  }
  if (options.cursor) {
    query.set('cursor', options.cursor)
  }
  if (options.name) {
    query.set('name', options.name)
  }
  if (options.status) {
    query.set('status', options.status)
  }
  const suffix = query.toString() ? `?${query}` : ''
  let payload: ListEnvelope<ResourceDTO<T>> | ResourceDTO<T>
  try {
    payload = await request<ListEnvelope<ResourceDTO<T>> | ResourceDTO<T>>(`${path}${suffix}`, jsonInit('GET', {
      cluster_id: clusterId,
      ...options.body
    }))
  } catch (err) {
    if (isResourceUnavailable(err)) {
      return emptyResourceList<T>()
    }
    throw err
  }
  const resourceItems = 'items' in payload ? payload.items ?? [] : [payload]
  const meta = 'items' in payload ? payload.meta : undefined
  const pagination = 'items' in payload ? payload.pagination : undefined
  const items = resourceItems.map((item) => {
    const data = toRecord(item.data)
    return {
    ...data,
    kind: item.kind,
    natural_key: item.natural_key,
    name: item.name ?? data.name,
    status: item.status ?? data.status,
    resource_version: item.resource_version,
    source: item.source,
    observed_at: item.observed_at,
    created_at: item.created_at,
    updated_at: item.updated_at,
    stale: item.stale
    }
  }) as unknown as Array<T & ApiRecord>
  return {
    items,
    nextCursor: pagination?.next_cursor,
    observedAt: meta?.observed_at ?? resourceItems[0]?.observed_at,
    stale: Boolean(meta?.stale ?? resourceItems.some((item) => item.stale)),
    staleReason: meta?.stale_reason
  } satisfies ResourceListResult<T>
}

export async function getResource<T = ApiRecord>(path: string, clusterId = requiredClusterId(), body: ApiRecord = {}, init?: ApiRequestInit) {
  const { data, response } = await requestWithResponse<ResourceDTO<T>>(path, jsonInit('GET', {
    cluster_id: clusterId,
    ...body
  }, init))
  return { item: data, etag: response.headers.get('ETag') } satisfies ResourceItemResult<T>
}

export async function getOptionalResource<T = ApiRecord>(path: string, clusterId = requiredClusterId(), body: ApiRecord = {}) {
  try {
    return await getResource<T>(path, clusterId, body, { suppressErrorNotification: true })
  } catch (err) {
    if (isResourceNotFound(err)) {
      return undefined
    }
    throw err
  }
}

export async function mutateResource(path: string, method: string, body: ApiRecord, options?: { ifMatch?: number | string }) {
  return request<ActionResult>(path, {
    ...jsonInit(method, body),
    headers: {
      'Content-Type': 'application/json',
      ...(options?.ifMatch ? { 'If-Match': String(options.ifMatch) } : {})
    }
  })
}

export async function refreshResource(input: { clusterId?: number, kind?: string, kinds?: string[], module?: string, modules?: string[], scope?: 'all' }) {
  const clusterId = input.clusterId ?? requiredClusterId()
  return request<ActionResult>('/resource/refresh', jsonInit('POST', {
    cluster_id: clusterId,
    ...(input.scope ? { scope: input.scope } : {}),
    ...(input.kind ? { kind: input.kind } : {}),
    ...(input.kinds && input.kinds.length > 0 ? { kinds: input.kinds } : {}),
    ...(input.module ? { module: input.module } : {}),
    ...(input.modules && input.modules.length > 0 ? { modules: input.modules } : {})
  }))
}

export async function listHosts(): Promise<ApiRecord[]> {
  return listResource('/hosts').then((payload) => payload.items)
}

export async function listHostDevices(host: string): Promise<ApiRecord[]> {
  return listResource('/host/devices', requiredClusterId(), { body: { host } }).then((payload) => payload.items)
}

export async function listOSDs(): Promise<ApiRecord[]> {
  return listResource('/osds').then((payload) => payload.items)
}

export async function listOSDFlags(): Promise<string[]> {
  const payload = await getOptionalResource<ApiRecord>('/osd/flag', requiredClusterId(), {})
  if (!payload) {
    return []
  }
  const data = toRecord(payload.item.data)
  const flags = data.flags
  if (Array.isArray(flags)) {
    return flags.map((flag) => textValue(flag, '')).filter(Boolean)
  }
  if (typeof flags === 'string') {
    return flags.split(',').map((flag) => flag.trim()).filter(Boolean)
  }
  return []
}

export function markOSD(id: string, action: string): Promise<ActionResult> {
  return mutateResource('/osd/action', 'POST', { cluster_id: requiredClusterId(), osd_id: id, action })
}

export function reweightOSD(id: string, weight: number): Promise<ActionResult> {
  return mutateResource('/osd/action', 'POST', { cluster_id: requiredClusterId(), osd_id: id, action: 'reweight', weight })
}

export function scrubOSD(id: string, deep = false): Promise<ActionResult> {
  return mutateResource('/osd/action', 'POST', { cluster_id: requiredClusterId(), osd_id: id, action: deep ? 'deep-scrub' : 'scrub' })
}

export function listDaemons(types?: string): Promise<ApiRecord[]> {
  return listResource('/daemons', requiredClusterId(), { body: types ? { daemon_type: types } : undefined }).then((payload) => payload.items)
}

export interface HostSSHPayload {
  cluster_id?: number
  hostname: string
  ssh_address: string
  ssh_port?: number
  ssh_user: string
  ssh_auth_method: string
  ssh_password?: string
  ssh_private_key?: string
  ssh_key_passphrase?: string
  notes?: string
}

export function getHostSSH(hostname: string, clusterId = requiredClusterId()): Promise<ApiRecord> {
  return request<ApiRecord>('/host/ssh', jsonInit('GET', { cluster_id: clusterId, hostname }))
}

export function saveHostSSH(values: HostSSHPayload, clusterId = requiredClusterId()): Promise<ApiRecord> {
  return request<ApiRecord>('/host/ssh', jsonInit('PATCH', {
    ...values,
    cluster_id: clusterId
  }))
}

export function applyDaemonAction(name: string, action: string, force = false): Promise<ActionResult> {
  return mutateResource('/daemon/action', 'POST', { cluster_id: requiredClusterId(), name, action, force })
}

export function listServices(): Promise<ApiRecord[]> {
  return listResource('/services').then((payload) => payload.items)
}

export function listMonitors(): Promise<ApiRecord> {
  return listResource('/monitors').then((payload) => ({ items: payload.items }))
}

export function listMgrModules(): Promise<ApiRecord[]> {
  return listResource('/manager/modules').then((payload) => payload.items)
}

export function setMgrModuleEnabled(name: string, enabled: boolean): Promise<ActionResult> {
  return mutateResource('/manager/module', 'PATCH', { cluster_id: requiredClusterId(), name, enabled })
}

export function listPools(): Promise<ApiRecord[]> {
  return listResource('/pools').then((payload) => payload.items)
}

export function listBlockImages(): Promise<ApiRecord[]> {
  return listResource('/rbd/images').then((payload) => payload.items)
}

export function getBlockMirroringSummary(): Promise<ApiRecord> {
  return getOptionalResource('/rbd/mirroring').then((payload) => toRecord(payload?.item.data))
}

export function listFilesystems(): Promise<ApiRecord[]> {
  return listResource('/filesystems').then((payload) => payload.items)
}

export function listObjectGateways(): Promise<ApiRecord[]> {
  return listResource('/services', requiredClusterId(), { body: { service_type: 'rgw' } }).then((payload) => payload.items)
}

export function listObjectUsers(): Promise<ApiRecord[]> {
  return listResource('/rgw/users').then((payload) => payload.items)
}

export function listObjectBuckets(): Promise<ApiRecord[]> {
  return listResource('/rgw/buckets').then((payload) => payload.items)
}

export function listConfiguration(): Promise<ApiRecord[]> {
  return listResource('/configuration/values').then((payload) => payload.items)
}

export function listLogs(): Promise<ApiRecord> {
  return listResource('/logs').then((payload) => ({ items: payload.items }))
}

export function unwrapList(payload: unknown): ApiRecord[] {
  return asArray(payload)
}

function requiredClusterId() {
  const clusterId = currentClusterId()
  if (!clusterId) {
    throw new Error('请先选择集群')
  }
  return clusterId
}

function emptyResourceList<T = ApiRecord>() {
  return {
    items: [],
    nextCursor: null,
    observedAt: null,
    stale: false,
    staleReason: null
  } satisfies ResourceListResult<T>
}

function isResourceNotFound(err: unknown) {
  return isApiError(err, 404, 'resource_not_found') || (isApiError(err, 404) && !err.code)
}

function isResourceUnavailable(err: unknown) {
  return isResourceNotFound(err) || isApiError(err, 501, 'capability_unavailable') || (isApiError(err, 501) && !err.code)
}

function toRecord(value: unknown): ApiRecord {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as ApiRecord : {}
}
