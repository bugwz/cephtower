import { Table } from 'antd'

type Notification = { id: string; topic: string; events: string[]; filters: { kind: string; name: string; value: string }[] }
export function bucketNotificationData(value: unknown): Notification[] | undefined {
  if (!Array.isArray(value) || value.some(rule => !rule || typeof rule.id !== 'string' || typeof rule.topic !== 'string' || !Array.isArray(rule.events) || rule.events.some((event: unknown) => typeof event !== 'string') || !Array.isArray(rule.filters) || rule.filters.some((filter: Notification['filters'][number]) => !filter || !['S3Key', 'S3Metadata', 'S3Tags'].includes(filter.kind) || typeof filter.name !== 'string' || typeof filter.value !== 'string'))) return undefined
  return value
}

export function RgwBucketNotifications({ value, configured }: { value: unknown; configured: unknown }) {
  const rules = bucketNotificationData(value)
  if (configured !== true || !rules) return <span>通知配置不可用，不能判断是否存在规则</span>
  return <div>
    <p>原生事件通知配置。事件及过滤条件按原值展示，不推断默认值、过滤组合语义、Topic 是否存在或投递是否成功。</p>
    <Table size="small" rowKey="index" dataSource={rules.map((rule, index) => ({ ...rule, index }))} pagination={rules.length > 5 ? { pageSize: 5 } : false} scroll={{ x: 650 }} locale={{ emptyText: '原生响应无事件通知规则' }} columns={[
      { title: '通知 ID', dataIndex: 'id', render: (value: string) => JSON.stringify(value) },
      { title: '目标 Topic ARN', dataIndex: 'topic', render: (value: string) => JSON.stringify(value) },
      { title: '事件（原生值）', dataIndex: 'events', render: (events: string[]) => events.length ? events.map((event, index) => <div key={index}>{JSON.stringify(event)}</div>) : '未返回事件；不推断默认订阅' },
      { title: '过滤条件', dataIndex: 'filters', render: (filters: Notification['filters']) => filters.length ? filters.map((filter, index) => <div key={index}>{filter.kind}：{JSON.stringify(filter.name)} = {JSON.stringify(filter.value)}</div>) : '未返回过滤条件' }
    ]} />
  </div>
}
