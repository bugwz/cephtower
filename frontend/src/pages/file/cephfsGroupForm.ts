import type { ApiRecord } from '../../api/client'
import type { MutationFormValues } from '../ResourceListPage'

export function cephFSSafeInteger(value: unknown): number | undefined {
  if (typeof value !== 'number' && (typeof value !== 'string' || !/^[0-9]+$/.test(value))) return undefined
  const number = Number(value)
  return Number.isSafeInteger(number) && number >= 0 ? number : undefined
}

export function groupPermissionMode(value: unknown): string | undefined {
  const mode = cephFSSafeInteger(value)
  if (mode === undefined || mode > 0o177777) return undefined
  return (mode & 0o7777).toString(8).padStart(4, '0')
}

export function groupUpdateInitialValues(row?: ApiRecord): MutationFormValues {
  const quota = cephFSQuotaDecimal(row?.bytes_quota)
  return {
    size: quota,
    unlimited: row?.bytes_quota === 'infinite',
    no_shrink: false,
    edit_attributes: false,
    pool: typeof row?.data_pool === 'string' ? row.data_pool : undefined,
    uid: cephFSSafeInteger(row?.uid), gid: cephFSSafeInteger(row?.gid), mode: groupPermissionMode(row?.mode)
  }
}

export function groupUpdateBody(values: MutationFormValues, clusterId: number, filesystem: string, group: string): ApiRecord {
  return {
    cluster_id: clusterId, fs: filesystem, group,
    ...cephFSQuotaUpdateValues(values),
    ...(values.edit_attributes ? { pool: String(values.pool ?? ''), uid: Number(values.uid), gid: Number(values.gid), mode: String(values.mode ?? '') } : {})
  }
}

export function cephFSQuotaUpdateValues(values: MutationFormValues): ApiRecord {
  if (values.unlimited) return { unlimited: true, no_shrink: false }
  const size = typeof values.size === 'string' ? cephFSQuotaDecimal(values.size) : undefined
  if (size === undefined) throw new Error('配额必须是 1 到 9223372036854775807 之间的十进制整数字节数')
  return { size, no_shrink: Boolean(values.no_shrink) }
}

export function cephFSQuotaDecimal(value: unknown): string | undefined {
  const text = typeof value === 'string' ? value : typeof value === 'number' && Number.isSafeInteger(value) ? String(value) : ''
  if (!/^[1-9][0-9]*$/.test(text)) return undefined
  return BigInt(text) <= 9223372036854775807n ? text : undefined
}
