export interface RbdConfigurationRow {
  key: string
  name: string
  value: string
  source: string
  sourceLabel: string
}

export function rbdConfigurationRows(value: unknown): RbdConfigurationRow[] | undefined {
  if (!Array.isArray(value)) return undefined
  const sourceLabels = new Map([['config', '客户端配置'], ['pool', '池级覆盖'], ['image', '镜像级覆盖']])
  const rows: RbdConfigurationRow[] = []
  for (const [index, item] of value.entries()) {
    if (!item || typeof item !== 'object' || Array.isArray(item) || typeof item.name !== 'string' || !item.name) return undefined
    const source = typeof item.source === 'string' ? item.source : ''
    let displayValue = '未返回或值无效'
    if (typeof item.value === 'string') displayValue = item.value === '' ? '（空字符串）' : item.value
    else if (typeof item.value === 'boolean' || (typeof item.value === 'number' && Number.isSafeInteger(item.value))) displayValue = String(item.value)
    rows.push({ key: String(index), name: item.name, value: displayValue, source, sourceLabel: sourceLabels.get(source) ?? (source ? `未知来源：${source}` : '来源未返回') })
  }
  return rows
}

export function filterRbdConfiguration(rows: RbdConfigurationRow[], query: string, source: string) {
  const needle = query.trim().toLowerCase()
  return rows.filter((row) => (!source || row.source === source) && (!needle || [row.name, row.value, row.sourceLabel].some((text) => text.toLowerCase().includes(needle))))
}
