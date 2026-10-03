import { Descriptions, Typography } from 'antd'

export function rbdParentDetails(value: unknown) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  const parent = value as Record<string, unknown>
  if (typeof parent.pool !== 'string' || !parent.pool || typeof parent.image !== 'string' || !parent.image || typeof parent.snapshot !== 'string' || !parent.snapshot) return undefined
  if (typeof parent.pool_namespace !== 'string') return undefined
  const namespace = parent.pool_namespace
  return {
    path: `${parent.pool}/${namespace ? `${namespace}/` : ''}${parent.image}@${parent.snapshot}`,
    id: typeof parent.id === 'string' && parent.id ? parent.id : '未返回',
    trash: parent.trash === true ? '位于回收站' : parent.trash === false ? '不在回收站' : '未返回'
  }
}

export function RbdParent({ value }: { value: unknown }) {
  const details = rbdParentDetails(value)
  if (!details) return <Typography.Text type="secondary">未返回完整父镜像信息</Typography.Text>
  return <Descriptions size="small" column={1} items={[
    { key: 'path', label: '父快照', children: details.path },
    { key: 'id', label: '父镜像 ID', children: details.id },
    { key: 'trash', label: '父镜像回收站状态', children: details.trash }
  ]} />
}
