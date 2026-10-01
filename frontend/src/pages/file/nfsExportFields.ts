import type { ApiRecord } from '../../api/client'

export const nfsSquashOptions = [
  { label: '映射 root 用户（root_squash）', value: 'root_squash' },
  { label: '映射 root 身份（root_id_squash）', value: 'root_id_squash' },
  { label: '映射所有用户（all_squash）', value: 'all_squash' },
  { label: '不映射 root（no_root_squash）', value: 'no_root_squash' }
]

export const nfsSecurityLabelOptions = [{ label: '启用', value: 'enabled' }, { label: '禁用', value: 'disabled' }]
export const nfsAccessOptions = [{ label: '读写（RW）', value: 'RW' }, { label: '只读（RO）', value: 'RO' }, { label: '默认禁止访问（NONE，客户端规则可覆盖）', value: 'NONE' }]
export const nfsProtocolOptions = [{ label: 'NFSv3', value: '3' }, { label: 'NFSv4', value: '4' }, { label: 'NFSv3 + NFSv4', value: '3,4' }]

export function nfsProtocolBody(value: unknown) {
  if (value === undefined || value === '') return {}
  if (!nfsProtocolOptions.some((option) => option.value === value)) throw new Error('请选择有效的 NFS 协议版本')
  return { protocols: String(value).split(',').map(Number) }
}

export const nfsTransportOptions = [{ label: 'TCP', value: 'TCP' }, { label: 'UDP', value: 'UDP' }, { label: 'TCP + UDP', value: 'TCP,UDP' }]

export function nfsTransportBody(value: unknown) {
  if (value === undefined || value === '') return {}
  if (!nfsTransportOptions.some((option) => option.value === value)) throw new Error('请选择有效的传输协议')
  return { transports: String(value).split(',') }
}

export function nfsSecurityTypeBody(value: unknown) {
  if (value === undefined || value === '') return {}
  if (value === 'default') return { sectype: [] }
  if (typeof value !== 'string') throw new Error('认证方式格式无效')
  const types = value.split(',').map((entry) => entry.trim())
  if (new Set(types).size !== types.length || types.some((entry) => !['none', 'sys', 'krb5', 'krb5i', 'krb5p'].includes(entry))) throw new Error('认证方式须为 none、sys、krb5、krb5i、krb5p，以逗号分隔且不重复')
  return { sectype: types }
}

export function nfsFSAL(row?: ApiRecord): ApiRecord {
  const value = row?.fsal
  return value && typeof value === 'object' && !Array.isArray(value) ? value as ApiRecord : {}
}

export function nfsExportEditReason(row: ApiRecord): string | undefined {
  if (nfsFSAL(row).name !== 'CEPH') return '当前编辑表单仅支持 CephFS 导出'
  return undefined
}

export function nfsExportInitialValues(row?: ApiRecord) {
  return {
    cluster: typeof row?.cluster_id === 'string' ? row.cluster_id : '',
    sectype: Array.isArray(row?.sectype) && row.sectype.every((value) => typeof value === 'string') ? row.sectype.join(',') : undefined,
    pseudo: typeof row?.pseudo === 'string' ? row.pseudo : '',
    path: typeof row?.path === 'string' ? row.path : '',
    filesystem: typeof nfsFSAL(row).fs_name === 'string' ? String(nfsFSAL(row).fs_name) : '',
    access_type: nfsAccessOptions.some((option) => option.value === row?.access_type) ? String(row?.access_type) : undefined,
    transports: Array.isArray(row?.transports) && row.transports.length > 0 && new Set(row.transports).size === row.transports.length && row.transports.every((value) => value === 'TCP' || value === 'UDP') ? [...row.transports].sort().join(',') : undefined,
    protocols: Array.isArray(row?.protocols) && row.protocols.length > 0 && new Set(row.protocols).size === row.protocols.length && row.protocols.every((version) => version === 3 || version === 4) ? [...row.protocols].sort().join(',') : undefined,
    security_label: row?.security_label === true ? 'enabled' : row?.security_label === false ? 'disabled' : undefined,
    squash: nfsSquashOptions.some((option) => option.value === row?.squash) ? String(row?.squash) : undefined
  }
}
