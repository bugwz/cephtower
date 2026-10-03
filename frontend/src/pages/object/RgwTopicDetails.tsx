export function topicText(value: unknown) {
  return typeof value === 'string' ? value === '' ? '原生空值' : value : '未返回或不可用'
}
export function topicBoolean(value: unknown) { return value === true ? '是' : value === false ? '否' : '未返回或不可用' }
export function topicEndpoint(row: Record<string, unknown>) {
  return `${topicText(row.push_endpoint)}${row.endpoint_redacted === true ? '（凭据、查询参数或片段已隐藏；非法地址不展示）' : ''}`
}
const topicOptionLabels: Record<string, string> = {
  'verify-ssl': '校验服务端证书', 'use-ssl': 'Kafka TLS', cloudevents: 'CloudEvents',
  'ca-location': 'CA 文件路径', 'amqp-version': 'AMQP 版本', 'amqp-exchange': 'AMQP Exchange',
  'amqp-ack-level': 'AMQP 确认级别', 'http-ack-level': 'HTTP 确认级别（原生尚未用于发送结果判定）',
  'kafka-ack-level': 'Kafka 确认级别', mechanism: 'Kafka SASL 机制', 'kafka-brokers': 'Kafka Brokers'
}
export function topicOptionRows(row: Record<string, unknown>) {
  if (row.endpoint_options_status !== 'parsed' || !Array.isArray(row.endpoint_options)) return []
  return Object.entries(topicOptionLabels).map(([name, title]) => {
    const matches = (row.endpoint_options as unknown[]).filter(item => item !== null && typeof item === 'object' && (item as Record<string, unknown>).name === name)
    const item = matches.length === 1 ? matches[0] as Record<string, unknown> : undefined
    let value = '未返回或不可用'
    if (item?.status === 'returned' && typeof item.value === 'string') value = topicText(item.value)
    else if (item?.status === 'unset') value = '未显式设置（不推断当前协议默认值）'
    else if (item?.status === 'duplicate') value = '重复参数，值不展示'
    else if (item?.status === 'hidden_invalid') value = '值未通过安全展示检查，已隐藏'
    return { name, title, value }
  })
}
export function RgwTopicDetails({ row }: { row: Record<string, unknown> }) {
  const fields = [
    ['metadata_key', '元数据键'], ['push_endpoint_topic', '端点 Topic'], ['persistent_queue', '持久化队列'],
    ['time_to_live', '消息 TTL（原生值）'], ['max_retries', '最大重试次数（原生值）'], ['retry_sleep_duration', '重试等待（原生值）'],
    ['opaqueData', 'Opaque Data'], ['policy', 'Topic Policy']
  ]
  return <div>
    <p>这是采集时的通知目标配置，不代表端点可达、消息已投递或桶通知规则已配置。敏感端点参数未入库，URL 用户凭据、查询参数和片段已隐藏。</p>
    <p>持久化：{topicBoolean(row.persistent)}；原生 stored_secret 标记：{topicBoolean(row.stored_secret)}（不展示秘密内容）</p>
    <dl>{fields.map(([key, title]) => <div key={key}><dt>{title}</dt><dd><pre style={{ whiteSpace: 'pre-wrap' }}>{topicText(row[key])}</pre></dd></div>)}</dl>
    <h4>投递协议参数</h4>
    <p>仅展示已识别且通过安全检查的非凭据字段；用户名、密码及未知参数始终隐藏。参数可能不适用于当前协议，配置存在不证明其生效。{row.endpoint_options_status !== 'parsed' ? '原生参数缺失或无法完整解析。' : ''}</p>
    <dl>{topicOptionRows(row).map(item => <div key={item.name}><dt>{item.title}（{item.name}）</dt><dd><pre style={{ whiteSpace: 'pre-wrap' }}>{item.value}</pre></dd></div>)}</dl>
  </div>
}
