import type { ApiRecord } from '../../api/client'
import type { MutationFormValues } from '../ResourceListPage'

function safeInteger(value: unknown): number | undefined {
  if (typeof value !== 'number' && (typeof value !== 'string' || !/^[0-9]+$/.test(value))) return undefined
  const number = Number(value)
  return Number.isSafeInteger(number) && number >= 0 ? number : undefined
}

export function groupPermissionMode(value: unknown): string | undefined {
  const mode = safeInteger(value)
  if (mode === undefined || mode > 0o177777) return undefined
  return (mode & 0o7777).toString(8).padStart(4, '0')
}

export function groupUpdateInitialValues(row?: ApiRecord): MutationFormValues {
  const quota = safeInteger(row?.bytes_quota)
  return {
    size: quota && quota > 0 ? quota : undefined,
    unlimited: row?.bytes_quota === 'infinite',
    no_shrink: false,
    edit_attributes: false,
    pool: typeof row?.data_pool === 'string' ? row.data_pool : undefined,
    uid: safeInteger(row?.uid), gid: safeInteger(row?.gid), mode: groupPermissionMode(row?.mode)
  }
}

export function groupUpdateBody(values: MutationFormValues, clusterId: number, filesystem: string, group: string): ApiRecord {
  return {
    cluster_id: clusterId, fs: filesystem, group,
    ...(values.unlimited ? { unlimited: true } : { size: Number(values.size) }),
    no_shrink: !values.unlimited && Boolean(values.no_shrink),
    ...(values.edit_attributes ? { pool: String(values.pool ?? ''), uid: Number(values.uid), gid: Number(values.gid), mode: String(values.mode ?? '') } : {})
  }
}
