import { Typography } from 'antd'
import { formatDateTime } from '../utils/time'

const { Text } = Typography

export function ResourceMetaBar({ observedAt, stale, staleReason }: { observedAt?: string | null; stale?: boolean; staleReason?: string | null }) {
  if (!observedAt && !staleReason && !stale) {
    return null
  }

  return (
    <div className="resource-meta-bar">
      <Text className="resource-meta-label">上次获取时间：</Text>
      <Text className="resource-meta-time">{formatDateTime(observedAt)}</Text>
      {staleReason ? <Text className="resource-meta-reason">{staleReason}</Text> : null}
    </div>
  )
}
