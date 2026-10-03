import { topicIdentityBlocked } from './rgwTopicPolicy'
import { topicPushEndpointFromFields } from './rgwTopicCreate'

export function topicEndpointBlocked(row: Record<string, unknown>) {
  return topicIdentityBlocked(row) ?? (typeof row.push_endpoint !== 'string' || typeof row.endpoint_redacted !== 'boolean' || typeof row.stored_secret !== 'boolean' ? '当前端点快照不完整，请刷新' : undefined)
}
export function topicEndpointInitial(row?: Record<string, unknown>) {
  if (!row) throw new Error('请选择 Topic')
  const blocked = topicEndpointBlocked(row)
  if (blocked) throw new Error(blocked)
  // Never fill a credential input from a redacted inventory URL.
  return { topic_id: row.natural_key as string, topic_arn: row.arn as string }
}
export function topicEndpointInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = topicEndpointInitial(row)
  if (values.topic_id !== initial.topic_id || values.topic_arn !== initial.topic_arn) throw new Error('不可更改 Topic 身份')
  if (values.endpoint_mode !== 'replace' && values.endpoint_mode !== 'clear' && values.endpoint_mode !== 'fields') throw new Error('请选择完整 URL、分项替换或清空端点')
  if (values.confirm_endpoint !== 'acknowledged') throw new Error('请确认完整替换、凭据与队列影响')
  const endpoint = values.endpoint_mode === 'clear' ? '' : values.endpoint_mode === 'fields' ? topicPushEndpointFromFields(values) : values.endpoint_secret
  if (typeof endpoint !== 'string' || (values.endpoint_mode === 'replace' && !endpoint)) throw new Error('请填写完整推送 URL')
  if (endpoint) {
    // Current RGW URL grammar requires a hostname and paired user:password.
    // Do not use URL.toString(), which could normalize credentials or paths.
    if (!/^(https?|amqps?|kafka):\/\/(([^:\s]+):([^@\s]+)@)?([A-Za-z0-9.:-]+)(\/[\x20-\x7E]*)?$/.test(endpoint)) throw new Error('URL 不符合原生格式；使用 http(s)、amqp(s) 或 kafka 和完整凭据对')
    try {
      const parsed = new URL(endpoint)
      if (!parsed.hostname || (!!parsed.username !== !!parsed.password)) throw new Error()
    } catch { throw new Error('推送 URL 无效') }
  }
  if (!row!.endpoint_redacted && endpoint === row!.push_endpoint) throw new Error('端点未发生变化')
  const input = { ...initial, expected_endpoint: row!.push_endpoint as string, expected_redacted: row!.endpoint_redacted as boolean, expected_stored_secret: row!.stored_secret as boolean, endpoint_secret: endpoint }
  if (new TextEncoder().encode(JSON.stringify(input)).length > 1024 * 1024 - 128) throw new Error('端点请求超过大小限制')
  return input
}
export function topicEndpointConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  topicEndpointInput(values, row)
  return `确认${values.endpoint_mode === 'clear' ? '清空' : '完整替换'} Topic ${row?.metadata_key} 的推送 URL？旧 URL 内的凭据和查询参数不会自动保留；新 URL 不在确认文字中回显。现有 EndpointArgs 保留且可能覆盖 URL 凭据，跨协议参数也不会自动清理。清空持久化端点可能删除队列并永久丢失未投递消息，恢复端点可能创建队列。需要 HTTPS RGW 管理端点；推送链路使用明文协议可能泄露通知和凭据，请自行确认目标可信及其传输安全。提交前只比较脱敏快照，无法发现隐藏凭据变化；不是原子锁，不会自动回滚或重试。不更改 Policy 或桶通知规则，回读成功不代表消息送达。`
}
