import type { ApiRecord } from '../../api/client'

export function rgwRoleMutationBlocked(row?: ApiRecord) {
  if (!row || typeof row.RoleName !== 'string' || row.RoleName.length === 0 || row.RoleName !== row.RoleName.trim() || typeof row.AccountId !== 'string' || row.AccountId !== row.AccountId.trim()) {
    return '角色名称或账户作用域未返回或格式无效，请重新采集后再操作'
  }
  return undefined
}

export function rgwRoleMutationIdentity(row?: ApiRecord) {
  const blocked = rgwRoleMutationBlocked(row)
  if (blocked || !row) throw new Error(blocked)
  return { name: row.RoleName as string, ...(row.AccountId !== '' ? { account_id: row.AccountId as string } : {}) }
}
