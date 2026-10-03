import type { ApiRecord } from '../../api/client'

export function rgwUserAccountMigrationBlocked(row?: ApiRecord) {
  if (row?.account_id !== '') return '仅支持已确认尚未关联账户的用户；不能迁出或更换账户'
  if (row.type !== 'rgw') return '仅支持原生 RGW 用户'
  if (typeof row.display_name !== 'string' || !/^[A-Za-z0-9_+=,.@-]{1,64}$/.test(row.display_name)) return '请先将显示名改为 1–64 位 IAM 用户名（字母、数字或 _+=,.@-）'
  return undefined
}

export function rgwUserAccountMigrationInput(values: Record<string, unknown>, row?: ApiRecord) {
  const blocked = rgwUserAccountMigrationBlocked(row)
  if (blocked) throw new Error(blocked)
  if (typeof values.target_account_id !== 'string' || !/^RGW[0-9]{17}$/.test(values.target_account_id)) throw new Error('请输入有效的目标账户 ID')
  if (typeof row?.uid !== 'string' || !row.uid || values.migration_confirm_uid !== row.uid) throw new Error('请完整输入当前用户 UID 以确认不可逆迁移')
  return { target_account_id: values.target_account_id, migration_confirm_uid: row.uid }
}
