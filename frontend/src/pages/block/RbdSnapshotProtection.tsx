import { Tag } from 'antd'

export function snapshotProtection(value: unknown) {
  if (value === true) return { label: '已保护', color: 'green' }
  if (value === false) return { label: '未保护', color: 'blue' }
  return { label: '保护状态未知', color: 'default' }
}

export function RbdSnapshotProtection({ value }: { value: unknown }) {
  const state = snapshotProtection(value)
  return <Tag color={state.color}>{state.label}</Tag>
}
