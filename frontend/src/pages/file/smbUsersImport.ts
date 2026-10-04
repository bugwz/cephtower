import { parseDocument } from 'yaml'
import type { ApiRecord } from '../../api/client'

const record = (value: unknown): value is Record<string, unknown> => !!value && typeof value === 'object' && !Array.isArray(value)
const id = (value: unknown): value is string => typeof value === 'string' && /^[a-zA-Z0-9]([a-zA-Z0-9-]{0,16}[a-zA-Z0-9])?$/.test(value)
const fields = (value: Record<string, unknown>, allowed: string[]) => Object.keys(value).every(key => allowed.includes(key))

export function smbUsersImport(text: string, row?: ApiRecord) {
  // Never propagate YAML diagnostics: their source snippets may contain passwords.
  try {
    if (new TextEncoder().encode(text).length > 1048576) throw new Error()
    const document = parseDocument(text, { uniqueKeys: true, strict: true, schema: 'core' })
    if (document.errors.length || document.warnings.length) throw new Error()
    const resource: unknown = document.toJS({ maxAliasCount: 0 })
    if (!record(resource) || !fields(resource, ['resource_type', 'users_groups_id', 'values', 'linked_to_cluster', 'intent']) || resource.resource_type !== 'ceph.smb.usersgroups' || !id(resource.users_groups_id) || (resource.intent !== undefined && resource.intent !== 'present')) throw new Error()
    if (row && (row.stale === true || resource.users_groups_id !== String(row.name ?? row.natural_key ?? '').trim())) throw new Error()
    if (resource.linked_to_cluster !== undefined && !id(resource.linked_to_cluster)) throw new Error()
    const values = resource.values
    if (!record(values) || !fields(values, ['users', 'groups']) || !Array.isArray(values.users) || !Array.isArray(values.groups) || values.users.length > 1000 || values.groups.length > 1000) throw new Error()
    const users = values.users.map(user => {
      if (!record(user) || !fields(user, ['name', 'password']) || typeof user.name !== 'string' || !user.name.trim() || typeof user.password !== 'string' || !user.password) throw new Error()
      return { name: user.name, password: user.password }
    })
    const groups = values.groups.map(group => {
      if (!record(group) || !fields(group, ['name']) || typeof group.name !== 'string' || !group.name || group.name.trim() !== group.name || /[\r\n]/.test(group.name)) throw new Error()
      return group.name
    })
    if (new Set(users.map(user => user.name)).size !== users.length || new Set(groups).size !== groups.length) throw new Error()
    return { name: resource.users_groups_id, users: JSON.stringify(users), groups: groups.join('\n'), clear_groups: groups.length === 0, linked_to_cluster: resource.linked_to_cluster as string | undefined }
  } catch {
    throw new Error('用户组资源文件无效或与当前编辑身份不一致')
  }
}
