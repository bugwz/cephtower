import type { ApiRecord } from '../../api/client'

export function rgwUserAccountRootBlocked(row?: ApiRecord) {
  if (typeof row?.account_id !== 'string' || !/^RGW[0-9]{17}$/.test(row.account_id)) return '仅支持已关联有效账户的用户；此操作不会迁移用户'
  if (row.type !== 'root' && row.type !== 'rgw') return '用户类型未返回或不支持此操作'
  return undefined
}

export function rgwUserAccountRootInput(values: Record<string, unknown>, row?: ApiRecord) {
  const blocked = rgwUserAccountRootBlocked(row)
  if (blocked) throw new Error(blocked)
  if (values.account_root !== 'enable' && values.account_root !== 'disable') throw new Error('请选择账户根用户状态')
  return { account_root: values.account_root === 'enable', expected_account_id: row!.account_id }
}
