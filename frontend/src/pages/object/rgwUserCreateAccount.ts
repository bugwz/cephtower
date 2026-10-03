import { loadRgwMigrationAccountOptions } from './rgwMigrationAccountOptions'

export function rgwUserCreateAccountInput(values: Record<string, unknown>) {
  if (values.account_mode === undefined || values.account_mode === 'independent') return {}
  if (values.account_mode !== 'account') throw new Error('请选择用户归属方式')
  if (typeof values.account_id !== 'string' || !/^RGW[0-9]{17}$/.test(values.account_id) || values.account_id.trim() !== values.account_id) throw new Error('请选择有效账户')
  if (values.account_root !== 'enable' && values.account_root !== 'disable') throw new Error('请选择是否创建账户根用户')
  if (typeof values.display_name !== 'string' || !/^[A-Za-z0-9_+=,.@-]{1,64}$/.test(values.display_name) || values.display_name.trim() !== values.display_name) throw new Error('账户用户显示名必须为 1–64 位 IAM 用户名')
  return { account_id: values.account_id, account_root: values.account_root === 'enable' }
}

export async function loadRgwCreateAccountOptions(clusterId: number, values?: Record<string, unknown>) {
  if (values?.account_mode !== 'account' || typeof values.uid !== 'string' || !values.uid) return []
  const delimiter = values.uid.indexOf('$')
  return loadRgwMigrationAccountOptions(clusterId, { tenant: delimiter < 0 ? '' : values.uid.slice(0, delimiter) })
}
