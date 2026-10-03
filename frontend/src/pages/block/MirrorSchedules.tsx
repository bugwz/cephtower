import { Alert, Table, Typography } from 'antd'
import { mirrorScheduleRows } from './rbdMirrorScheduleRows'

export function MirrorSchedules({ value, status }: { value: unknown; status: unknown }) {
  if (status !== 'available') return <Typography.Text type="secondary">调度不可用或尚未采集</Typography.Text>
  const rows = mirrorScheduleRows(value)
  if (!rows) return <Alert type="warning" message="调度数据不完整，请重新采集" />
  if (!rows.length) return <Typography.Text type="secondary">未配置镜像快照调度</Typography.Text>
  return <div style={{ minWidth: 480 }}>
    <Typography.Paragraph type="secondary">优先级：镜像 → 命名空间 → 池 → 集群。以下为配置范围，不表示每条都对同一镜像生效。</Typography.Paragraph>
    <Table size="small" rowKey="key" pagination={false} dataSource={rows} columns={[
      { title: '范围', dataIndex: 'scope' },
      { title: '目标', dataIndex: 'target' },
      { title: '间隔', dataIndex: 'interval' },
      { title: '起始时间（保留原时区）', dataIndex: 'startTime' }
    ]} />
  </div>
}
