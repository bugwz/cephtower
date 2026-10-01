import type { ApiRecord } from '../../api/client'

export function smbCephFS(row?: ApiRecord): ApiRecord {
  const fs = row?.cephfs
  return fs && typeof fs === 'object' && !Array.isArray(fs) ? fs as ApiRecord : {}
}

export function smbShareInitialValues(row?: ApiRecord) {
  const fs = smbCephFS(row)
  return { cluster: typeof row?.cluster_id === 'string' ? row.cluster_id : '', filesystem: typeof fs.volume === 'string' ? fs.volume : '', path: typeof fs.path === 'string' ? fs.path : '' }
}

export function smbBooleanText(value: unknown) {
  return value === true ? '是' : value === false ? '否' : '未知'
}
