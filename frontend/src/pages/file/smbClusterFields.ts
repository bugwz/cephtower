import type { ApiRecord } from '../../api/client'

export function smbClusterInitialValues(row?: ApiRecord) {
  const sources = row?.user_group_settings
  return {
    user_group_ref: Array.isArray(sources) && sources.every((source) => source && typeof source === 'object' && source.source_type === 'resource' && typeof source.ref === 'string') ? sources.map((source) => source.ref).join('\n') : undefined,
    auth_mode: row?.auth_mode === 'user' || row?.auth_mode === 'active-directory' ? row.auth_mode : undefined,
    custom_dns: Array.isArray(row?.custom_dns) && row.custom_dns.every((value) => typeof value === 'string') ? row.custom_dns.join('\n') : undefined
  }
}

export function smbClusterUserGroupsBody(values: ApiRecord) {
  if (values.auth_mode !== 'user' || values.user_group_ref === undefined) return {}
  if (typeof values.user_group_ref !== 'string') throw new Error('用户组引用格式无效')
  const refs = values.user_group_ref.split(/[\s,]+/).filter(Boolean)
  if (!refs.length) throw new Error('本地用户模式至少需要一个用户组资源引用')
  return { user_group_ref: refs }
}

export function smbClusterDNSBody(values: ApiRecord) {
  if (values.custom_dns === undefined) return {}
  if (typeof values.custom_dns !== 'string') throw new Error('DNS 地址格式无效')
  return { custom_dns: values.custom_dns.split(/[\s,]+/).filter(Boolean) }
}
