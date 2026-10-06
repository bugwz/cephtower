import { Alert, Card, Progress } from 'antd'
import { AppTable } from '../../components/AppTable'
import type { ApiRecord } from '../../api/client'

export function progressEventRows(value: unknown): ApiRecord[] | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  const rows: ApiRecord[] = []
  for (const [id, event] of Object.entries(value)) {
    if (!event || typeof event !== 'object' || Array.isArray(event)) return null
    rows.push({ ...event, event_id: id })
  }
  return rows
}

export function eventPercent(value: unknown): number | undefined {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 1 ? Math.round(value * 1000) / 10 : undefined
}

export function ProgressEvents({ value }: { value: unknown }) {
  const rows = progressEventRows(value)
  return <Card title="Ceph 当前进度事件">
    <Alert type="info" message="来自最近一次 ceph status 采集，不是实时任务历史；100% 不代表本项目操作已成功完成。刷新集群可更新快照。" />
    {rows === null ? <Alert type="warning" message="进度事件未提供或格式异常，无法确认是否有进行中的事件。" /> : <AppTable<ApiRecord>
      dataSource={rows} rowKey="event_id" size="small" pagination={{ defaultPageSize: 5 }} locale={{ emptyText: '本次状态快照没有进度事件' }} columns={[
        { title: '事件 ID', dataIndex: 'event_id' },
        { title: '说明', dataIndex: 'message', render: value => typeof value === 'string' ? <span style={{ whiteSpace: 'pre-wrap' }}>{value}</span> : '未提供' },
        { title: '进度', dataIndex: 'progress', render: value => { const percent = eventPercent(value); return percent === undefined ? '未知' : <Progress percent={percent} status="normal" /> } }
      ]} />}
  </Card>
}
