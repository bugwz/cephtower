import { Space, Tag, Typography } from 'antd'
import { formatDateTime } from '../../utils/time'

export function MonSnapshotState({ observedAt, stale }: { observedAt: unknown; stale: unknown }) {
  const validTime = typeof observedAt === 'string' && observedAt.trim() !== '' && Number.isFinite(Date.parse(observedAt))
  const label = stale === true ? '历史快照，请重新采集' : stale === false && validTime ? '采集快照（非实时）' : '采集状态未知'
  return <Space wrap>
    <Tag color={stale === true || !validTime || typeof stale !== 'boolean' ? 'warning' : 'default'}>{label}</Tag>
    <Typography.Text type="secondary">采集时间：{validTime ? formatDateTime(observedAt as string) : '未返回或格式无效'}</Typography.Text>
  </Space>
}
