import { Space, Tag, Typography } from 'antd'
import { poolPGDistribution } from '../overview/pgCategory'

export function PoolPGStateTags({ value }: { value: unknown }) {
  const distribution = poolPGDistribution(value)
  if (!distribution) return <Typography.Text type="secondary">PG 状态不可用</Typography.Text>
  if (!distribution.rows.length) return <Typography.Text type="secondary">未返回 PG（0）</Typography.Text>
  return <Space wrap size={[0, 4]}>{distribution.rows.map((row) => <Tag key={row.name} color={row.category.color} title={row.category.label}>
    {row.category.label}：{row.name} × {row.count}
  </Tag>)}</Space>
}
