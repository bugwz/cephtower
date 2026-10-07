import { Alert, Space } from 'antd'
import type { ApiRecord } from '../../api/client'

export function MetricNotices({ meta, source }: { meta?: ApiRecord; source?: string }) {
  const prefix = source ? `${source} · ` : ''
  const messages = (key: string) => Array.isArray(meta?.[key]) ? (meta[key] as unknown[]).filter((value): value is string => typeof value === 'string' && value.length > 0) : []
  return <Space direction="vertical" style={{ width: '100%' }}>
    {messages('warnings').map((message, index) => <Alert key={`warning-${index}`} type="warning" showIcon message={`${prefix}Prometheus 查询警告（结果可能不完整）`} description={message} />)}
    {messages('infos').map((message, index) => <Alert key={`info-${index}`} type="info" showIcon message={`${prefix}Prometheus 查询提示`} description={message} />)}
  </Space>
}
