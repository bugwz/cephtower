import { Badge, Tag } from 'antd'

interface HealthBadgeProps {
  status?: string
}

export function HealthBadge({ status = 'unknown' }: HealthBadgeProps) {
  const normalized = status.toLowerCase()
  const color = normalized.includes('ok')
    ? 'success'
    : normalized.includes('warn')
      ? 'warning'
      : normalized.includes('err')
        ? 'error'
        : 'default'

  return (
    <Tag className="health-tag" color={color}>
      <Badge status={color === 'default' ? 'default' : color} />
      {status}
    </Tag>
  )
}

