import { Alert, Space, Tag, Typography } from 'antd'

const { Text } = Typography

export function ResourceMetaBar({ observedAt, stale, staleReason }: { observedAt?: string | null; stale?: boolean; staleReason?: string | null }) {
  if (!observedAt && !stale) {
    return null
  }

  return (
    <Alert
      className="resource-meta-bar"
      type={stale ? 'warning' : 'info'}
      showIcon
      message={
        <Space size={8} wrap>
          <Text>{stale ? '资源快照可能已过期' : '资源快照已加载'}</Text>
          {observedAt ? <Tag>{formatTime(observedAt)}</Tag> : null}
          {staleReason ? <Text type="secondary">{staleReason}</Text> : null}
        </Space>
      }
    />
  )
}

function formatTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

