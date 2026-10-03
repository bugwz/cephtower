import { Alert, Descriptions, Table, Typography } from 'antd'
import { imageMirrorScheduleDetails, mirrorScheduleRows } from './rbdMirrorScheduleRows'

export function ImageMirrorSchedule({ value }: { value: unknown }) {
  const details = imageMirrorScheduleDetails(value)
  if (!details) return <Typography.Text type="secondary">未返回完整调度信息（不能据此判断调度已停用）</Typography.Text>
  return <div style={{ minWidth: 360 }}>
    <Descriptions size="small" column={1} items={[
      { key: 'origin', label: '配置来源', children: details.origin },
      { key: 'target', label: '配置范围', children: details.target },
      { key: 'next', label: '下次运行（命令原值）', children: details.nextRun }
    ]} />
    <Table size="small" pagination={false} rowKey="key" dataSource={details.intervals} columns={[
      { title: '间隔', dataIndex: 'interval' },
      { title: '起始时间（保留原时区）', dataIndex: 'startTime' }
    ]} />
    <Typography.Text type="secondary">匹配的调度配置不代表镜像正在同步；请结合同步模式、角色和状态确认。</Typography.Text>
  </div>
}

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
