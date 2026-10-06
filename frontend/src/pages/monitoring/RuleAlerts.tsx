import { Alert, Card, Descriptions } from 'antd'
import type { ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'

export function ruleAlertInstances(value: unknown): ApiRecord[] | null {
  if (!Array.isArray(value) || value.some(item => !item || typeof item !== 'object' || Array.isArray(item))) return null
  return value
}

export function ruleAlertText(value: unknown) {
  return typeof value === 'string' ? value : '未提供'
}

export function ruleEvaluation(row: ApiRecord): { type: 'error' | 'warning' | 'info'; message: string } {
  if (row.health === 'err' || (typeof row.lastError === 'string' && row.lastError.trim())) {
    return { type: 'error', message: '规则报告评估错误；即使实例列表为空，也不能据此认定监控对象正常。' }
  }
  if (row.health !== 'ok') return { type: 'warning', message: '规则评估健康状态未知或尚未确认，请检查原生状态和最近评估时间。' }
  return { type: 'info', message: '原生评估健康状态为 ok，仅表示规则评估状态，不代表监控对象健康或通知已送达。' }
}

export function evaluationSeconds(value: unknown) {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 ? `${value} 秒` : '未提供'
}

export function RuleAlerts({ row }: { row: ApiRecord }) {
  const items = ruleAlertInstances(row.alerts)
  const evaluation = ruleEvaluation(row)
  return <Card title="规则告警实例" style={{ marginTop: 16 }}>
    <Alert type="info" message="来自本次 Prometheus 规则查询快照，不代表 Alertmanager 的通知或静默状态。刷新规则列表并重新打开详情可更新快照。" />
    <Alert type={evaluation.type} message={evaluation.message} />
    <Descriptions title="规则评估快照" size="small" column={1} items={[
      { key: 'health', label: '原生健康状态', children: ruleAlertText(row.health) },
      { key: 'lastEvaluation', label: '最近评估时间（原生）', children: ruleAlertText(row.lastEvaluation) },
      { key: 'evaluationTime', label: '评估耗时', children: evaluationSeconds(row.evaluationTime) },
      { key: 'lastError', label: '最近评估错误', children: <pre style={{ whiteSpace: 'pre-wrap' }}>{ruleAlertText(row.lastError)}</pre> }
    ]} />
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
