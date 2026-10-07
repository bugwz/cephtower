import type { ApiRecord } from '../../api/client'

export function smbClusterInitialValues(row?: ApiRecord) {
  const sources = row?.user_group_settings
  const domain = row?.domain_settings as ApiRecord | undefined
  const joins = domain?.join_sources
  const count = (row?.placement as ApiRecord | undefined)?.count
  const clustering = row?.clustering
  return {
    clustering: typeof clustering === 'string' && ['default', 'always', 'never'].includes(clustering) ? clustering : undefined,
    count: typeof count === 'number' ? count : undefined,
    domain_realm: typeof domain?.realm === 'string' ? domain.realm : undefined,
    domain_join_ref: Array.isArray(joins) && joins.every((source) => source && source.source_type === 'resource' && typeof source.ref === 'string') ? joins.map((source) => source.ref) : undefined,
    user_group_ref: Array.isArray(sources) && sources.every((source) => source && typeof source === 'object' && source.source_type === 'resource' && typeof source.ref === 'string') ? sources.map((source) => source.ref) : undefined,
    auth_mode: row?.auth_mode === 'user' || row?.auth_mode === 'active-directory' ? row.auth_mode : undefined,
    custom_dns: Array.isArray(row?.custom_dns) && row.custom_dns.every((value) => typeof value === 'string') ? row.custom_dns.join('\n') : undefined
  }
}

export function smbClusterCountBody(values: ApiRecord) {
  if (values.count === undefined || values.count === null || values.count === '') return {}
  const count = values.count
  if (typeof count !== 'number' || !Number.isSafeInteger(count) || count < 1) throw new Error('实例数量必须为正整数')
  return { count }
}

export function smbClusterClusteringBody(values: ApiRecord) {
  if (values.clustering === undefined) return {}
  if (typeof values.clustering !== 'string' || !['default', 'always', 'never'].includes(values.clustering)) throw new Error('请选择有效的集群协作模式')
  return { clustering: values.clustering }
}

export function smbClusterHostsBody(values: ApiRecord) {
  if (values.smb_label !== undefined && values.smb_label !== '') {
    if (typeof values.smb_label !== 'string') throw new Error('请选择有效主机标签')
    if (Array.isArray(values.smb_hosts) && values.smb_hosts.length) throw new Error('主机和标签不能同时指定')
    return { smb_label: values.smb_label }
  }
  if (values.smb_hosts === undefined) return {}
  if (!Array.isArray(values.smb_hosts) || values.smb_hosts.some((value) => typeof value !== 'string' || !value)) throw new Error('主机列表格式无效')
  const hosts = values.smb_hosts
  return hosts.length ? { smb_hosts: hosts } : {}
}

export function smbClusterUpdateHostsBody(values: ApiRecord) {
  if (values.replace_smb_label === true) {
    if (values.replace_smb_hosts === true) throw new Error('不能同时替换部署主机和标签')
    if (typeof values.smb_label !== 'string' || !values.smb_label) throw new Error('请选择部署标签')
    return { smb_label: values.smb_label }
  }
  if (values.replace_smb_hosts !== true) return {}
  const body = smbClusterHostsBody({ smb_hosts: values.smb_hosts })
  if (!body.smb_hosts?.length) throw new Error('替换部署主机时至少选择一台主机')
  return body
}

export function smbClusterDomainBody(values: ApiRecord) {
  if (values.auth_mode !== 'active-directory' || (values.domain_realm === undefined && values.domain_join_ref === undefined)) return {}
  if (typeof values.domain_realm !== 'string' || !values.domain_realm.trim() || !Array.isArray(values.domain_join_ref) || values.domain_join_ref.some((ref) => typeof ref !== 'string' || !ref)) throw new Error('请填写域名并选择域加入凭据')
  const refs = values.domain_join_ref
  if (!refs.length) throw new Error('至少需要一个域加入凭据资源 ID')
  return { domain_realm: values.domain_realm.trim(), domain_join_ref: refs }
}

export function smbClusterUserGroupsBody(values: ApiRecord) {
  if (values.auth_mode !== 'user' || values.user_group_ref === undefined) return {}
  if (!Array.isArray(values.user_group_ref) || values.user_group_ref.some((ref) => typeof ref !== 'string' || !ref)) throw new Error('用户组引用格式无效')
  const refs = values.user_group_ref
  if (!refs.length) throw new Error('本地用户模式至少需要一个用户组资源引用')
  return { user_group_ref: refs }
}

export function smbClusterDNSBody(values: ApiRecord) {
  if (values.custom_dns === undefined) return {}
  if (typeof values.custom_dns !== 'string') throw new Error('DNS 地址格式无效')
  return { custom_dns: values.custom_dns.split(/[\s,]+/).filter(Boolean) }
}

export function smbClusterPublicAddressesBody(values: ApiRecord) {
  if (values.smb_public_addresses === undefined) return {}
  if (typeof values.smb_public_addresses !== 'string') throw new Error('客户端访问地址格式无效')
  const addresses = values.smb_public_addresses.split(/[\s,]+/).filter(Boolean)
  return addresses.length ? { smb_public_addresses: addresses } : {}
}

export function smbClusterUpdatePublicAddressesBody(values: ApiRecord) {
  if (values.replace_smb_public_addresses !== true) return {}
  return { smb_public_addresses: smbClusterPublicAddressesBody(values).smb_public_addresses ?? [] }
}
