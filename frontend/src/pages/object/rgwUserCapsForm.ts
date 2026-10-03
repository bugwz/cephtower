import type { ApiRecord } from '../../api/client'

// RGWUserCaps::is_valid_cap_type in the local Ceph reference.
export const rgwCapabilityTypes = ['user', 'users', 'buckets', 'metadata', 'info', 'usage', 'zone', 'bilog', 'mdlog', 'datalog', 'roles', 'user-policy', 'amz-cache', 'oidc-provider', 'user-info-without-keys', 'ratelimit', 'accounts']

export function rgwCapabilityOptions(row: ApiRecord | undefined, action: unknown) {
  if (action === 'add') return rgwCapabilityTypes.map(value => ({ value, label: value }))
  if ((action !== 'rm' && action !== 'replace') || !Array.isArray(row?.caps)) return []
  const seen = new Set<string>()
  const options: { value: string; label: string }[] = []
  for (const item of row.caps) {
    if (!item || typeof item !== 'object' || typeof item.type !== 'string' || !rgwCapabilityTypes.includes(item.type) || !['<none>', 'read', 'write', '*'].includes(item.perm) || seen.has(item.type)) return []
    seen.add(item.type)
    options.push({ value: item.type, label: `${item.type}（当前：${item.perm}）` })
  }
  return options
}

export function rgwCapabilityInput(values: Record<string, unknown>, row?: ApiRecord) {
  const { action, type, permission } = values
  if (!rgwCapabilityOptions(row, action).some(option => option.value === type)) throw new Error('请选择有效的管理权限类型；替换和移除要求当前已有权限')
  if (typeof permission !== 'string' || !['read', 'write', 'read,write', '*'].includes(permission)) throw new Error('请选择有效的管理权限')
  return { action, type, permission }
}
