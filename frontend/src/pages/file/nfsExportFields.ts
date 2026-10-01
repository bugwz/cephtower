import type { ApiRecord } from '../../api/client'

export const nfsSquashOptions = [
  { label: '映射 root 用户（root_squash）', value: 'root_squash' },
  { label: '映射 root 身份（root_id_squash）', value: 'root_id_squash' },
  { label: '映射所有用户（all_squash）', value: 'all_squash' },
  { label: '不映射 root（no_root_squash）', value: 'no_root_squash' }
]

export const nfsSecurityLabelOptions = [{ label: '启用', value: 'enabled' }, { label: '禁用', value: 'disabled' }]

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
    read_only: row?.access_type === 'RO',
    security_label: row?.security_label === true ? 'enabled' : row?.security_label === false ? 'disabled' : undefined,
    squash: nfsSquashOptions.some((option) => option.value === row?.squash) ? String(row?.squash) : undefined
  }
}
