export function rgwUserKeyRows(value: unknown) {
  if (!Array.isArray(value)) return undefined
  const rows: { key: number; user: string; state: string; created: string }[] = []
  for (const [key, entry] of value.entries()) {
    if (!entry || typeof entry !== 'object' || Array.isArray(entry)) return undefined
    rows.push({
      key,
      user: typeof entry.user === 'string' && entry.user !== '' ? entry.user : '所属用户未返回或无效',
      state: entry.active === true ? '已启用' : entry.active === false ? '未启用' : '启用状态未返回或无效',
      created: typeof entry.create_date === 'string' && entry.create_date.trim() !== '' ? entry.create_date : '创建时间未返回或无效'
    })
  }
  return rows
}
