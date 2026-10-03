export function topicText(value: unknown) {
  return typeof value === 'string' ? value === '' ? '原生空值' : value : '未返回或不可用'
}
export function topicBoolean(value: unknown) { return value === true ? '是' : value === false ? '否' : '未返回或不可用' }
export function topicEndpoint(row: Record<string, unknown>) {
  return `${topicText(row.push_endpoint)}${row.endpoint_redacted === true ? '（凭据、查询参数或片段已隐藏；非法地址不展示）' : ''}`
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
  </div>
}
