import { asArray, request, type ApiRecord } from './client'

export function listHosts(): Promise<ApiRecord[]> {
  return request<unknown>('/host?include_service_instances=true').then(asArray)
}

export function listOSDs(): Promise<ApiRecord[]> {
  return request<unknown>('/osd').then(asArray)
}

export function listOSDFlags(): Promise<string[]> {
  return request<{ flags?: string[] }>('/osd/flag').then((payload) => payload.flags ?? [])
}

export function markOSD(id: string, action: string): Promise<unknown> {
  return request(`/osd/${encodeURIComponent(id)}/mark`, jsonInit('PUT', { action }))
}

export function reweightOSD(id: string, weight: number): Promise<unknown> {
  return request(`/osd/${encodeURIComponent(id)}/reweight`, jsonInit('POST', { weight }))
}

export function scrubOSD(id: string, deep = false): Promise<unknown> {
  return request(`/osd/${encodeURIComponent(id)}/scrub`, jsonInit('POST', { deep }))
}

export function listDaemons(types?: string): Promise<ApiRecord[]> {
  const query = types ? `?types=${encodeURIComponent(types)}` : ''
  return request<unknown>(`/daemon${query}`).then(asArray)
}

export function applyDaemonAction(name: string, action: string, force = false): Promise<unknown> {
  return request(`/daemon/${encodeURIComponent(name)}/action`, jsonInit('PUT', { action, force }))
}

export function listServices(): Promise<ApiRecord[]> {
  return request<unknown>('/service').then(asArray)
}

export function listMonitors(): Promise<ApiRecord> {
  return request<ApiRecord>('/monitor')
}

export function listMgrModules(): Promise<ApiRecord[]> {
  return request<unknown>('/mgr/module').then(asArray)
}

export function setMgrModuleEnabled(name: string, enabled: boolean): Promise<unknown> {
  const action = enabled ? 'enable' : 'disable'
  return request(`/mgr/module/${encodeURIComponent(name)}/${action}`, { method: 'POST' })
}

export function listPools(): Promise<ApiRecord[]> {
  return request<unknown>('/pool').then(asArray)
}

export function listBlockImages(): Promise<ApiRecord[]> {
  return request<unknown>('/block/image').then(asArray)
}

export function getBlockMirroringSummary(): Promise<ApiRecord> {
  return request<ApiRecord>('/block/mirroring/summary')
}

export function listFilesystems(): Promise<ApiRecord[]> {
  return request<unknown>('/filesystem').then(asArray)
}

export function listObjectGateways(): Promise<ApiRecord[]> {
  return request<unknown>('/object/gateway').then(asArray)
}

export function listObjectUsers(): Promise<ApiRecord[]> {
  return request<unknown>('/object/user').then(asArray)
}

export function listObjectBuckets(): Promise<ApiRecord[]> {
  return request<unknown>('/object/bucket').then(asArray)
}

export function listConfiguration(): Promise<ApiRecord[]> {
  return request<unknown>('/configuration').then(asArray)
}

export function listLogs(): Promise<ApiRecord> {
  return request<ApiRecord>('/log')
}

function jsonInit(method: string, body: unknown): RequestInit {
  return {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  }
}
