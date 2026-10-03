import type { ApiRecord } from '../../api/client'

export const rgwSubuserPermissionOptions = [
  { label: '无权限', value: 'none' }, { label: '只读', value: 'read' },
  { label: '只写', value: 'write' }, { label: '读写', value: 'readwrite' },
  { label: '完全控制', value: 'full' }
]

export function rgwSubuserOptions(row?: ApiRecord) {
  if (typeof row?.uid !== 'string' || !row.uid || !Array.isArray(row.subusers)) return []
  const ids = row.subusers.map(item => item && typeof item === 'object' ? (item as ApiRecord).id : undefined)
  const prefix = row.uid + ':'
  return ids.flatMap(id => {
    if (typeof id !== 'string' || !id.startsWith(prefix) || ids.filter(value => value === id).length !== 1) return []
    const name = id.slice(prefix.length)
    if (!/^[A-Za-z0-9_.@-]+$/.test(name) || name.startsWith('-')) return []
    return [{ label: id, value: name }]
  })
}

export function rgwSubuserInput(values: Record<string, unknown>, row?: ApiRecord) {
  if (values.action !== 'modify' && values.action !== 'rm') throw new Error('请选择子用户操作')
  if (!rgwSubuserOptions(row).some(option => option.value === values.subuser)) throw new Error('请选择该用户已有的子用户；请先刷新库存')
  if (values.confirm_subuser !== `${row!.uid}:${values.subuser}`) throw new Error('请输入完整子用户 ID 确认')
  const common = { action: values.action, subuser: values.subuser, confirm_subuser: values.confirm_subuser }
  if (values.action === 'rm') return { ...common, action: 'rm' as const }
  if (!rgwSubuserPermissionOptions.some(option => option.value === values.subuser_permission)) throw new Error('请选择子用户权限')
  return { ...common, action: 'modify' as const, subuser_permission: values.subuser_permission }
}
