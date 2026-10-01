import { Alert, Card, Space, Table, Typography } from 'antd'
import { snapshotDependencies, snapshotPendingText } from './cephfsSnapshotDependencies'

export function SnapshotCloneDependenciesPanel({ row }: { row: Record<string, unknown> }) {
  const dependencies = snapshotDependencies(row)
  return <Card title="快照克隆依赖">
    <Space direction="vertical" style={{ width: '100%' }}>
      <Typography.Text>{snapshotPendingText(dependencies.pending)}</Typography.Text>
      {dependencies.pending === 'yes' && <Alert type="warning" showIcon message="存在克隆依赖，当前不能删除此快照" description="请在子卷页面查看目标克隆状态；等待完成或取消后刷新快照信息。不会自动取消或清理任何克隆。" />}
      {dependencies.pending === 'unknown' && <Alert type="info" showIcon message="未获取到原生快照依赖状态，请刷新；未知不表示没有克隆依赖。" />}
      <Typography.Text>孤儿克隆记录：{dependencies.orphans === undefined ? '未知' : dependencies.orphans}</Typography.Text>
      {dependencies.orphans !== undefined && dependencies.orphans > 0 && <Alert type="warning" showIcon message="检测到孤儿克隆记录" description="目标索引可能不存在，但快照仍可能被元数据引用；请检查集群状态。本页面不会自动清理索引。" />}
      {dependencies.clones === undefined ? <Typography.Text type="secondary">待处理克隆目标列表不可用</Typography.Text> : <Table
        size="small" pagination={false} dataSource={dependencies.clones} rowKey={(item) => JSON.stringify([item.group, item.name])}
        columns={[{ title: '目标子卷组', dataIndex: 'group' }, { title: '目标子卷', dataIndex: 'name' }]}
        locale={{ emptyText: '无待处理克隆目标' }}
      />}
    </Space>
  </Card>
}
