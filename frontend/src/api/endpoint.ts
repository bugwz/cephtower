import { jsonInit, request, type ApiRecord } from './client'
import type { ListEnvelope } from './types'

export interface CredentialView {
  kind: string
  fingerprint: string
  created_at: string
  updated_at: string
}

export interface EndpointView {
  endpoint_id: number
  kind: string
  name: string
  url: string
  tls_mode: string
  ca_credential_id?: number | null
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface EndpointInput {
  kind: string
  name?: string
  url: string
  tls_mode?: string
  ca_credential_id?: number | null
  timeout_seconds?: number
  enabled?: boolean
}

export function listCredentials(clusterId: number) {
  return request<ListEnvelope<CredentialView>>('/credentials', jsonInit('GET', { cluster_id: clusterId }))
    .then((payload) => payload.items ?? [])
}

export function putCredential(clusterId: number, kind: string, credential: ApiRecord) {
  return request<CredentialView>('/credential', jsonInit('PUT', { cluster_id: clusterId, kind, credential }))
}

export function deleteCredential(clusterId: number, kind: string) {
  return request<{ message?: string }>('/credential', jsonInit('DELETE', { cluster_id: clusterId, kind }))
}

export function listEndpoints(clusterId: number) {
  return request<ListEnvelope<EndpointView>>('/endpoints', jsonInit('GET', { cluster_id: clusterId }))
    .then((payload) => payload.items ?? [])
}

export function createEndpoint(clusterId: number, input: EndpointInput) {
  return request<EndpointView>('/endpoint', jsonInit('POST', { cluster_id: clusterId, ...input }))
}

export function updateEndpoint(clusterId: number, endpointId: number, input: EndpointInput) {
  return request<EndpointView>('/endpoint', jsonInit('PATCH', { cluster_id: clusterId, endpoint_id: endpointId, ...input }))
}

export function deleteEndpoint(clusterId: number, endpointId: number) {
  return request<{ message?: string }>('/endpoint', jsonInit('DELETE', { cluster_id: clusterId, endpoint_id: endpointId }))
}

