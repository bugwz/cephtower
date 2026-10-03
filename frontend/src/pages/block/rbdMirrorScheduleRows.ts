export interface MirrorScheduleRow {
  key: string
  scope: string
  target: string
  interval: string
  startTime: string
}

export function mirrorScheduleRows(value: unknown): MirrorScheduleRow[] | undefined {
  if (!Array.isArray(value)) return undefined
  const rows: MirrorScheduleRow[] = []
  for (const [index, entry] of value.entries()) {
    if (!entry || typeof entry !== 'object' || Array.isArray(entry)) return undefined
    const { pool, namespace, image, items } = entry
    if (typeof pool !== 'string' || !pool || typeof namespace !== 'string' || typeof image !== 'string' || !image || !Array.isArray(items)) return undefined
    let scope: string
    let target: string
    if (pool === '-' && namespace === '-' && image === '-') {
      scope = '集群'
      target = '所有池（可被更具体的调度覆盖）'
    } else if (pool !== '-' && namespace === '-' && image === '-') {
      scope = '池'
      target = pool
    } else if (pool !== '-' && namespace !== '-' && image === '-') {
      scope = '命名空间'
      target = `${pool} / ${namespace || '默认命名空间'}`
    } else if (pool !== '-' && namespace !== '-' && image !== '-') {
      scope = '镜像'
      target = `${pool} / ${namespace || '默认命名空间'} / ${image}`
    } else return undefined
    for (const [itemIndex, item] of items.entries()) {
      if (!item || typeof item !== 'object' || typeof item.interval !== 'string' || !/^[1-9][0-9]*[mhd]$/.test(item.interval) || (item.start_time != null && typeof item.start_time !== 'string')) return undefined
      rows.push({ key: `${index}:${itemIndex}`, scope, target, interval: item.interval, startTime: item.start_time || '未指定' })
    }
  }
  return rows
}
