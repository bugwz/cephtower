const states: Record<string, { text: string; color: string }> = {
  complete: { text: '可用', color: 'success' },
  init: { text: '初始化中', color: 'processing' },
  pending: { text: '等待克隆', color: 'processing' },
  'in-progress': { text: '克隆进行中', color: 'processing' },
  failed: { text: '克隆失败', color: 'error' },
  canceled: { text: '克隆已取消', color: 'warning' },
  'snapshot-retained': { text: '已删除，仅保留快照', color: 'warning' }
}

export function subvolumeState(value: unknown) {
  return typeof value === 'string' && Object.prototype.hasOwnProperty.call(states, value) ? states[value] : { text: typeof value === 'string' && value ? `未知（${value}）` : '未知', color: 'default' }
}

export function subvolumeType(value: unknown) {
  return value === 'subvolume' ? '普通子卷' : value === 'clone' ? '克隆子卷' : typeof value === 'string' && value ? `未知（${value}）` : '未知'
}

export function subvolumeReadyReason(row: { state?: unknown }): string | undefined {
  return row.state === 'complete' ? undefined : row.state === 'snapshot-retained' ? '子卷目录已删除，仅保留快照' : '只有原生状态为 complete 的子卷可以执行此操作，请刷新状态'
}
