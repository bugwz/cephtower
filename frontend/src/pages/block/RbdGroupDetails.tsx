import { Table, Typography } from 'antd'

export function rbdGroupMembers(value: unknown) {
  if (!Array.isArray(value)) return undefined
  const rows: { key: number; pool: string; namespace: string; image: string; state: string }[] = []
  for (const [key, member] of value.entries()) {
    if (!member || typeof member !== 'object' || typeof member.pool !== 'string' || typeof member.namespace !== 'string' || typeof member.image !== 'string') return undefined
    rows.push({ key, pool: member.pool || '未解析', namespace: member.namespace || '默认命名空间', image: member.image || '未解析',
      state: member.state === 0 ? '已加入' : member.state === 1 ? '未完成' : typeof member.state === 'number' && Number.isSafeInteger(member.state) ? `未知状态（${member.state}）` : '状态未返回或无效' })
  }
  return rows
}

export function rbdGroupSnapshots(value: unknown) {
  if (!Array.isArray(value)) return undefined
  const rows: { key: number; id: string; name: string; state: string }[] = []
  for (const [key, snapshot] of value.entries()) {
    if (!snapshot || typeof snapshot !== 'object' || typeof snapshot.id !== 'string' || typeof snapshot.snapshot !== 'string') return undefined
    rows.push({ key, id: snapshot.id || '未返回', name: snapshot.snapshot || '未返回',
      state: snapshot.state === 'complete' ? '已完成' : snapshot.state === 'incomplete' ? '未完成' : typeof snapshot.state === 'string' && snapshot.state ? `未知状态：${snapshot.state}` : '状态未返回或无效' })
  }
  return rows
}

export function RbdGroupMembers({ value }: { value: unknown }) {
  const rows = rbdGroupMembers(value)
  if (!rows) return <Typography.Text type="secondary">组成员信息不可用或不完整</Typography.Text>
  if (!rows.length) return <Typography.Text type="secondary">本次采集未返回组成员</Typography.Text>
  return <Table size="small" rowKey="key" pagination={{ pageSize: 5 }} dataSource={rows} columns={[
    { title: '池', dataIndex: 'pool' }, { title: '命名空间', dataIndex: 'namespace' },
    { title: '镜像', dataIndex: 'image' }, { title: '成员状态', dataIndex: 'state' }
  ]} />
}

export function RbdGroupSnapshots({ value }: { value: unknown }) {
  const rows = rbdGroupSnapshots(value)
  if (!rows) return <Typography.Text type="secondary">组快照信息不可用或不完整</Typography.Text>
  if (!rows.length) return <Typography.Text type="secondary">本次采集未返回组快照</Typography.Text>
  return <Table size="small" rowKey="key" pagination={{ pageSize: 5 }} dataSource={rows} columns={[
    { title: '快照 ID', dataIndex: 'id' }, { title: '快照名称', dataIndex: 'name' }, { title: '完成状态', dataIndex: 'state' }
  ]} />
}
