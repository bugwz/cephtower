import type { ApiRecord } from '../../api/client'

export function nfsFSAL(row?: ApiRecord): ApiRecord {
  const value = row?.fsal
  return value && typeof value === 'object' && !Array.isArray(value) ? value as ApiRecord : {}
}

export function nfsExportEditReason(row: ApiRecord): string | undefined {
  if (nfsFSAL(row).name !== 'CEPH') return '当前编辑表单仅支持 CephFS 导出'
  if (row.access_type !== 'RO' && row.access_type !== 'RW') return '当前访问类型不能使用只读开关编辑'
  return undefined
}

export function nfsExportInitialValues(row?: ApiRecord) {
  return {
    cluster: typeof row?.cluster_id === 'string' ? row.cluster_id : '',
    pseudo: typeof row?.pseudo === 'string' ? row.pseudo : '',
    path: typeof row?.path === 'string' ? row.path : '',
    filesystem: typeof nfsFSAL(row).fs_name === 'string' ? String(nfsFSAL(row).fs_name) : '',
    read_only: row?.access_type === 'RO'
  }
}
