export function rgwBucketVersioning(value: unknown) {
  switch (value) {
    case 'enabled': return '已启用（enabled）'
    case 'suspended': return '已暂停（suspended），不等同于从未启用'
    case 'off': return '未启用（off）'
    default: return typeof value === 'string' && value !== '' ? `未知状态：${value}` : '版本控制状态未返回或无效'
  }
}

export function rgwBucketBooleanState(value: unknown) {
  return value === true ? '已启用' : value === false ? '未启用' : '状态未返回或无效'
}

export function rgwBucketReshardState(value: unknown) {
  switch (value) {
    case 'None': return '当前未处于重新分片阶段（None）'
    case 'InLogrecord': return '日志记录阶段（InLogrecord）'
    case 'InProgress': return '重新分片进行中（InProgress）'
    default: return typeof value === 'string' && value !== '' ? `未知重新分片状态：${value}` : '重新分片状态未返回或无效'
  }
}
