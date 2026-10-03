import { Alert, Checkbox, Input, Select, Space } from 'antd'

export const topicCreateOptionFields = [
  { name: 'verify-ssl', label: '校验服务端证书', choices: ['true', 'false'] },
  { name: 'use-ssl', label: 'Kafka TLS', choices: ['true', 'false'] },
  { name: 'cloudevents', label: 'CloudEvents', choices: ['true', 'false'] },
  { name: 'ca-location', label: 'CA 文件路径（RGW 服务端路径，可显式为空）' },
  { name: 'amqp-version', label: 'AMQP 版本', choices: ['0-9-1', '1-0'] },
  { name: 'amqp-exchange', label: 'AMQP Exchange（可显式为空）' },
  { name: 'amqp-ack-level', label: 'AMQP 确认级别', choices: ['none', 'broker', 'routable'] },
  { name: 'http-ack-level', label: 'HTTP 确认级别（any、non-error 或 100–599；原生尚未用于发送结果判定）' },
  { name: 'kafka-ack-level', label: 'Kafka 确认级别', choices: ['none', 'broker'] },
  { name: 'mechanism', label: 'Kafka SASL 机制', choices: ['PLAIN', 'SCRAM-SHA-256', 'SCRAM-SHA-512', 'GSSAPI', 'OAUTHBEARER'] },
  { name: 'kafka-brokers', label: 'Kafka Brokers（逗号分隔 host:port，不含凭据）' }
]

export function RgwTopicOptionsEditor({ value, onChange, disabled = false }: { value?: string; onChange?: (value: string) => void; disabled?: boolean }) {
  let options: Record<string, string>
  try {
    const parsed: unknown = JSON.parse(value ?? '{}')
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed) || Object.entries(parsed).some(([key, item]) => !topicCreateOptionFields.some(field => field.name === key) || typeof item !== 'string')) throw new Error()
    options = parsed as Record<string, string>
  } catch {
    return <Alert type="error" message="协议参数草稿无效，请重新打开表单；未自动丢弃原配置" />
  }
  const change = (name: string, next: string | undefined) => {
    if (disabled) return
    const updated = { ...options }
    if (next === undefined) delete updated[name]
    else updated[name] = next
    onChange?.(JSON.stringify(updated))
  }
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message="勾选表示显式设置；取消勾选不发送该参数，不推断原生默认值。空值仅 CA 路径和 Exchange 支持。参数是否适用取决于推送协议和 RGW 编译能力；不会随协议选择静默清除配置。" />
    {topicCreateOptionFields.map(field => {
      const selected = Object.prototype.hasOwnProperty.call(options, field.name)
      return <Space direction="vertical" key={field.name} style={{ width: '100%' }}>
        <Checkbox disabled={disabled} checked={selected} onChange={event => change(field.name, event.target.checked ? '' : undefined)}>{field.label}（{field.name}）</Checkbox>
        {selected ? field.choices ? <Select aria-label={field.name} disabled={disabled} value={options[field.name] || undefined} placeholder="明确选择值" style={{ width: '100%' }} options={field.choices.map(item => ({ value: item, label: item }))} onChange={next => change(field.name, next)} /> : <Input aria-label={field.name} disabled={disabled} value={options[field.name]} onChange={event => change(field.name, event.target.value)} /> : null}
      </Space>
    })}
  </Space>
}
