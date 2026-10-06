import { Alert } from 'antd'
import type { ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { RecordDetail } from '../../components/RecordDetail'
import type { ExternalListPageDefinition } from '../ExternalListPage'
import { alertField } from './alertColumns'

export function groupedAlertItems(value: unknown): ApiRecord[] | null {
  return Array.isArray(value) && value.every(item => item && typeof item === 'object' && !Array.isArray(item)) ? value : null
}

function GroupInstances({ row }: { row: ApiRecord }) {
  const items = groupedAlertItems(row.alerts)
  if (!items) return <Alert type="warning" message="分组实例数据缺失或异常" />
  return <><Alert type="info" message="这是 Alertmanager 的接收器与路由分组，不是按名称合并。每个实例保留自身状态；同一告警可能出现在不同接收器组。" />
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
    { key: 'alerts', title: '实例数', render: value => groupedAlertItems(value)?.length ?? '未知' }
  ],
  detailContent: row => <GroupInstances row={row} />
}
