import { Alert, Card } from 'antd'
import type { ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'

export function ruleAlertInstances(value: unknown): ApiRecord[] | null {
  if (!Array.isArray(value) || value.some(item => !item || typeof item !== 'object' || Array.isArray(item))) return null
  return value
}

export function ruleAlertText(value: unknown) {
  return typeof value === 'string' ? value : '未提供'
}

export function RuleAlerts({ row }: { row: ApiRecord }) {
  const items = ruleAlertInstances(row.alerts)
  return <Card title="规则告警实例" style={{ marginTop: 16 }}>
    <Alert type="info" message="来自本次 Prometheus 规则查询快照，不代表 Alertmanager 的通知或静默状态。刷新规则列表并重新打开详情可更新快照。" />
    {items === null ? <Alert type="warning" message="未返回有效的实例列表，不能据此认定没有告警。" /> : <AppTable<ApiRecord>
      size="small" dataSource={items} rowKey={(_, index) => String(index)} pagination={{ defaultPageSize: 10 }}
      locale={{ emptyText: '该规则快照没有告警实例' }}
      columns={[
        { title: '标签', render: (_, item) => <pre style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(item.labels ?? null, null, 2)}</pre> },
        { title: '状态', render: (_, item) => ruleAlertText(item.state) },
        { title: '激活时间', render: (_, item) => ruleAlertText(item.activeAt) },
        { title: '值（原始字符串）', render: (_, item) => ruleAlertText(item.value) },
        { title: '注释', render: (_, item) => <pre style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(item.annotations ?? null, null, 2)}</pre> }
      ]}
    />}
  </Card>
}
