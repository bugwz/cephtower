import { topicIdentityBlocked } from './rgwTopicPolicy'
export const topicWritableOptions = [
  {value:'verify-ssl',label:'校验证书（true / false）'}, {value:'use-ssl',label:'Kafka TLS（true / false）'},
  {value:'cloudevents',label:'CloudEvents（true / false）'}, {value:'ca-location',label:'CA 文件路径'},
  {value:'amqp-exchange',label:'AMQP Exchange'}, {value:'amqp-ack-level',label:'AMQP 确认（none / broker / routable）'},
  {value:'kafka-ack-level',label:'Kafka 确认（none / broker）'}, {value:'mechanism',label:'Kafka SASL 机制'}
]
export function topicOptionBlocked(row: Record<string, unknown>) {
  return topicIdentityBlocked(row) ?? (row.endpoint_options_status !== 'parsed' || !Array.isArray(row.endpoint_options) ? '参数快照不可用，请刷新' : undefined)
}
export function topicOptionInitial(row?: Record<string, unknown>) {
  if (!row) throw new Error('请选择 Topic')
  const blocked = topicOptionBlocked(row)
  if (blocked) throw new Error(blocked)
  return {topic_id:row.natural_key as string,topic_arn:row.arn as string}
}
export function topicOptionInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial=topicOptionInitial(row)
  if (values.topic_id!==initial.topic_id || values.topic_arn!==initial.topic_arn) throw new Error('不可更改 Topic 身份')
  const option=String(values.option ?? '')
  if (!topicWritableOptions.some(item=>item.value===option)) throw new Error('请选择支持的参数')
  const matches=(row!.endpoint_options as Record<string,unknown>[]).filter(item=>item && item.name===option)
  const current=matches.length===1 ? matches[0] : undefined
  if (!current || (current.status!=='unset' && current.status!=='returned') || (current.status==='returned' && typeof current.value!=='string')) throw new Error('参数重复、隐藏或不可用，不能编辑')
  if (values.confirm_option!=='acknowledged') throw new Error('请确认协议及安全影响')
  if (values.value_mode!=='set' && values.value_mode!=='empty') throw new Error('请选择设置或显式空值')
  const value=values.value_mode==='empty' ? '' : values.value
  if (typeof value!=='string') throw new Error('请填写参数值')
  let valid=false
  if (['verify-ssl','use-ssl','cloudevents'].includes(option)) valid=/^(true|false)$/i.test(value)
  if (option==='ca-location') valid=value==='' || (value.length<=4096 && /^[A-Za-z0-9_./ -]+$/.test(value))
  if (option==='amqp-exchange') valid=/^[A-Za-z0-9_.-]{0,256}$/.test(value)
  if (option==='amqp-ack-level') valid=['none','broker','routable'].includes(value)
  if (option==='kafka-ack-level') valid=['none','broker'].includes(value)
  if (option==='mechanism') valid=['PLAIN','SCRAM-SHA-256','SCRAM-SHA-512','GSSAPI','OAUTHBEARER'].includes(value)
  if (!valid || (values.value_mode==='set' && value==='')) throw new Error('参数值无效；空值仅用于 CA 路径或 Exchange，且不等于删除参数')
  const expected_value=current.status==='unset' ? '' : current.value as string
  if (current.status==='returned' && expected_value===value) throw new Error('参数未发生变化')
  return {...initial,option,value,expected_status:current.status,expected_value}
}
export function topicOptionConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input=topicOptionInput(values,row)
  return `确认修改 Topic ${row?.metadata_key} 的 ${input.option}？关闭证书校验或 TLS 可能泄露凭据和通知；确认级别及协议设置可能改变投递可靠性。空值不等于删除参数或恢复默认。现有凭据、其他参数和推送 URL 保留；原生子串匹配存在歧义时后端拒绝提交。需要 HTTPS RGW 端点及 SNS 读写权限。快照核验不是原子锁，失败可能已生效，不会自动回滚或重试，回读不代表端点可达或参数适用于当前协议。`
}
