import type { ApiRecord } from '../../api/client'

export function smbCephFS(row?: ApiRecord): ApiRecord {
  const fs = row?.cephfs
  return fs && typeof fs === 'object' && !Array.isArray(fs) ? fs as ApiRecord : {}
}

export function smbShareInitialValues(row?: ApiRecord) {
  const fs = smbCephFS(row)
  return { comment: typeof row?.comment === 'string' ? row.comment : undefined, cluster: typeof row?.cluster_id === 'string' ? row.cluster_id : '', share_name: typeof row?.name === 'string' ? row.name : undefined, filesystem: typeof fs.volume === 'string' ? fs.volume : '', path: typeof fs.path === 'string' ? fs.path : '', readonly: typeof row?.readonly === 'boolean' ? String(row.readonly) : undefined, browseable: typeof row?.browseable === 'boolean' ? String(row.browseable) : undefined }
}

export function smbShareAccessBody(values: ApiRecord) {
  const result: ApiRecord = {}
  if (values.max_connections !== undefined && values.max_connections !== null && values.max_connections !== '') {
    const value = values.max_connections
    if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < 0) throw new Error('最大连接数必须为非负整数')
    result.max_connections = value
  }
  if (values.comment !== undefined) {
    if (typeof values.comment !== 'string' || /[\x00\r\n]/.test(values.comment)) throw new Error('共享描述必须为单行文本')
    result.comment = values.comment
  }
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
