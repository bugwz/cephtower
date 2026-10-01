import type { ApiRecord } from '../../api/client'

export function smbClusterInitialValues(row?: ApiRecord) {
  return {
    auth_mode: row?.auth_mode === 'user' || row?.auth_mode === 'active-directory' ? row.auth_mode : undefined,
    custom_dns: Array.isArray(row?.custom_dns) && row.custom_dns.every((value) => typeof value === 'string') ? row.custom_dns.join('\n') : undefined
  }
}

export function smbClusterDNSBody(values: ApiRecord) {
  if (values.custom_dns === undefined) return {}
  if (typeof values.custom_dns !== 'string') throw new Error('DNS 地址格式无效')
  return { custom_dns: values.custom_dns.split(/[\s,]+/).filter(Boolean) }
}
