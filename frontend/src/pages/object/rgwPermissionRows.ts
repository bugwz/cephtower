export function rgwPermissionRows(value: unknown, subusers = false) {
  if (!Array.isArray(value)) return undefined
  const identity = subusers ? 'id' : 'type'
  const permission = subusers ? 'permissions' : 'perm'
  if (value.some(row => !row || typeof row !== 'object' || Array.isArray(row)
    || typeof row[identity] !== 'string' || row[identity].length === 0
    || typeof row[permission] !== 'string')) return undefined
  return value.map((row, index) => ({ key: index, identity: row[identity] as string,
    permission: row[permission] === '' ? '未指定权限' : row[permission] as string }))
}
