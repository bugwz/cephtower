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

export function listDaemons(types?: string): Promise<ApiRecord[]> {
  const query = types ? `?types=${encodeURIComponent(types)}` : ''
  return request<unknown>(`/daemons${query}`).then(asArray)
}

export function listServices(): Promise<ApiRecord[]> {
  return request<unknown>('/services').then(asArray)
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

