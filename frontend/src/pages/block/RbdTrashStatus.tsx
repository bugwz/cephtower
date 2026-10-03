import { Space, Tag, Typography } from 'antd'

export function rbdTrashRestoreReason(row: Record<string, unknown>): string | undefined {
  if (typeof row.trash_source === 'string' && ['USER', 'USER_PARENT', 'MIRRORING'].includes(row.trash_source)) return undefined
  if (row.trash_source === 'MIGRATION') return '迁移来源不支持直接恢复，请通过迁移流程处理'
  if (row.trash_source === 'REMOVING') return '镜像正在删除，不支持恢复'
  return '回收站来源未知或不支持恢复，请重新采集并确认'
}

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
