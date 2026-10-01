import type { ApiRecord } from '../../api/client'

export function smbCephFS(row?: ApiRecord): ApiRecord {
  const fs = row?.cephfs
  return fs && typeof fs === 'object' && !Array.isArray(fs) ? fs as ApiRecord : {}
}

export function smbShareInitialValues(row?: ApiRecord) {
  const fs = smbCephFS(row)
  return { cluster: typeof row?.cluster_id === 'string' ? row.cluster_id : '', filesystem: typeof fs.volume === 'string' ? fs.volume : '', path: typeof fs.path === 'string' ? fs.path : '', readonly: typeof row?.readonly === 'boolean' ? String(row.readonly) : undefined, browseable: typeof row?.browseable === 'boolean' ? String(row.browseable) : undefined }
}

export function smbShareAccessBody(values: ApiRecord) {
  const result: ApiRecord = {}
  for (const field of ['readonly', 'browseable']) {
    const value = values[field]
    if (value === undefined || value === '') continue
    if (value !== 'true' && value !== 'false') throw new Error('请选择有效的共享访问设置')
    result[field] = value === 'true'
  }
  return result
}

export function smbBooleanText(value: unknown) {
  return value === true ? '是' : value === false ? '否' : '未知'
}
