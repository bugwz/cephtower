import { Table, Typography } from 'antd'

export function rbdSnapshotDeleteReason(row: Record<string, unknown>): string | undefined {
  if (row.is_protected === true) return '请先取消快照保护后再删除'
  if (row.is_protected !== false) return '快照保护状态未知，请重新采集'
  if (!Array.isArray(row.children)) return '子镜像依赖信息不可用，请重新采集'
  if (row.children.length > 0) return '快照仍有子镜像（可能位于回收站），请先处理依赖'
  return undefined
}

export function rbdChildRows(value: unknown) {
  if (!Array.isArray(value)) return undefined
  const rows: { key: number; pool: string; namespace: string; image: string; id: string; trash: string }[] = []
  for (const [key, child] of value.entries()) {
    if (!child || typeof child !== 'object' || typeof child.pool !== 'string' || typeof child.pool_namespace !== 'string' || typeof child.image !== 'string') return undefined
    rows.push({ key, pool: child.pool || '名称未解析', namespace: child.pool_namespace || '默认命名空间', image: child.image || '名称未解析',
      id: typeof child.id === 'string' && child.id ? child.id : '未返回',
      trash: child.trash === true ? '位于回收站' : child.trash === false ? '不在回收站' : '未返回' })
  }
  return rows
}

export function RbdChildren({ value }: { value: unknown }) {
  const rows = rbdChildRows(value)
  if (!rows) return <Typography.Text type="secondary">子镜像信息不可用或不完整</Typography.Text>
  if (!rows.length) return <Typography.Text type="secondary">本次库存未返回子镜像</Typography.Text>
  return <Table size="small" rowKey="key" pagination={{ pageSize: 5 }} dataSource={rows} columns={[
    { title: '池', dataIndex: 'pool' }, { title: '命名空间', dataIndex: 'namespace' },
    { title: '子镜像', dataIndex: 'image' }, { title: '镜像 ID', dataIndex: 'id' },
    { title: '回收站状态', dataIndex: 'trash' }
  ]} />
}
