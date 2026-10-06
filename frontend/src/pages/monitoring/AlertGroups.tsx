import { Alert } from 'antd'
import type { ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { RecordDetail } from '../../components/RecordDetail'
import type { ExternalListPageDefinition } from '../ExternalListPage'
import { alertField } from './alertColumns'

export function groupedAlertItems(value: unknown): ApiRecord[] | null {
  return Array.isArray(value) && value.every(item => item && typeof item === 'object' && !Array.isArray(item)) ? value : null
}

export function groupAlertFacets(value: unknown, parent: string, field: string) {
  const items = groupedAlertItems(value)
  if (!items) return { values: [], summary: '未知' }
  const counts = new Map<string, number>()
  let missing = 0
  for (const item of items) {
    const container = item[parent]
    const value = container && typeof container === 'object' && !Array.isArray(container) ? (container as ApiRecord)[field] : undefined
    if (typeof value !== 'string' || !value.trim()) { missing++; continue }
    counts.set(value, (counts.get(value) ?? 0) + 1)
  }
  const summary = Array.from(counts, ([value, count]) => `${value}: ${count}`)
  if (missing) summary.push(`未提供: ${missing}`)
  return { values: Array.from(counts.keys()), summary: summary.join('；') || '无实例' }
}

function GroupInstances({ row }: { row: ApiRecord }) {
  const items = groupedAlertItems(row.alerts)
  if (!items) return <Alert type="warning" message="分组实例数据缺失或异常" />
  return <><Alert type="info" message="这是 Alertmanager 的接收器与路由分组，不是按名称合并。筛选按组内任一实例匹配，详情始终显示该组全部实例；不同列条件可能由不同实例满足。同一告警可能出现在不同接收器组。" />
    <AppTable<ApiRecord> dataSource={items} rowKey={(item, index) => JSON.stringify([item.fingerprint, index])} pagination={{ defaultPageSize: 10 }} columns={[
      { title: '名称', render: (_, item) => alertField(item.labels, 'alertname') },
      { title: '严重程度', render: (_, item) => alertField(item.labels, 'severity') },
      { title: '状态', render: (_, item) => alertField(item.status, 'state') },
      { title: '描述', render: (_, item) => alertField(item.annotations, 'description') },
      { title: '开始时间', dataIndex: 'startsAt' }
    ]} expandable={{ expandedRowRender: item => <RecordDetail record={item} /> }} />
  </>
}

export const alertGroupDefinition: ExternalListPageDefinition = {
  title: '告警原生分组', path: '/alert/groups', requiredEndpoints: ['alertmanager'], autoRefreshMs: 5000,
  columns: [
    { key: 'receiver', title: '接收器', render: value => alertField(value, 'name') },
    { key: 'labels', title: '分组标签' },
    { key: 'states', title: '实例状态分布', render: (_, row) => groupAlertFacets(row.alerts, 'status', 'state').summary,
      localFilter: { options: ['active', 'suppressed', 'unprocessed'], value: row => groupAlertFacets(row.alerts, 'status', 'state').values } },
    { key: 'severities', title: '严重程度分布', render: (_, row) => groupAlertFacets(row.alerts, 'labels', 'severity').summary,
      localFilter: { options: ['critical', 'warning', 'info'], value: row => groupAlertFacets(row.alerts, 'labels', 'severity').values } },
    { key: 'alerts', title: '实例数', render: value => groupedAlertItems(value)?.length ?? '未知' }
  ],
  detailContent: row => <GroupInstances row={row} />
}
