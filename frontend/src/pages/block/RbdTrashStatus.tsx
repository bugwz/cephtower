import { Space, Tag, Typography } from 'antd'

export function rbdTrashStatus(value: unknown) {
  if (typeof value !== 'string' || !value.trim()) return { label: '延期状态未知', color: 'default', deadline: '未返回' }
  for (const [prefix, label, color] of [['protected until ', '延期保护中（采集时）', 'blue'], ['expired at ', '已到期（采集时）', 'orange']]) {
    if (value.startsWith(prefix) && value.slice(prefix.length).trim()) return { label, color, deadline: value.slice(prefix.length) }
  }
  return { label: '无法识别的延期状态', color: 'default', deadline: value }
}

export function RbdTrashStatus({ value }: { value: unknown }) {
  const state = rbdTrashStatus(value)
  return <Space direction="vertical" size={0}>
    <Tag color={state.color}>{state.label}</Tag>
    <Typography.Text>截止时间 / 状态原值：{state.deadline}</Typography.Text>
    <Typography.Text type="secondary">按命令原时区显示；到期不代表已删除。</Typography.Text>
  </Space>
}
