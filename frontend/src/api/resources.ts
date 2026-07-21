import { asArray, request, type ApiRecord } from './client'

export function listHosts(): Promise<ApiRecord[]> {
  return request<unknown>('/hosts?include_service_instances=true').then(asArray)
}

export function listOSDs(): Promise<ApiRecord[]> {
  return request<unknown>('/osds').then(asArray)
}

export function listOSDFlags(): Promise<string[]> {
  return request<{ flags?: string[] }>('/osds/flags').then((payload) => payload.flags ?? [])
}

export function markOSD(id: string, action: string): Promise<unknown> {
  return request(`/osds/${encodeURIComponent(id)}/mark`, jsonInit('PUT', { action }))
}

export function reweightOSD(id: string, weight: number): Promise<unknown> {
  return request(`/osds/${encodeURIComponent(id)}/reweight`, jsonInit('POST', { weight }))
}

export function scrubOSD(id: string, deep = false): Promise<unknown> {
  return request(`/osds/${encodeURIComponent(id)}/scrub`, jsonInit('POST', { deep }))
}

export function listDaemons(types?: string): Promise<ApiRecord[]> {
  const query = types ? `?types=${encodeURIComponent(types)}` : ''
  return request<unknown>(`/daemons${query}`).then(asArray)
}

export function applyDaemonAction(name: string, action: string, force = false): Promise<unknown> {
  return request(`/daemons/${encodeURIComponent(name)}/action`, jsonInit('PUT', { action, force }))
}

export function listServices(): Promise<ApiRecord[]> {
  return request<unknown>('/services').then(asArray)
}

export function listMonitors(): Promise<ApiRecord> {
  return request<ApiRecord>('/monitors')
}

export function listMgrModules(): Promise<ApiRecord[]> {
  return request<unknown>('/mgr/modules').then(asArray)
}

export function setMgrModuleEnabled(name: string, enabled: boolean): Promise<unknown> {
  const action = enabled ? 'enable' : 'disable'
  return request(`/mgr/modules/${encodeURIComponent(name)}/${action}`, { method: 'POST' })
}

export function listPools(): Promise<ApiRecord[]> {
  return request<unknown>('/pools').then(asArray)
}

export function listBlockImages(): Promise<ApiRecord[]> {
  return request<unknown>('/block/images').then(asArray)
}

export function getBlockMirroringSummary(): Promise<ApiRecord> {
  return request<ApiRecord>('/block/mirroring/summary')
}

export function listFilesystems(): Promise<ApiRecord[]> {
  return request<unknown>('/filesystems').then(asArray)
}

export function listObjectGateways(): Promise<ApiRecord[]> {
  return request<unknown>('/object/gateways').then(asArray)
}

export function listObjectUsers(): Promise<ApiRecord[]> {
  return request<unknown>('/object/users').then(asArray)
}

export function listObjectBuckets(): Promise<ApiRecord[]> {
  return request<unknown>('/object/buckets').then(asArray)
}

export function listConfiguration(): Promise<ApiRecord[]> {
  return request<unknown>('/configuration').then(asArray)
}

export function listLogs(): Promise<ApiRecord> {
  return request<ApiRecord>('/logs')
}

function jsonInit(method: string, body: unknown): RequestInit {
  return {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  }
}
