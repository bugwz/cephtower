import { Alert, Card, Descriptions, Typography } from 'antd'
import type { ApiRecord } from '../../api/client'

export function alertRoutingValues(value: unknown, receivers = false): string[] | null {
  if (!Array.isArray(value)) return null
  const items = receivers ? value.map(item => item && typeof item === 'object' && !Array.isArray(item) ? item.name : undefined) : value
  if (items.some(item => typeof item !== 'string' || !item.trim())) return null
  return items
}

export function AlertRoutingDetails({ row }: { row: ApiRecord }) {
  const status = row.status && typeof row.status === 'object' && !Array.isArray(row.status) ? row.status as ApiRecord : {}
  const entries = [
    { label: '关联静默 ID', values: alertRoutingValues(status.silencedBy) },
    { label: '抑制源告警指纹', values: alertRoutingValues(status.inhibitedBy) },
    { label: '接收器', values: alertRoutingValues(row.receivers, true) }
  ]
  return <Card title="通知路由与抑制信息" style={{ marginTop: 16 }}>
    <Alert type="info" message="以下为 Alertmanager 本次查询返回的关联信息。接收器存在不代表通知已送达；关联为空也不保证立即发送通知。刷新告警列表并重新打开详情可更新快照。" />
    <Descriptions column={1} bordered style={{ marginTop: 12 }}>
      {entries.map(entry => <Descriptions.Item key={entry.label} label={entry.label}>
        {entry.values === null ? <Typography.Text type="warning">未提供或格式异常，无法确认</Typography.Text> : entry.values.length === 0 ? '本次响应没有关联项' : entry.values.map((value, index) => <div key={index}><Typography.Text copyable>{value}</Typography.Text></div>)}
      </Descriptions.Item>)}
    </Descriptions>
  </Card>
}
