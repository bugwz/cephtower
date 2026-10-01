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

export function nfsClientsBody(value: unknown) {
  if (value === undefined || value === '') return {}
  if (typeof value !== 'string') throw new Error('客户端规则必须为 JSON 数组')
  let clients: unknown
  try { clients = JSON.parse(value) } catch { throw new Error('客户端规则 JSON 格式无效') }
  if (Array.isArray(clients)) clients = clients.map((client) => client && typeof client === 'object' && Array.isArray(client.addresses) ? { ...client, addresses: client.addresses.map((address: unknown) => typeof address === 'string' ? address.trim() : address) } : client)
  if (!Array.isArray(clients) || clients.some((client) => !client || typeof client !== 'object' || !Array.isArray(client.addresses) || !client.addresses.length || client.addresses.some((address: unknown) => typeof address !== 'string' || !/^[A-Za-z0-9_.:@*?/-]+$/.test(address)))) throw new Error('每条客户端规则需包含非空 addresses 数组及有效地址')
  return { client_rules: clients.map((client) => ({ ...client, access_type: client.access_type ?? '', squash: client.squash ?? '' })) }
}

export function nfsExportEditReason(row: ApiRecord): string | undefined {
  if (nfsFSAL(row).name !== 'CEPH' && nfsFSAL(row).name !== 'RGW') return '当前编辑表单不支持此存储后端'
  return undefined
}

export function nfsFSALBody(values: ApiRecord) {
  if (values.fsal_type === 'RGW' && values.rgw_export_type === 'bucket') {
    let bucket: unknown
    try { bucket = JSON.parse(String(values.rgw_bucket ?? '')) } catch { throw new Error('请选择 RGW 桶') }
    if (!Array.isArray(bucket) || bucket.length !== 2 || bucket.some((value) => typeof value !== 'string') || !bucket[1]) throw new Error('RGW 桶身份无效')
    return { fsal_type: 'RGW', path: bucket[1], rgw_bucket_tenant: bucket[0] }
  }
  if (values.fsal_type === 'RGW') return { fsal_type: 'RGW', path: '/', rgw_user_id: String(values.rgw_user_id ?? '') }
  return { fsal_type: 'CEPH', filesystem: String(values.filesystem ?? ''), ...(values.cmount_path ? { cmount_path: String(values.cmount_path) } : {}), ...(values.sec_label_xattr !== undefined ? { sec_label_xattr: String(values.sec_label_xattr) } : {}) }
}

export function nfsRGWUserChoices(rows: ApiRecord[]) {
  const choices = new Map<string, { label: string; value: string }>()
  for (const row of rows) {
    // The collector's uid comes from user list and retains the tenant prefix.
    if (typeof row.uid !== 'string' || !row.uid.trim()) continue
    const uid = row.uid.trim()
    choices.set(uid, { value: uid, label: typeof row.display_name === 'string' && row.display_name ? `${row.display_name} (${uid})` : uid })
  }
  return [...choices.values()]
}

export function nfsRGWBucketChoices(rows: ApiRecord[]) {
  const choices = new Map<string, { label: string; value: string }>()
  for (const row of rows) {
    if (typeof row.tenant !== 'string' || typeof row.bucket !== 'string' || !row.bucket || row.bucket.includes('/')) continue
    const value = JSON.stringify([row.tenant, row.bucket])
    choices.set(value, { label: row.tenant ? `${row.tenant}/${row.bucket}` : row.bucket, value })
  }
  return [...choices.values()]
}

export function nfsExportInitialValues(row?: ApiRecord) {
  return {
    cluster: typeof row?.cluster_id === 'string' ? row.cluster_id : '',
    fsal_type: typeof nfsFSAL(row).name === 'string' ? String(nfsFSAL(row).name) : 'CEPH',
    rgw_export_type: nfsFSAL(row).name === 'RGW' && row?.path !== '/' ? 'bucket' : 'user',
    rgw_bucket: nfsFSAL(row).name === 'RGW' && typeof row?.path === 'string' && row.path !== '/' && typeof nfsFSAL(row).user_id === 'string' ? JSON.stringify([String(nfsFSAL(row).user_id).includes('$') ? String(nfsFSAL(row).user_id).split('$')[0] : '', row.path]) : undefined,
    rgw_user_id: typeof nfsFSAL(row).user_id === 'string' ? String(nfsFSAL(row).user_id) : '',
    clients: Array.isArray(row?.clients) ? JSON.stringify(row.clients, null, 2) : undefined,
    sectype: Array.isArray(row?.sectype) && row.sectype.every((value) => typeof value === 'string') ? row.sectype.join(',') : undefined,
    pseudo: typeof row?.pseudo === 'string' ? row.pseudo : '',
    path: typeof row?.path === 'string' ? row.path : '',
    filesystem: typeof nfsFSAL(row).fs_name === 'string' ? String(nfsFSAL(row).fs_name) : '',
    sec_label_xattr: typeof nfsFSAL(row).sec_label_xattr === 'string' ? String(nfsFSAL(row).sec_label_xattr) : undefined,
    cmount_path: typeof nfsFSAL(row).cmount_path === 'string' ? String(nfsFSAL(row).cmount_path) : undefined,
    access_type: nfsAccessOptions.some((option) => option.value === row?.access_type) ? String(row?.access_type) : undefined,
    transports: Array.isArray(row?.transports) && row.transports.length > 0 && new Set(row.transports).size === row.transports.length && row.transports.every((value) => value === 'TCP' || value === 'UDP') ? [...row.transports].sort().join(',') : undefined,
    protocols: Array.isArray(row?.protocols) && row.protocols.length > 0 && new Set(row.protocols).size === row.protocols.length && row.protocols.every((version) => version === 3 || version === 4) ? [...row.protocols].sort().join(',') : undefined,
    security_label: row?.security_label === true ? 'enabled' : row?.security_label === false ? 'disabled' : undefined,
    squash: nfsSquashOptions.some((option) => option.value === row?.squash) ? String(row?.squash) : undefined
  }
}
