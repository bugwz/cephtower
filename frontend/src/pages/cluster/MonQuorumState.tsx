import { Tag } from 'antd'

export function MonQuorumState({ value }: { value: unknown }) {
  if (value === true) return <Tag color="success">仲裁中（采集时）</Tag>
  if (value === false) return <Tag color="default">未加入仲裁（采集时）</Tag>
  return <Tag color="warning">仲裁状态未知，请重新采集</Tag>
}
