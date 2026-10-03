import { topicIdentityBlocked } from './rgwTopicPolicy'

export const topicAttributeOptions = [
  { value: 'OpaqueData', label: 'Opaque Data（加入通知正文）' },
  { value: 'persistent', label: '持久化（关闭可能删除队列）' },
  { value: 'time_to_live', label: 'TTL 秒数（0 为无限）' },
  { value: 'max_retries', label: '最大重试次数（0 为无限）' },
  { value: 'retry_sleep_duration', label: '重试间隔秒数（0 为无延迟）' }
]
const rowFields: Record<string, string> = { OpaqueData: 'opaqueData', persistent: 'persistent', time_to_live: 'time_to_live', max_retries: 'max_retries', retry_sleep_duration: 'retry_sleep_duration' }
export function topicAttributeBlocked(row: Record<string, unknown>) { return topicIdentityBlocked(row) }
export function topicAttributeInitial(row?: Record<string, unknown>) {
  if (!row) throw new Error('请选择 Topic')
  const blocked = topicAttributeBlocked(row)
  if (blocked) throw new Error(blocked)
  return { topic_id: row.natural_key as string, topic_arn: row.arn as string }
}
export function topicAttributeInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = topicAttributeInitial(row)
  if (values.topic_id !== initial.topic_id || values.topic_arn !== initial.topic_arn) throw new Error('不可更改 Topic 身份')
  const attribute = String(values.attribute ?? '')
  if (!Object.prototype.hasOwnProperty.call(rowFields, attribute)) throw new Error('请选择支持的属性')
  const current = row![rowFields[attribute]]
  if (attribute === 'persistent' ? typeof current !== 'boolean' : typeof current !== 'string') throw new Error('当前属性未返回，请刷新')
  const expected_value = String(current)
  let value: string
  if (attribute === 'OpaqueData') {
    if (values.mode !== 'set' && values.mode !== 'clear') throw new Error('请选择设置或清除 Opaque Data')
    if (values.mode === 'clear') value = ''
    else if (typeof values.value === 'string' && values.value.length) value = values.value
    else throw new Error('请填写 Opaque Data 或明确选择清除')
  } else if (attribute === 'persistent') {
    if (values.persistent !== 'true' && values.persistent !== 'false') throw new Error('请选择开启或关闭持久化')
    value = values.persistent
  } else {
    if (values.mode !== 'set' && values.mode !== 'default') throw new Error('请选择设置数值或使用全局默认')
    value = values.mode === 'default' ? 'None' : String(values.value ?? '')
    if (value !== 'None' && (!/^(0|[1-9][0-9]*)$/.test(value) || Number(value) > 2147483647)) throw new Error('数值必须为 0 至 2147483647 的整数')
    if (values.mode === 'set' && value === 'None') throw new Error('请明确选择使用全局默认')
  }
  if (value === expected_value) throw new Error('属性未发生变化')
  const input = { ...initial, attribute, expected_value, value }
  if (new TextEncoder().encode(JSON.stringify(input)).length > 1024 * 1024 - 128) throw new Error('属性请求超过大小限制')
  return input
}
export function topicAttributeConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = topicAttributeInput(values, row)
  const effect = input.attribute === 'persistent'
    ? '关闭持久化可能删除原生队列并永久丢失未投递消息；开启可能创建队列，但不代表投递成功。'
    : input.attribute === 'OpaqueData' ? 'Opaque Data 将加入通知正文，请勿放入秘密。' : 'TTL 或重试次数的 0 表示无限，重试间隔的 0 表示无延迟；全局默认不是 0。修改可能改变消息过期和重试行为。'
  return `确认修改 Topic ${row?.metadata_key} 的 ${input.attribute}？${effect}需要 HTTPS RGW 端点及 SNS 读写权限。不修改 Policy、推送地址或桶通知规则；快照核验不是原子锁，失败可能部分生效，不会自动回滚或重试。`
}
