import { Alert, Card, Progress, Tabs } from 'antd'
import type { ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { RecordDetail } from '../../components/RecordDetail'
import { eventPercent } from './ProgressEvents'

export function progressHistoryLists(value: unknown): { events: ApiRecord[]; completed: ApiRecord[] } | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  const record = value as ApiRecord
  const valid = (items: unknown): items is ApiRecord[] => Array.isArray(items) && items.every(item => item && typeof item === 'object' && !Array.isArray(item))
  return valid(record.events) && valid(record.completed) ? { events: record.events, completed: record.completed } : null
}

export function progressTime(value: unknown) {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '未知'
  const time = new Date(value * 1000)
  return Number.isFinite(time.getTime()) ? time.toISOString() : '未知'
}

export function progressSeconds(value: unknown) {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 ? `${Number(value.toFixed(3))} 秒` : '未知'
}

export function completedDuration(row: ApiRecord) {
  const start = row.started_at
  const end = row.finished_at
  return typeof start === 'number' && Number.isFinite(start) && typeof end === 'number' && Number.isFinite(end)
    ? progressSeconds(end - start) : '未知'
}

export function nativeDuration(value: unknown) {
  return typeof value === 'string' && value.trim() ? value : '未知'
}

export function ProgressHistory({ value }: { value: unknown }) {
  const lists = progressHistoryLists(value)
  return <Card title="Ceph Progress 模块任务">
    <Alert type="info" message="来自 ceph progress json 的采集快照；剩余时间是 Ceph 在采集时的估计，不是完成保证或实时倒计时。已完成记录受模块保留上限影响，不是完整审计历史，也不是本项目操作队列。展开可查看原生 refs、失败信息等字段。" />
    {!lists ? <Alert type="warning" message="Progress 模块数据不可用或格式异常，请确认模块已启用并刷新集群。" /> : <Tabs items={(['events', 'completed'] as const).map(kind => ({ key: kind, label: kind === 'events' ? '进行中' : '已完成', children: <AppTable<ApiRecord>
      dataSource={lists[kind]} rowKey={(row, index) => JSON.stringify([row.id, index])} pagination={{ defaultPageSize: 5 }} size="small"
      expandable={{ expandedRowRender: row => <RecordDetail record={row} /> }} columns={[
        { title: 'ID', dataIndex: 'id' },
        { title: '说明', dataIndex: 'message' },
        { title: '原生失败标记', dataIndex: 'failed', render: (value: unknown) => value === true ? '失败' : value === undefined ? '未报告失败' : value === false ? '否' : '未知' },
        { title: '失败说明', dataIndex: 'failure_message' },
        { title: '开始时间（UTC）', dataIndex: 'started_at', render: progressTime },
        ...(kind === 'completed' ? [
          { title: '结束时间（UTC）', dataIndex: 'finished_at', render: progressTime },
          { title: '耗时', render: (_: unknown, row: ApiRecord) => completedDuration(row) }
        ] : [
          { title: '进度', dataIndex: 'progress', render: (value: unknown) => { const percent = eventPercent(value); return percent === undefined ? '未知' : <Progress percent={percent} status="normal" /> } },
          { title: '已持续（原生）', dataIndex: 'duration', render: nativeDuration },
          { title: '预计剩余（快照）', dataIndex: 'time_remaining', render: progressSeconds }
        ])
      ]} /> }))} />}
  </Card>
}
