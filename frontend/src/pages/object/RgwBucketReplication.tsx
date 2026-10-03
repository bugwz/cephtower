import { Table } from 'antd'

type Replication = { role: string; rules: { id: string; status: string; priority: string | null; destination_bucket: string }[] }
export function bucketReplicationData(value: unknown): Replication | undefined {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  const data = value as Replication
  if (typeof data.role !== 'string' || !Array.isArray(data.rules) || data.rules.some(rule => !rule || typeof rule.id !== 'string' || typeof rule.status !== 'string' || !rule.status || (rule.priority !== null && typeof rule.priority !== 'string') || typeof rule.destination_bucket !== 'string')) return undefined
  return data
}
export function bucketReplicationStatus(status: string) {
  return status === 'Enabled' ? '规则启用（不代表复制完成）' : status === 'Disabled' ? '规则停用' : `未知状态 ${JSON.stringify(status)}`
}
export function RgwBucketReplication({ value, configured }: { value: unknown; configured: unknown }) {
  if (configured === false && value === null) return <span>原生 S3 复制配置不存在；不代表多站点同步已停用</span>
  const data = bucketReplicationData(value)
  if (configured !== true || !data) return <span>复制配置不可用</span>
  return <div>
    <p>仅展示 S3 复制规则摘要；不代表同步策略是否生效、复制进度或数据一致性。</p>
    <p>Role（原生值）：{JSON.stringify(data.role)}。筛选条件、来源/目标 Zone 等完整设置请查看原始 XML 文档。</p>
    <Table size="small" rowKey="index" dataSource={data.rules.map((rule, index) => ({ ...rule, index }))} pagination={data.rules.length > 5 ? { pageSize: 5 } : false} scroll={{ x: 600 }} locale={{ emptyText: '原生响应无 S3 复制规则（不推断多站点同步状态）' }} columns={[
      { title: '规则 ID', dataIndex: 'id', render: (id: string) => JSON.stringify(id) },
      { title: '配置状态', dataIndex: 'status', render: bucketReplicationStatus },
      { title: '优先级（原生值）', dataIndex: 'priority', render: (value: string | null) => value === null ? '未返回' : JSON.stringify(value) },
      { title: '目标 Bucket', dataIndex: 'destination_bucket', render: (bucket: string) => bucket === '' ? '原生空值（不推断目标）' : JSON.stringify(bucket) }
    ]} />
  </div>
}
