export interface MirrorScheduleRow {
  key: string
  scope: string
  target: string
  interval: string
  startTime: string
}

export function imageMirrorScheduleDetails(value: unknown) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  const info = value as Record<string, unknown>
  if (typeof info.name !== 'string' || !Array.isArray(info.schedule_interval) || !info.schedule_interval.length) return undefined
  const origins = new Map([['cluster', '继承集群'], ['pool', '继承池'], ['namespace', '继承命名空间'], ['', '镜像专属']])
  const origin = info.inherited_from == null ? '' : info.inherited_from
  if (typeof origin !== 'string' || !origins.has(origin)) return undefined
  if (info.schedule_time != null && typeof info.schedule_time !== 'string') return undefined
  const intervals: { key: string; interval: string; startTime: string }[] = []
  for (const [index, item] of info.schedule_interval.entries()) {
    if (!item || typeof item !== 'object' || typeof item.interval !== 'string' || !/^[1-9][0-9]*[mhd]$/.test(item.interval) || (item.start_time != null && typeof item.start_time !== 'string')) return undefined
    intervals.push({ key: String(index), interval: item.interval, startTime: item.start_time || '未指定' })
  }
  const nextRun = info.schedule_status === 'available'
    ? info.schedule_time || '本次采集未返回此镜像的待执行任务'
    : '运行状态不可用或尚未采集'
  return { origin: origins.get(origin)!, target: origin === 'cluster' ? '集群' : info.name || '未返回', nextRun, intervals }
}

export function mirrorScheduleRows(value: unknown): MirrorScheduleRow[] | undefined {
  if (!Array.isArray(value)) return undefined
  const rows: MirrorScheduleRow[] = []
  const seenScopes = new Set<string>()
  for (const [index, entry] of value.entries()) {
    if (!entry || typeof entry !== 'object' || Array.isArray(entry)) return undefined
    const { pool, namespace, image, items } = entry
    if (typeof pool !== 'string' || !pool || typeof namespace !== 'string' || typeof image !== 'string' || !image || !Array.isArray(items)) return undefined
    const identity = JSON.stringify([pool, namespace, image])
    if (seenScopes.has(identity)) return undefined
    seenScopes.add(identity)
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
