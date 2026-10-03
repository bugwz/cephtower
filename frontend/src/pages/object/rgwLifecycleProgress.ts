export function rgwLifecycleProgress(value: unknown): string {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return '生命周期进度不可用'
  const progress = value as Record<string, unknown>
  if (progress.found === false && progress.entry === null) return '无生命周期处理记录（不代表未配置规则）'
  if (progress.found !== true || !progress.entry || typeof progress.entry !== 'object' || Array.isArray(progress.entry)) return '生命周期进度不可用'
  const entry = progress.entry as Record<string, unknown>
  if (typeof entry.bucket !== 'string' || !entry.bucket || typeof entry.status !== 'string' || !entry.status || (entry.started !== null && typeof entry.started !== 'string')) return '生命周期进度不可用'
  const states: Record<string, string> = { UNINITIAL: '尚未初始化', PROCESSING: '处理中', FAILED: '处理失败', COMPLETE: '本轮处理完成（不保证所有对象均已过期或转换）' }
  const status = Object.prototype.hasOwnProperty.call(states, entry.status) ? states[entry.status] : `未知状态 ${entry.status}`
  return `${status}；原生状态：${entry.status}；${entry.status === 'UNINITIAL' ? '尚无有效开始时间' : `开始时间：${entry.started || '不可用'}`}`
}
