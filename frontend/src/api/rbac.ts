import { jsonInit, request } from './client'
import type { ListEnvelope } from './types'

export interface RoleView {
  id: number
  name: string
  description?: string | null
  created_at: string
  updated_at: string
}

export interface RoleBindingView {
  role_binding_id: number
  user_id: number
  username: string
  role: string
  cluster_id: number
  created_at: string
}

export function listRoles() {
  return request<ListEnvelope<RoleView>>('/role', { method: 'GET' })
    .then((payload) => payload.items ?? [])
}

export function createRole(input: { name: string; description?: string }) {
  return request<RoleView>('/role', jsonInit('POST', input))
}

export function listRoleBindings(clusterId: number) {
  return request<ListEnvelope<RoleBindingView>>('/role/bindings', jsonInit('GET', { cluster_id: clusterId }))
    .then((payload) => payload.items ?? [])
}

export function createRoleBinding(input: { clusterId: number; userId: number; role: string }) {
  return request<RoleBindingView>('/role/binding', jsonInit('POST', {
    cluster_id: input.clusterId,
    user_id: input.userId,
    role: input.role
  }))
}

export function deleteRoleBinding(clusterId: number, bindingId: number) {
  return request<{ message?: string }>('/role/binding', jsonInit('DELETE', {
    cluster_id: clusterId,
    binding_id: bindingId
  }))
}
