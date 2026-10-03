export function runtimeObject(value: unknown): Record<string, unknown> | undefined {
  return value !== null && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, unknown> : undefined
}

export function runtimeValue(value: unknown): string {
  if (value === true) return '是'
  if (value === false) return '否'
  if (typeof value === 'string') return value || '未返回'
  if (typeof value === 'number' && Number.isSafeInteger(value)) return String(value)
  return '未返回或值无效'
}

export function runtimeWatchers(value: unknown) {
  if (!Array.isArray(value)) return undefined
  const rows: { key: string; address: string; client: string; cookie: string }[] = []
  for (const [index, item] of value.entries()) {
    const watcher = runtimeObject(item)
    if (!watcher) return undefined
    rows.push({ key: String(index), address: runtimeValue(watcher.address), client: runtimeValue(watcher.client), cookie: runtimeValue(watcher.cookie) })
  }
  return rows
}

export const migrationFields = [
  ['state', '迁移状态'], ['state_description', '状态说明'], ['source_spec', '外部源规格'],
  ['source_pool_name', '源池'], ['source_pool_namespace', '源命名空间'], ['source_image_name', '源镜像'], ['source_image_id', '源镜像 ID'],
  ['dest_pool_name', '目标池'], ['dest_pool_namespace', '目标命名空间'], ['dest_image_name', '目标镜像'], ['dest_image_id', '目标镜像 ID']
]

export const cacheFields = [
  ['host', '主机'], ['path', '缓存路径'], ['mode', '模式'], ['stats_timestamp', '统计时间（原值）'],
  ['present', '缓存存在'], ['empty', '缓存为空'], ['clean', '缓存干净'],
  ['size', '容量（bytes）'], ['allocated_bytes', '已分配（bytes）'], ['cached_bytes', '已缓存（bytes）'], ['dirty_bytes', '脏数据（bytes）'], ['free_bytes', '空闲（bytes）'],
  ['hits_full', '完整命中次数'], ['hits_full_percent', '完整命中（%）'], ['hits_partial', '部分命中次数'], ['hits_partial_percent', '部分命中（%）'],
  ['misses', '未命中次数'], ['hit_bytes', '命中（bytes）'], ['hit_bytes_percent', '字节命中（%）'], ['miss_bytes', '未命中（bytes）']
]

export function runtimeDetails(value: unknown, fields: string[][]) {
  const data = runtimeObject(value)
  if (!data) return undefined
  return fields.map(([key, label]) => ({ key, label, children: key.endsWith('_pool_namespace') && data[key] === '' ? '默认命名空间' : runtimeValue(data[key]) }))
}
